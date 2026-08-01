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
	CurrentUser(ctx context.Context, userID string) (*CurrentUserResponse, error)
}

// TokenPair is a freshly minted session handed to the client (plaintext tokens
// plus expiries for cookie lifetimes). It mirrors jwt.TokenPair but lives here so
// the TokenIssuer port does not leak the jwt package into the handler's tests;
// see jwt.TokenPair for why the pair carries UserID.
type TokenPair struct {
	UserID           string
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

// TokenRefresher exchanges a valid refresh token for a rotated session;
// NewJWTTokenRefresher adapts *jwt.Refresher to it, mirroring the TokenIssuer
// precedent so the handler tests without real signing keys or a database.
type TokenRefresher interface {
	RefreshToken(ctx context.Context, rawRefreshToken string) (TokenPair, error)
}

// CookieConfig carries deployment-dependent cookie attributes. Secure and Domain
// come from config; HttpOnly and SameSite=Strict are fixed (docs/security.md).
type CookieConfig struct {
	Domain string
	Secure bool
}

// Handler is the HTTP transport for the auth endpoints
// (register/login/refresh/logout): it decodes requests into service calls, mints
// session cookies via TokenIssuer/TokenRefresher, and maps service errors onto
// the standard envelope via errors.HandleError.
type Handler struct {
	service       AccountService
	issuer        TokenIssuer
	refresher     TokenRefresher
	refreshTokens RefreshTokenRepository
	cookies       CookieConfig
}

// NewHandler builds an auth Handler from its collaborators.
func NewHandler(
	service AccountService,
	issuer TokenIssuer,
	refresher TokenRefresher,
	refreshTokens RefreshTokenRepository,
	cookies CookieConfig,
) *Handler {
	return &Handler{
		service:       service,
		issuer:        issuer,
		refresher:     refresher,
		refreshTokens: refreshTokens,
		cookies:       cookies,
	}
}

// RegisterPublicRoutes mounts the auth endpoints that cannot require a valid
// access token: registration (a dedicated per-IP rate limit), login (throttled by
// the service's email-keyed progressive delay, not HTTP middleware), and refresh
// (008-T149: no auth middleware — it authenticates with the refresh cookie).
//
// Full paths are registered rather than an r.Route("/auth", ...) subtree because
// the caller mounts this group and RegisterProtectedRoutes as siblings under the
// same /api/v1 tree, and chi panics when one pattern is routed twice.
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	registerLimit := middleware.RateLimit(registerRateLimitPerMinute, registerRateLimitWindow)

	r.With(registerLimit).Post("/auth/register", h.handleRegister)
	r.Post("/auth/login", h.handleLogin)
	r.Post("/auth/refresh", h.handleRefresh)
}

// RegisterProtectedRoutes mounts the auth endpoints that require a valid access
// token. The gate itself (middleware.Authenticate) is applied by the caller's
// group in cmd/api/routes.go (008-T208), not here — see RegisterPublicRoutes for
// why these are full paths.
func (h *Handler) RegisterProtectedRoutes(r chi.Router) {
	r.Post("/auth/logout", h.handleLogout)
	r.Get("/auth/me", h.handleCurrentUser)
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
// @Failure     500 {object} errors.ErrorResponse
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
// @Failure     500 {object} errors.ErrorResponse
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

// refreshResponse is the 200 body of POST /auth/refresh; the rotated tokens ride
// in cookies, never in the body (specs/008-auth-collaboration-ux/contracts/api.md).
type refreshResponse struct {
	Message string `json:"message"`
}

// handleRefresh godoc
// @Summary     Refresh the session
// @Description Exchanges the refresh_token cookie for a new access token and a rotated refresh
// @Description token, both set as HTTP-only, Secure, SameSite=Strict cookies. Public by design: it
// @Description authenticates with the refresh cookie, not an access token. A missing, expired, or
// @Description already-used refresh token returns the uniform 401 authentication_required envelope.
// @Tags        auth
// @Produce     json
// @Success     200 {object} refreshResponse
// @Failure     401 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Failure     503 {object} errors.ErrorResponse
// @Router      /auth/refresh [post]
func (h *Handler) handleRefresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenCookie)
	if err != nil || cookie.Value == "" {
		// Must reuse the Refresher's own constant, not a local literal: this
		// body has to be byte-identical to the one an invalid token produces
		// (see jwt.RefreshFailureMessage).
		domainerrors.HandleError(w, r, domainerrors.Unauthorized(authjwt.RefreshFailureMessage))
		return
	}

	pair, err := h.refresher.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	h.setSessionCookies(w, pair)

	observability.LogSecurityEvent(
		correlationID(r.Context()), observability.EventAuthTokenRefresh, pair.UserID,
		observability.SeverityInfo, clientIP(r), r.UserAgent(), nil,
	)

	respondJSON(w, http.StatusOK, refreshResponse{Message: "Token refreshed successfully"})
}

// handleLogout godoc
// @Summary     Log out
// @Description Revokes the presented refresh token and clears both session cookies. Idempotent: a
// @Description missing or already-revoked refresh token still clears cookies and returns 204.
// @Tags        auth
// @Success     204 "No Content"
// @Failure     401 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
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

// handleCurrentUser godoc
// @Summary     Get the authenticated account
// @Description Returns the account behind the access_token cookie. This is how a freshly loaded
// @Description page learns who is signed in: the session cookies are HTTP-only, so the browser
// @Description cannot read them, and no user data survives a reload client-side. The record is read
// @Description from the database, so has_subscription reflects the current state rather than the
// @Description possibly-stale claim baked into the access token. A token whose account no longer
// @Description exists returns the uniform 401 rather than a 404 — the session is dead, not the
// @Description resource. The 401's error code tells the client what to do next:
// @Description authentication_required (no token, an invalid one, or a deleted account) means log
// @Description in again, while token_expired — emitted only for a validly-signed but expired token
// @Description — means refresh and retry.
// @Tags        auth
// @Produce     json
// @Success     200 {object} CurrentUserResponse
// @Failure     401 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /auth/me [get]
func (h *Handler) handleCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		// Unreachable behind middleware.Authenticate, which is the only way this
		// route is mounted. Kept so a future routing mistake fails closed rather
		// than querying for the zero-value id.
		domainerrors.HandleError(w, r, domainerrors.Unauthorized(authjwt.RefreshFailureMessage))
		return
	}

	resp, err := h.service.CurrentUser(r.Context(), userID)
	if err != nil {
		if isNotFound(err) {
			// The token is validly signed but its subject is gone (deleted
			// account). Surfacing 404 would answer "who am I?" with "you do not
			// exist" while the client still holds cookies it believes are good;
			// the uniform 401 is what drives it to clear them and log in again.
			domainerrors.HandleError(w, r, domainerrors.Unauthorized(authjwt.RefreshFailureMessage))
			return
		}
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, resp)
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

	return toTokenPair(pair), nil
}

// jwtTokenRefresher adapts *jwt.Refresher (which works in jwt.TokenPair) to the
// handler's TokenRefresher port, mirroring jwtTokenIssuer above.
type jwtTokenRefresher struct {
	refresher *authjwt.Refresher
}

// NewJWTTokenRefresher wires a TokenRefresher backed by the jwt.Refresher.
func NewJWTTokenRefresher(refresher *authjwt.Refresher) TokenRefresher {
	return &jwtTokenRefresher{refresher: refresher}
}

// RefreshToken delegates to the jwt.Refresher, mapping its TokenPair onto the
// handler-facing one. Errors pass through unchanged: the Refresher already
// returns the uniform authentication_required domain error for any invalid,
// expired, or revoked token.
func (a *jwtTokenRefresher) RefreshToken(ctx context.Context, rawRefreshToken string) (TokenPair, error) {
	pair, err := a.refresher.RefreshToken(ctx, rawRefreshToken)
	if err != nil {
		return TokenPair{}, err
	}

	return toTokenPair(pair), nil
}

// toTokenPair maps the jwt package's TokenPair onto the handler-facing one,
// shared by both adapters so the two stay in step.
func toTokenPair(pair authjwt.TokenPair) TokenPair {
	return TokenPair{
		UserID:           pair.UserID.String(),
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}
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
