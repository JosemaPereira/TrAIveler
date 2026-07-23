package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	authjwt "github.com/JosemaPereira/TrAIveler/backend/internal/auth/jwt"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	"github.com/JosemaPereira/TrAIveler/backend/internal/observability"
)

const (
	// accessTokenCookie name MUST match middleware.Authenticate's cookie name so
	// the gate (wired in cmd/api) reads what these handlers set (docs/security.md).
	accessTokenCookie  = "access_token"
	refreshTokenCookie = "refresh_token"
	// cookiePath scopes both cookies to the app so logout can clear them by path.
	cookiePath = "/"

	// registerRateLimitPerMinute is the per-IP registration cap (spec 008 api.md).
	registerRateLimitPerMinute = 10
	registerRateLimitWindow    = time.Minute
)

// AccountService is the business-logic seam the HTTP handlers depend on
// (satisfied by *Service), kept an interface so handlers unit-test with a mock
// and no database. Named AccountService, not AuthService, to avoid stuttering as
// auth.AuthService (revive) — the concrete Service name is already taken.
type AccountService interface {
	Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req LoginRequest, ipAddress, userAgent string) (*LoginResponse, error)
}

// TokenPair is a freshly minted session handed to the client (plaintext tokens
// plus expiries for cookie lifetimes). It mirrors jwt.TokenPair but lives here so
// the TokenIssuer port does not leak the jwt package into the handler's tests.
type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}

// TokenIssuer mints a new session for an authenticated user; NewJWTTokenIssuer
// adapts *jwt.Issuer to it, so the handler tests without real signing keys.
type TokenIssuer interface {
	IssueTokens(ctx context.Context, userID string, hasSubscription bool) (TokenPair, error)
}

// CookieConfig carries deployment-dependent cookie attributes. Secure and Domain
// come from config; HttpOnly and SameSite=Strict are fixed (docs/security.md).
type CookieConfig struct {
	Domain string
	Secure bool
}

// Handler is the HTTP transport for the auth endpoints (register/login/logout):
// it decodes requests into service calls, mints session cookies via TokenIssuer,
// and maps service errors onto the standard envelope via errors.HandleError.
type Handler struct {
	service       AccountService
	issuer        TokenIssuer
	refreshTokens RefreshTokenRepository
	cookies       CookieConfig
}

// NewHandler builds an auth Handler from its collaborators.
func NewHandler(service AccountService, issuer TokenIssuer, refreshTokens RefreshTokenRepository, cookies CookieConfig) *Handler {
	return &Handler{service: service, issuer: issuer, refreshTokens: refreshTokens, cookies: cookies}
}

// RegisterRoutes mounts the auth endpoints on r (already scoped under /api/v1 by
// the caller). Registration has a dedicated per-IP rate limit; login's throttle
// is the service's email-keyed progressive delay (not HTTP middleware), and
// logout's auth requirement is enforced by the Authenticate gate wired in cmd/api
// (008-T208).
func (h *Handler) RegisterRoutes(r chi.Router) {
	registerLimit := middleware.RateLimit(registerRateLimitPerMinute, registerRateLimitWindow)

	r.Route("/auth", func(r chi.Router) {
		r.With(registerLimit).Post("/register", h.handleRegister)
		r.Post("/login", h.handleLogin)
		r.Post("/logout", h.handleLogout)
	})
}

// handleRegister godoc
// @Summary     Register a new account
// @Description Creates a new user (admin role). When a payment_method_token is supplied a
// @Description subscription is activated and has_subscription is true; otherwise the user is a
// @Description Free User (has_subscription false). On success an access/refresh token pair is set as
// @Description HTTP-only, Secure, SameSite=Strict cookies.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body RegisterRequest true "Registration payload"
// @Success     201 {object} RegisterResponse
// @Failure     400 {object} invalidRequestEnvelope
// @Failure     409 {object} errors.ErrorResponse
// @Failure     422 {object} errors.ErrorResponse
// @Router      /auth/register [post]
func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w, r, "request body must be valid JSON")
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	if err := h.issueSession(w, r, resp.User.ID, resp.User.HasSubscription); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// handleLogin godoc
// @Summary     Log in with email and password
// @Description Authenticates a user and, on success, sets an access/refresh token pair as HTTP-only,
// @Description Secure, SameSite=Strict cookies. Unknown email and wrong password return an identical
// @Description 401 (no account enumeration); repeated failures trigger a progressive-delay 429.
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       body body LoginRequest true "Login payload"
// @Success     200 {object} LoginResponse
// @Failure     400 {object} invalidRequestEnvelope
// @Failure     401 {object} errors.ErrorResponse
// @Failure     422 {object} errors.ErrorResponse
// @Failure     429 {object} errors.ErrorResponse
// @Router      /auth/login [post]
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w, r, "request body must be valid JSON")
		return
	}

	resp, err := h.service.Login(r.Context(), req, clientIP(r), r.UserAgent())
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	if err := h.issueSession(w, r, resp.User.ID, resp.User.HasSubscription); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// handleLogout godoc
// @Summary     Log out
// @Description Revokes the presented refresh token and clears both session cookies. Idempotent: a
// @Description missing or already-revoked refresh token still clears cookies and returns 204.
// @Tags        auth
// @Success     204 "No Content"
// @Failure     401 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    BearerAuth
// @Router      /auth/logout [post]
func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	if err := h.revokePresentedRefreshToken(r); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	h.clearSessionCookies(w)

	userID, _ := middleware.UserIDFromContext(r.Context())
	observability.LogSecurityEvent(
		correlationID(r.Context()), observability.EventAuthLogout, userID,
		observability.SeverityInfo, clientIP(r), r.UserAgent(), nil,
	)

	w.WriteHeader(http.StatusNoContent)
}

// revokePresentedRefreshToken revokes the token in the refresh_token cookie, if
// any. A missing cookie or unknown token is a no-op success (logout is
// idempotent); a real storage error is surfaced so logout fails loudly rather
// than leaving a live session.
func (h *Handler) revokePresentedRefreshToken(r *http.Request) error {
	cookie, err := r.Cookie(refreshTokenCookie)
	if err != nil || cookie.Value == "" {
		return nil
	}

	token, err := h.refreshTokens.GetRefreshTokenByHash(r.Context(), authjwt.HashRefreshToken(cookie.Value))
	if err != nil {
		if isNotFound(err) {
			return nil
		}
		return err
	}

	return h.refreshTokens.RevokeRefreshToken(r.Context(), token.ID)
}

// issueSession mints a token pair and sets the session cookies. A failure here is
// returned for the caller to map to 500 — the account already exists, so the
// client can retry via login.
func (h *Handler) issueSession(w http.ResponseWriter, r *http.Request, userID string, hasSubscription bool) error {
	pair, err := h.issuer.IssueTokens(r.Context(), userID, hasSubscription)
	if err != nil {
		return fmt.Errorf("issue session: %w", err)
	}
	h.setSessionCookies(w, pair)
	return nil
}

// setSessionCookies writes the access and refresh tokens as HTTP-only, Secure,
// SameSite=Strict cookies, each expiring with its token (docs/security.md).
func (h *Handler) setSessionCookies(w http.ResponseWriter, pair TokenPair) {
	http.SetCookie(w, h.sessionCookie(accessTokenCookie, pair.AccessToken, pair.AccessExpiresAt))
	http.SetCookie(w, h.sessionCookie(refreshTokenCookie, pair.RefreshToken, pair.RefreshExpiresAt))
}

// clearSessionCookies expires both session cookies (MaxAge<0), matching the
// attributes they were set with so the browser drops them.
func (h *Handler) clearSessionCookies(w http.ResponseWriter) {
	for _, name := range []string{accessTokenCookie, refreshTokenCookie} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     cookiePath,
			Domain:   h.cookies.Domain,
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   h.cookies.Secure,
			SameSite: http.SameSiteStrictMode,
		})
	}
}

// sessionCookie builds one session cookie with the fixed security attributes and
// a lifetime bounded by the token's expiry.
func (h *Handler) sessionCookie(name, value string, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     cookiePath,
		Domain:   h.cookies.Domain,
		Expires:  expiresAt,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.cookies.Secure,
		SameSite: http.SameSiteStrictMode,
	}
}

// jwtTokenIssuer adapts *jwt.Issuer (which works in uuid.UUID and jwt.TokenPair)
// to the handler's TokenIssuer port (string user id, auth.TokenPair).
type jwtTokenIssuer struct {
	issuer *authjwt.Issuer
}

// NewJWTTokenIssuer wires a TokenIssuer backed by the jwt.Issuer.
func NewJWTTokenIssuer(issuer *authjwt.Issuer) TokenIssuer {
	return &jwtTokenIssuer{issuer: issuer}
}

// IssueTokens parses the string user id and delegates to the jwt.Issuer, mapping
// its TokenPair onto the handler-facing one.
func (a *jwtTokenIssuer) IssueTokens(ctx context.Context, userID string, hasSubscription bool) (TokenPair, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return TokenPair{}, fmt.Errorf("parse user id: %w", err)
	}

	pair, err := a.issuer.Issue(ctx, uid, hasSubscription)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}, nil
}

// invalidRequestEnvelope is the invalid_request/400 body for a malformed request.
// errors.DomainError has no 400 constructor, so this handler writes it directly
// (mirrors internal/example/handler.go).
type invalidRequestEnvelope struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// clientIP is the host portion of RemoteAddr, for security logging and the login
// rate-limit context. As in the rate-limit middleware, X-Forwarded-For is not
// trusted here (client-spoofable).
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// writeInvalidRequest writes the "invalid_request"/400 envelope for a malformed
// request body — see invalidRequestEnvelope for why this bypasses HandleError.
func writeInvalidRequest(w http.ResponseWriter, r *http.Request, message string) {
	requestID, _ := middleware.RequestIDFromContext(r.Context())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(invalidRequestEnvelope{
		Error:     "invalid_request",
		Message:   message,
		RequestID: requestID,
	}); err != nil {
		slog.Default().Error("failed to write invalid_request envelope", "error", err, "request_id", requestID)
	}
}

// respondJSON writes v as a JSON response body with the given status.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Error("failed to write auth response", "error", err)
	}
}
