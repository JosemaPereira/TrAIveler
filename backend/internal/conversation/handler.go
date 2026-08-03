package conversation

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
)

// Manager is the business-logic seam the HTTP handler depends on
// (satisfied by *Service), kept an interface so the handler unit-tests with
// a mock and no database — mirrors auth.AccountService/trip.Manager.
type Manager interface {
	SendMessage(ctx context.Context, tripID, content string) (*Message, bool, error)
	GetHistory(ctx context.Context, tripID string) (*Session, []*Message, error)
}

// TripAccess resolves trip ownership before the handler reaches
// Manager, satisfying the anti-enumeration precondition Service
// itself deliberately leaves unchecked — see Service's own doc comment
// ("the future Conversation HTTP handler (001-T039) is expected to call
// that first"). A trip owned by someone else (or nonexistent) surfaces the
// same domain NotFound trip.Service.Get already produces, so a non-owner
// cannot distinguish the two cases via this endpoint either. Satisfied
// structurally by trip.Service — a small consumer-owned port, hand-faked in
// tests, deliberately not registered in backend/.mockery.yaml (mirrors
// conversation.ItineraryGenerator/trip.SubscriptionLookup).
type TripAccess interface {
	Get(ctx context.Context, userID, tripID string) (*trip.Trip, error)
}

// Handler is the HTTP transport for the Conversation endpoints (001-T039):
// it verifies trip ownership via TripAccess before reaching
// Manager, and maps service errors onto the standard envelope
// via domainerrors.HandleError.
type Handler struct {
	service Manager
	trips   TripAccess
}

// NewHandler builds a Handler from its collaborators.
func NewHandler(service Manager, trips TripAccess) *Handler {
	return &Handler{service: service, trips: trips}
}

// RegisterRoutes mounts the Conversation endpoints. Full paths are
// registered (not a nested r.Route("/trips/{id}/conversation", ...)
// subtree) so this handler can share the /trips prefix with trip.Handler's
// own routes without chi panicking on a duplicate pattern — mirrors
// auth.Handler's RegisterPublicRoutes/RegisterProtectedRoutes split. The
// caller mounts this inside the authenticated /api/v1 group
// (cmd/api/routes.go).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/trips/{id}/conversation", h.handleSendMessage)
	r.Get("/trips/{id}/conversation", h.handleGetHistory)
}

// SendMessageRequest is the POST /trips/{id}/conversation request body.
type SendMessageRequest struct {
	Message string `json:"message"`
}

// Validate reports a missing/blank message.
func (r *SendMessageRequest) Validate() *domainerrors.DomainError {
	if strings.TrimSpace(r.Message) == "" {
		return domainerrors.Validation("One or more fields failed validation", domainerrors.ValidationError{
			Field: "message",
			Error: "Message is required",
		})
	}
	return nil
}

// SendMessageResponse is the single SSE event payload for
// POST /trips/{id}/conversation, matching
// specs/001-product-vision-scope/contracts/api.md's Itinerary Generation
// section. Continue (conversation.ItineraryGenerator, backed by
// internal/itinerary) is synchronous — it returns one complete reply per
// turn, not token-by-token deltas — so this is framed as a single SSE event
// carrying the full turn, not a token relay (see 001-T039's Notes in
// docs/roadmap.md).
type SendMessageResponse struct {
	SessionID      string `json:"session_id"`
	Role           string `json:"role"`
	Message        string `json:"message"`
	ItineraryReady bool   `json:"itinerary_ready"`
}

// HistoryResponse is the GET /trips/{id}/conversation success body.
type HistoryResponse struct {
	SessionID string     `json:"session_id"`
	Status    string     `json:"status"`
	Messages  []*Message `json:"messages"`
}

// handleSendMessage godoc
// @Summary     Send a conversation turn
// @Description Appends a user message to the trip's planning conversation and returns the AI's next
// @Description turn as a single Server-Sent Event. Continue is synchronous — one complete reply per
// @Description turn, not token-by-token streaming — so exactly one "data:" frame is written.
// @Description itinerary_ready signals the conversation has produced a persisted itinerary.
// @Tags        conversation
// @Accept      json
// @Produce     text/event-stream
// @Param       id path string true "Trip ID"
// @Param       body body SendMessageRequest true "User message"
// @Success     200 {object} SendMessageResponse
// @Failure     400 {object} invalidRequestEnvelope
// @Failure     401 {object} errors.ErrorResponse
// @Failure     404 {object} errors.ErrorResponse
// @Failure     422 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Failure     503 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips/{id}/conversation [post]
func (h *Handler) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	tripID := chi.URLParam(r, "id")

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w, r, "request body must be valid JSON")
		return
	}
	if err := req.Validate(); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	// Ownership/anti-enumeration check first (Service itself performs none —
	// see TripAccess's doc comment) and BEFORE any response header is
	// written, so a failure here still goes through the normal JSON error
	// path rather than a half-open SSE stream.
	if _, err := h.trips.Get(r.Context(), userID, tripID); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	message, ready, err := h.service.SendMessage(r.Context(), tripID, req.Message)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	writeSSEEvent(w, SendMessageResponse{
		SessionID:      message.SessionID,
		Role:           message.Role,
		Message:        message.Content,
		ItineraryReady: ready,
	})
}

// handleGetHistory godoc
// @Summary     Get conversation history
// @Description Returns the trip's most recent planning conversation session and its messages in
// @Description chronological order.
// @Tags        conversation
// @Produce     json
// @Param       id path string true "Trip ID"
// @Success     200 {object} HistoryResponse
// @Failure     401 {object} errors.ErrorResponse
// @Failure     404 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips/{id}/conversation [get]
func (h *Handler) handleGetHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	tripID := chi.URLParam(r, "id")

	if _, err := h.trips.Get(r.Context(), userID, tripID); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	session, messages, err := h.service.GetHistory(r.Context(), tripID)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, HistoryResponse{
		SessionID: session.ID,
		Status:    session.Status,
		Messages:  messages,
	})
}

// writeSSEEvent writes payload as a single Server-Sent Events frame
// ("data: <json>\n\n") and flushes it immediately. Headers are only set here
// — after every possible error path has already returned via
// domainerrors.HandleError — so a failure never leaves a half-open stream.
func writeSSEEvent(w http.ResponseWriter, payload SendMessageResponse) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	body, err := json.Marshal(payload)
	if err != nil {
		slog.Default().Error("failed to marshal SSE payload", "error", err)
		return
	}

	if _, err := w.Write([]byte("data: ")); err != nil {
		slog.Default().Error("failed to write SSE frame", "error", err)
		return
	}
	if _, err := w.Write(body); err != nil {
		slog.Default().Error("failed to write SSE frame", "error", err)
		return
	}
	if _, err := w.Write([]byte("\n\n")); err != nil {
		slog.Default().Error("failed to write SSE frame", "error", err)
		return
	}

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// requireUserID resolves the authenticated caller from context, mirroring
// trip.requireUserID/auth.Handler.handleCurrentUser's defensive check.
func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		domainerrors.HandleError(w, r, domainerrors.Unauthorized("Authentication required"))
		return "", false
	}
	return userID, true
}

// invalidRequestEnvelope is the invalid_request/400 body for a malformed
// request, mirroring auth.invalidRequestEnvelope/trip.invalidRequestEnvelope.
type invalidRequestEnvelope struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// writeInvalidRequest writes the "invalid_request"/400 envelope for a
// malformed request, mirroring auth.writeInvalidRequest.
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

// respondJSON writes v as a JSON response body with the given status,
// mirroring auth.respondJSON/trip.respondJSON.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Error("failed to write conversation response", "error", err)
	}
}
