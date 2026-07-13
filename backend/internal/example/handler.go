package example

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
)

// examplesBasePath is used both for route mounting and for building the
// Location header on 201 Created responses.
const examplesBasePath = "/api/v1/examples"

// createRequest is the POST /examples request body.
type createRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// updateRequest is the PUT /examples/{id} request body (full replacement,
// per docs/api-design-standards.md §4). Email is immutable via this
// endpoint — see UpdateInput's doc comment for why.
type updateRequest struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// listResponse is the GET /examples response envelope
// (docs/api-design-standards.md §6/§9).
type listResponse struct {
	Data       []*Example         `json:"data"`
	Pagination paginationEnvelope `json:"pagination"`
}

type paginationEnvelope struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// invalidRequestEnvelope mirrors docs/api-design-standards.md §7's error
// shape for the "invalid_request"/400 case (malformed JSON, bad query
// params). internal/errors.DomainError has no constructor for this code —
// its catalog only covers 404/422/401/403/409 — so this handler writes the
// envelope directly here, the same way internal/middleware/errors.go does
// for its own 413/500 cases, rather than extending the shared package for
// a single local use.
type invalidRequestEnvelope struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// Handler exposes Example's HTTP transport layer, translating requests into
// Service calls and Service errors into the standard error envelope via
// errors.HandleError.
type Handler struct {
	service Service
}

// NewHandler builds a Handler backed by service.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts this package's endpoints on r, expected to already
// be scoped under /api/v1 by the caller (see cmd/api/routes.go).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/examples", func(r chi.Router) {
		r.Post("/", h.handleCreate)
		r.Get("/", h.handleList)
		r.Get("/{id}", h.handleGet)
		r.Put("/{id}", h.handleUpdate)
		r.Delete("/{id}", h.handleDelete)
	})
}

// handleCreate godoc
// @Summary     Create a new example
// @Description Creates an example resource for demonstration purposes
// @Tags        examples
// @Accept      json
// @Produce     json
// @Param       body body createRequest true "Example payload"
// @Success     201 {object} Example
// @Failure     400 {object} invalidRequestEnvelope
// @Failure     409 {object} errors.ErrorResponse
// @Failure     422 {object} errors.ErrorResponse
// @Security    BearerAuth
// @Router      /examples [post]
func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w, r, "request body must be valid JSON")
		return
	}

	ex, err := h.service.CreateExample(r.Context(), CreateInput(req))
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("%s/%s", examplesBasePath, ex.ID))
	respondJSON(w, http.StatusCreated, ex)
}

// handleGet godoc
// @Summary     Get an example by ID
// @Description Retrieves a single example resource by its ID
// @Tags        examples
// @Produce     json
// @Param       id path string true "Example ID"
// @Success     200 {object} Example
// @Failure     404 {object} errors.ErrorResponse
// @Security    BearerAuth
// @Router      /examples/{id} [get]
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	ex, err := h.service.GetExample(r.Context(), id)
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, ex)
}

// handleUpdate godoc
// @Summary     Update an example
// @Description Fully replaces an example resource (PUT semantics); the caller-known version is
// @Description carried on the If-Match header rather than the request body, matching the
// @Description concurrency-control convention documented for versioned entities in docs/data-model.md
// @Description (Trip/Activity's "Concurrency Control" sections).
// @Tags        examples
// @Accept      json
// @Produce     json
// @Param       id path string true "Example ID"
// @Param       If-Match header string true "Expected current version, for optimistic-locking concurrency control"
// @Param       body body updateRequest true "Example payload"
// @Success     200 {object} Example
// @Failure     400 {object} invalidRequestEnvelope
// @Failure     404 {object} errors.ErrorResponse
// @Failure     409 {object} errors.ErrorResponse
// @Failure     422 {object} errors.ErrorResponse
// @Security    BearerAuth
// @Router      /examples/{id} [put]
func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	version, err := parseIfMatch(r)
	if err != nil {
		writeInvalidRequest(w, r, err.Error())
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidRequest(w, r, "request body must be valid JSON")
		return
	}

	ex, err := h.service.UpdateExample(r.Context(), id, version, UpdateInput(req))
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, ex)
}

// handleDelete godoc
// @Summary     Delete an example
// @Description Deletes an example resource by its ID, responding 204 with no body on success
// @Description per docs/api-design-standards.md §8
// @Tags        examples
// @Param       id path string true "Example ID"
// @Success     204 "No Content"
// @Failure     404 {object} errors.ErrorResponse
// @Security    BearerAuth
// @Router      /examples/{id} [delete]
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.DeleteExample(r.Context(), id); err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleList godoc
// @Summary     List examples
// @Description Returns a paginated list of example resources, optionally filtered by status.
// @Description Validates page/per_page query params up front (400 on malformed or out-of-range
// @Description values, per docs/api-design-standards.md §9) rather than silently clamping them.
// @Tags        examples
// @Produce     json
// @Param       status query string false "Filter by status" Enums(active, inactive)
// @Param       page query int false "Page number" default(1)
// @Param       per_page query int false "Items per page" default(20)
// @Success     200 {object} listResponse
// @Failure     400 {object} invalidRequestEnvelope
// @Security    BearerAuth
// @Router      /examples [get]
func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	page, err := parsePositiveIntParam(r, "page", defaultPage)
	if err != nil {
		writeInvalidRequest(w, r, err.Error())
		return
	}

	perPage, err := parsePositiveIntParam(r, "per_page", defaultPerPage)
	if err != nil {
		writeInvalidRequest(w, r, err.Error())
		return
	}
	if perPage > maxPerPage {
		writeInvalidRequest(w, r, fmt.Sprintf("per_page must not exceed %d", maxPerPage))
		return
	}

	out, err := h.service.ListExamples(r.Context(), ListFilters{
		Status:  r.URL.Query().Get("status"),
		Page:    page,
		PerPage: perPage,
	})
	if err != nil {
		domainerrors.HandleError(w, r, err)
		return
	}

	respondJSON(w, http.StatusOK, listResponse{
		Data: out.Items,
		Pagination: paginationEnvelope{
			Page:       out.Page,
			PerPage:    out.PerPage,
			Total:      out.Total,
			TotalPages: out.TotalPages,
			HasNext:    out.Page < out.TotalPages,
			HasPrev:    out.Page > 1,
		},
	})
}

// parseIfMatch reads and validates the If-Match header carrying the
// caller's expected version for optimistic-locking updates.
func parseIfMatch(r *http.Request) (int, error) {
	raw := strings.Trim(r.Header.Get("If-Match"), `"`)
	if raw == "" {
		return 0, fmt.Errorf("If-Match header with the current version is required")
	}

	version, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("If-Match header must be an integer version")
	}

	return version, nil
}

// parsePositiveIntParam reads a positive-integer query parameter, returning
// def when the parameter is absent.
func parsePositiveIntParam(r *http.Request, name string, def int) (int, error) {
	raw := r.URL.Query().Get(name)
	if raw == "" {
		return def, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}

	return value, nil
}

// writeInvalidRequest writes the "invalid_request"/400 envelope — see
// invalidRequestEnvelope's doc comment for why this bypasses
// errors.HandleError.
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
		slog.Default().Error("failed to write example response", "error", err)
	}
}
