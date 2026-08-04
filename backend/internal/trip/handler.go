package trip

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// ifMatchHeader is the header docs/data-model.md's Trip "Concurrency
// Control" section requires on version-checked writes: "Client sends
// If-Match: <version> header on PUT/DELETE requests." Only Update
// (PUT) consults it here: Service.Delete (001-T035, already merged) takes
// no expectedVersion parameter — deletion has no optimistic-locking check
// at the service layer, so requiring the header there would promise a
// guarantee the service does not provide. Extending Delete's signature is
// out of this ticket's scope (001-T038 wires handlers onto the existing
// service, it does not change service signatures).
const ifMatchHeader = "If-Match"

// Manager is the business-logic seam the HTTP handler depends on
// (satisfied by *Service), kept an interface so the handler unit-tests with
// a mock and no database — mirrors auth.AccountService.
type Manager interface {
	Create(ctx context.Context, userID, role, title string, description *string) (*Trip, error)
	List(ctx context.Context, userID string) ([]*Trip, error)
	Get(ctx context.Context, userID, tripID string) (*Trip, error)
	Update(
		ctx context.Context,
		userID, tripID string,
		expectedVersion int,
		title string,
		description *string,
		status string,
	) (*Trip, error)
	Delete(ctx context.Context, userID, tripID string) error
	GetItinerary(ctx context.Context, userID, tripID string) (*Itinerary, error)
}

// RoleLookup resolves the authenticated caller's current role, so the
// handler can pass it to Manager.Create for its admin-only
// authorization check (Service.Create takes role as an explicit parameter
// rather than trusting a client-supplied value or a not-yet-built
// RequireRole middleware — see 001-T019, still Backlog). Mirrors the
// SubscriptionLookup precedent in service.go: a 1-method interface,
// hand-faked in tests, deliberately not registered in
// backend/.mockery.yaml. Satisfied structurally by auth.UserRepository via
// GetUserByID.
type RoleLookup interface {
	GetUserByID(ctx context.Context, id string) (*auth.User, error)
}

// Handler is the HTTP transport for the Trip endpoints (001-T038): it
// decodes requests into Manager calls, resolves the caller's role via
// RoleLookup for Create's authorization check, and maps service errors onto
// the standard envelope via domainerrors.HandleError.
type Handler struct {
	service Manager
	roles   RoleLookup
}

// NewHandler builds a Handler from its collaborators.
func NewHandler(service Manager, roles RoleLookup) *Handler {
	return &Handler{service: service, roles: roles}
}

// RegisterRoutes mounts the Trip endpoints. Full paths are registered
// (rather than an r.Route("/trips", ...) subtree) so a sibling handler
// (conversation.Handler) can mount its own /trips/{id}/conversation routes
// on the same router without chi panicking on a duplicate pattern — mirrors
// auth.Handler's RegisterPublicRoutes/RegisterProtectedRoutes split. The
// caller mounts this inside the authenticated /api/v1 group
// (cmd/api/routes.go): every trip endpoint requires authentication
// (specs/001-product-vision-scope/contracts/api.md).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/trips", h.handleCreate)
	r.Get("/trips", h.handleList)
	r.Get("/trips/{id}", h.handleGet)
	r.Put("/trips/{id}", h.handleUpdate)
	r.Delete("/trips/{id}", h.handleDelete)
	r.Get("/trips/{id}/itinerary", h.handleGetItinerary)
}

// CreateTripRequest is the POST /trips request body.
type CreateTripRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

// UpdateTripRequest is the PUT /trips/{id} request body. The optimistic-
// locking version rides in the If-Match header (docs/data-model.md), not
// this body.
type UpdateTripRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
}

// Response is the success body for endpoints returning one trip,
// mirroring auth.RegisterResponse's { "user": {...} } envelope convention
// (this codebase's real precedent, not docs/api-design-standards.md §6's
// generic flat-object guidance — see 001-T038's Notes in docs/roadmap.md).
type Response struct {
	Trip *Trip `json:"trip"`
}

// ListResponse is the GET /trips success body. No pagination envelope:
// ListTripsByUser's own doc comment already commits to none for MVP scope.
type ListResponse struct {
	Trips []*Trip `json:"trips"`
}

// handleCreate godoc
// @Summary     Create a trip
// @Description Creates a new draft trip owned by the authenticated user. Only an admin may create a
// @Description trip; any other role is rejected with 403 forbidden.
// @Tags        trips
// @Accept      json
// @Produce     json
// @Param       body body CreateTripRequest true "Trip payload"
// @Success     201 {object} Response
// @Failure     400 {object} invalidRequestEnvelope
// @Failure     401 {object} errors.ErrorResponse
// @Failure     403 {object} errors.ErrorResponse
// @Failure     422 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips [post]
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w, r, "request body must be valid JSON")
		return
	}

	role, err := h.resolveRole(r.Context(), userID)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	tr, err := h.service.Create(r.Context(), userID, role, req.Title, req.Description)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusCreated, Response{Trip: tr})
}

// handleList godoc
// @Summary     List trips
// @Description Returns every non-archived trip owned by the authenticated user.
// @Tags        trips
// @Produce     json
// @Success     200 {object} ListResponse
// @Failure     401 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips [get]
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	trips, err := h.service.List(r.Context(), userID)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, ListResponse{Trips: trips})
}

// handleGet godoc
// @Summary     Get a trip
// @Description Returns the trip identified by id, provided the authenticated user owns it. A trip
// @Description owned by someone else is reported as 404, not 403, so a non-owner cannot distinguish
// @Description "doesn't exist" from "exists but isn't yours" (anti-enumeration).
// @Tags        trips
// @Produce     json
// @Param       id path string true "Trip ID"
// @Success     200 {object} Response
// @Failure     401 {object} errors.ErrorResponse
// @Failure     404 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips/{id} [get]
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	tripID := chi.URLParam(r, "id")

	tr, err := h.service.Get(r.Context(), userID, tripID)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, Response{Trip: tr})
}

// handleUpdate godoc
// @Summary     Update a trip
// @Description Updates title/description/status for the trip identified by id, provided the
// @Description authenticated user is its creator and currently holds an active-or-grace-period
// @Description subscription. Requires an If-Match header carrying the trip's current version
// @Description (optimistic locking); a stale version returns 409 conflict.
// @Tags        trips
// @Accept      json
// @Produce     json
// @Param       id path string true "Trip ID"
// @Param       If-Match header string true "Current trip version"
// @Param       body body UpdateTripRequest true "Trip payload"
// @Success     200 {object} Response
// @Failure     400 {object} invalidRequestEnvelope
// @Failure     401 {object} errors.ErrorResponse
// @Failure     403 {object} errors.ErrorResponse
// @Failure     404 {object} errors.ErrorResponse
// @Failure     409 {object} errors.ErrorResponse
// @Failure     422 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips/{id} [put]
func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	tripID := chi.URLParam(r, "id")

	version, ok := parseIfMatch(w, r)
	if !ok {
		return
	}

	var req UpdateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w, r, "request body must be valid JSON")
		return
	}

	tr, err := h.service.Update(r.Context(), userID, tripID, version, req.Title, req.Description, req.Status)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, Response{Trip: tr})
}

// handleDelete godoc
// @Summary     Delete a trip
// @Description Deletes the trip identified by id (cascading to its days/activities/conversation
// @Description sessions), provided the authenticated user is its creator and currently holds a
// @Description strictly active subscription.
// @Tags        trips
// @Param       id path string true "Trip ID"
// @Success     204 "No Content"
// @Failure     401 {object} errors.ErrorResponse
// @Failure     403 {object} errors.ErrorResponse
// @Failure     404 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips/{id} [delete]
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	tripID := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), userID, tripID); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetItinerary godoc
// @Summary     Get a trip's itinerary
// @Description Returns the full nested itinerary (days, each with its destination and activities)
// @Description for the trip identified by id, provided the authenticated user owns it. A trip
// @Description owned by someone else is reported as 404, not 403, so a non-owner cannot distinguish
// @Description "doesn't exist" from "exists but isn't yours" (anti-enumeration), mirroring handleGet.
// @Tags        trips
// @Produce     json
// @Param       id path string true "Trip ID"
// @Success     200 {object} Itinerary
// @Failure     401 {object} errors.ErrorResponse
// @Failure     404 {object} errors.ErrorResponse
// @Failure     500 {object} errors.ErrorResponse
// @Security    CookieAuth
// @Router      /trips/{id}/itinerary [get]
func (h *Handler) handleGetItinerary(w http.ResponseWriter, r *http.Request) {
	userID, ok := requireUserID(w, r)
	if !ok {
		return
	}

	tripID := chi.URLParam(r, "id")

	itin, err := h.service.GetItinerary(r.Context(), userID, tripID)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, itin)
}

// resolveRole looks up userID's current role via RoleLookup, for Create's
// authorization check.
func (h *Handler) resolveRole(ctx context.Context, userID string) (string, error) {
	user, err := h.roles.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return user.Role, nil
}

// requireUserID resolves the authenticated caller from context. Unreachable
// behind middleware.Authenticate, which is the only way these routes are
// mounted in production — kept so a future routing mistake fails closed
// rather than querying with a zero-value id, mirroring
// auth.Handler.handleCurrentUser's defensive check.
func requireUserID(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		domainerrors.HandleError(w, r, domainerrors.Unauthorized("Authentication required"))
		return "", false
	}
	return userID, true
}

// parseIfMatch reads and parses the If-Match header (the trip's expected
// version, per docs/data-model.md's Concurrency Control section). A missing
// or non-integer header is a malformed request (400), not a domain error.
func parseIfMatch(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.Header.Get(ifMatchHeader)
	if raw == "" {
		writeInvalidRequest(w, r, "If-Match header is required")
		return 0, false
	}
	version, err := strconv.Atoi(raw)
	if err != nil {
		writeInvalidRequest(w, r, "If-Match header must be an integer version")
		return 0, false
	}
	return version, true
}

// invalidRequestEnvelope is the invalid_request/400 body for a malformed
// request, mirroring auth.invalidRequestEnvelope (duplicated rather than
// shared: it is a small package-private wire type, and each handler package
// already follows this precedent independently).
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
// mirroring auth.respondJSON.
func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Default().Error("failed to write trip response", "error", err)
	}
}
