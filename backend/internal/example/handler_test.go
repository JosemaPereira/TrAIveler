//go:build test

package example_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/example"
	examplemocks "github.com/JosemaPereira/TrAIveler/backend/internal/example/mocks"
)

// newTestRouter mounts the handler's routes on a real chi.Router so
// chi.URLParam(r, "id") resolves correctly, matching how it's actually
// wired in cmd/api/routes.go.
func newTestRouter(h *example.Handler) http.Handler {
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func decodeErrorBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

func TestUnitHandleCreate_ValidRequest_Returns201WithLocationHeader(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	created := &example.Example{ID: "abc-123", Name: "Ada Lovelace", Email: "ada@example.com", Status: example.StatusActive}
	svc.EXPECT().
		CreateExample(mock.Anything, example.CreateInput{Name: "Ada Lovelace", Email: "ada@example.com"}).
		Return(created, nil).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	body := bytes.NewBufferString(`{"name":"Ada Lovelace","email":"ada@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/examples", body)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "/api/v1/examples/abc-123", rec.Header().Get("Location"))

	var got example.Example
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "abc-123", got.ID)
}

func TestUnitHandleCreate_MalformedJSON_Returns400(t *testing.T) {
	svc := examplemocks.NewMockService(t) // no .EXPECT(): service must not be touched

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/examples", bytes.NewBufferString(`{not valid json`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "invalid_request", body["error"])
}

func TestUnitHandleCreate_DuplicateEmail_Returns409(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().
		CreateExample(mock.Anything, mock.Anything).
		Return(nil, domainerrors.Conflict("an example with this email already exists")).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/examples", bytes.NewBufferString(`{"name":"Ada","email":"ada@example.com"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "conflict", body["error"])
}

func TestUnitHandleCreate_ValidationFailure_Returns422(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().
		CreateExample(mock.Anything, mock.Anything).
		Return(nil, domainerrors.Validation("One or more fields failed validation",
			domainerrors.ValidationError{Field: "name", Error: "Name is required"})).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/examples", bytes.NewBufferString(`{"name":"","email":"ada@example.com"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "validation_failed", body["error"])
}

func TestUnitHandleGet_ExistingID_Returns200(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	want := &example.Example{ID: "abc-123", Name: "Ada Lovelace"}
	svc.EXPECT().GetExample(mock.Anything, "abc-123").Return(want, nil).Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/examples/abc-123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got example.Example
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "abc-123", got.ID)
}

func TestUnitHandleGet_MissingID_Returns404(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().
		GetExample(mock.Anything, "missing").
		Return(nil, domainerrors.NotFound("example", "missing")).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/examples/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "not_found", body["error"])
}

func TestUnitHandleUpdate_ValidRequestWithIfMatch_Returns200(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	updated := &example.Example{ID: "abc-123", Name: "New Name", Status: example.StatusActive, Version: 4}
	svc.EXPECT().
		UpdateExample(mock.Anything, "abc-123", 3, example.UpdateInput{Name: "New Name", Status: example.StatusActive, Count: 5}).
		Return(updated, nil).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPut, "/examples/abc-123", bytes.NewBufferString(`{"name":"New Name","status":"active","count":5}`))
	req.Header.Set("If-Match", "3")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var got example.Example
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	assert.Equal(t, "New Name", got.Name)
}

func TestUnitHandleUpdate_MissingIfMatchHeader_Returns400(t *testing.T) {
	svc := examplemocks.NewMockService(t) // no .EXPECT(): service must not be touched

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPut, "/examples/abc-123", bytes.NewBufferString(`{"name":"New Name","status":"active"}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "invalid_request", body["error"])
}

func TestUnitHandleUpdate_VersionConflict_Returns409(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().
		UpdateExample(mock.Anything, "abc-123", 3, mock.Anything).
		Return(nil, domainerrors.Conflict("example was modified by another request")).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodPut, "/examples/abc-123", bytes.NewBufferString(`{"name":"New Name","status":"active"}`))
	req.Header.Set("If-Match", "3")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "conflict", body["error"])
}

func TestUnitHandleDelete_ExistingID_Returns204NoBody(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().DeleteExample(mock.Anything, "abc-123").Return(nil).Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/examples/abc-123", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

func TestUnitHandleDelete_MissingID_Returns404(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().
		DeleteExample(mock.Anything, "missing").
		Return(domainerrors.NotFound("example", "missing")).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/examples/missing", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUnitHandleList_DefaultPagination_Returns200WithEnvelope(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().
		ListExamples(mock.Anything, example.ListFilters{Page: 1, PerPage: 20}).
		Return(&example.ListOutput{
			Items:      []*example.Example{{ID: "1"}, {ID: "2"}},
			Page:       1,
			PerPage:    20,
			Total:      2,
			TotalPages: 1,
		}, nil).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/examples", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Data       []example.Example `json:"data"`
		Pagination struct {
			Page       int  `json:"page"`
			PerPage    int  `json:"per_page"`
			Total      int  `json:"total"`
			TotalPages int  `json:"total_pages"`
			HasNext    bool `json:"has_next"`
			HasPrev    bool `json:"has_prev"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Len(t, body.Data, 2)
	assert.Equal(t, 1, body.Pagination.Page)
	assert.Equal(t, 20, body.Pagination.PerPage)
	assert.Equal(t, 2, body.Pagination.Total)
	assert.False(t, body.Pagination.HasNext)
	assert.False(t, body.Pagination.HasPrev)
}

func TestUnitHandleList_PerPageAboveMax_Returns400(t *testing.T) {
	svc := examplemocks.NewMockService(t) // no .EXPECT(): service must not be touched

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/examples?per_page=500", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	body := decodeErrorBody(t, rec)
	assert.Equal(t, "invalid_request", body["error"])
}

func TestUnitHandleList_StatusFilter_PassedThroughToService(t *testing.T) {
	svc := examplemocks.NewMockService(t)
	svc.EXPECT().
		ListExamples(mock.Anything, example.ListFilters{Status: example.StatusInactive, Page: 1, PerPage: 20}).
		Return(&example.ListOutput{Items: []*example.Example{}, Page: 1, PerPage: 20}, nil).
		Once()

	h := example.NewHandler(svc)
	router := newTestRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/examples?status=inactive", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}
