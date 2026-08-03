//go:build test

package trip_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/auth"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
	tripmocks "github.com/JosemaPereira/TrAIveler/backend/internal/trip/mocks"
)

const (
	handlerUserID = "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab"
	handlerTripID = "11111111-2222-3333-4444-555555555555"
)

// fakeClaimsValidator is a hand fake for middleware.TokenValidator (one
// method, per docs/mock-standards.md), letting these tests drive the real
// Authenticate gate without RSA keys — mirrors auth/handler_test.go's
// identical helper.
type fakeClaimsValidator struct {
	userID string
}

func (f *fakeClaimsValidator) ValidateToken(_ context.Context, _ string) (middleware.AuthClaims, error) {
	return middleware.AuthClaims{UserID: f.userID}, nil
}

// fakeRoleLookup hand-fakes trip.RoleLookup, the small consumer-owned port
// `make mocks` would otherwise auto-generate (see docs/mock-standards.md and
// .github/memory/patterns-discovered.md's "mockery --all Regenerates a Mock
// for Every Interface..." entry) — its generated mock is deliberately not
// committed.
type fakeRoleLookup struct {
	user *auth.User
	err  error
}

func (f *fakeRoleLookup) GetUserByID(_ context.Context, _ string) (*auth.User, error) {
	return f.user, f.err
}

func adminLookup() *fakeRoleLookup {
	return &fakeRoleLookup{user: &auth.User{ID: handlerUserID, Role: auth.RoleAdmin}}
}

// newGatedRouter mounts trip.Handler's routes behind the real
// middleware.Authenticate, driving the actual gate rather than hand-injecting
// a context value — mirrors auth/handler_test.go's newGatedAuthRouter.
func newGatedRouter(t *testing.T, h *trip.Handler) http.Handler {
	t.Helper()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route("/api/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(&fakeClaimsValidator{userID: handlerUserID}))
			h.RegisterRoutes(r)
		})
	})
	return r
}

// newUngatedRouter mounts the routes with no Authenticate middleware at all,
// for the "no user in context" defense-in-depth tests.
func newUngatedRouter(h *trip.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route("/api/v1", func(r chi.Router) {
		h.RegisterRoutes(r)
	})
	return r
}

func doAuthenticatedRequest(
	t *testing.T, router http.Handler, method, path, body string, headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "any-non-empty-token"})
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return body
}

// --- Create -----------------------------------------------------------

func TestUnitHandleCreate_AdminValidBody_Returns201WithTripEnvelope(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	want := &trip.Trip{ID: handlerTripID, CreatorID: handlerUserID, Title: "Japan 2027", Status: trip.TripStatusDraft}
	svc.EXPECT().
		Create(mock.Anything, handlerUserID, auth.RoleAdmin, "Japan 2027", (*string)(nil)).
		Return(want, nil).
		Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, "/api/v1/trips",
		`{"title":"Japan 2027"}`, nil)

	require.Equal(t, http.StatusCreated, rec.Code)
	body := decodeBody(t, rec)
	tripBody, ok := body["trip"].(map[string]any)
	require.True(t, ok, "response must wrap the trip in a \"trip\" envelope")
	assert.Equal(t, handlerTripID, tripBody["id"])
}

func TestUnitHandleCreate_MalformedJSON_Returns400InvalidRequest(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	h := trip.NewHandler(svc, adminLookup())

	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, "/api/v1/trips",
		`{not json`, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "invalid_request", decodeBody(t, rec)["error"])
	svc.AssertNotCalled(t, "Create")
}

func TestUnitHandleCreate_NonAdminRole_Returns403Forbidden(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	svc.EXPECT().
		Create(mock.Anything, handlerUserID, auth.RolePartner, mock.Anything, mock.Anything).
		Return(nil, domainerrors.Forbidden("Only admin users can create trips")).
		Once()
	roles := &fakeRoleLookup{user: &auth.User{ID: handlerUserID, Role: auth.RolePartner}}

	h := trip.NewHandler(svc, roles)
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, "/api/v1/trips",
		`{"title":"Japan 2027"}`, nil)

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "forbidden", decodeBody(t, rec)["error"])
}

func TestUnitHandleCreate_RoleLookupFails_PropagatesError(t *testing.T) {
	svc := tripmocks.NewMockManager(t) // no expects: must not be reached
	roles := &fakeRoleLookup{err: domainerrors.NotFound("user", handlerUserID)}

	h := trip.NewHandler(svc, roles)
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, "/api/v1/trips",
		`{"title":"Japan 2027"}`, nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertNotCalled(t, "Create")
}

func TestUnitHandleCreate_NoUserInContext_Returns401(t *testing.T) {
	svc := tripmocks.NewMockManager(t) // no expects: must not be reached

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newUngatedRouter(h), http.MethodPost, "/api/v1/trips", `{"title":"x"}`, nil)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "authentication_required", decodeBody(t, rec)["error"])
	svc.AssertNotCalled(t, "Create")
}

// --- List ---------------------------------------------------------------

func TestUnitHandleList_Returns200WithTripsEnvelope(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	want := []*trip.Trip{{ID: handlerTripID, CreatorID: handlerUserID}}
	svc.EXPECT().List(mock.Anything, handlerUserID).Return(want, nil).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodGet, "/api/v1/trips", "", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	body := decodeBody(t, rec)
	trips, ok := body["trips"].([]any)
	require.True(t, ok, "response must wrap the list in a \"trips\" envelope")
	assert.Len(t, trips, 1)
}

func TestUnitHandleList_NoUserInContext_Returns401(t *testing.T) {
	svc := tripmocks.NewMockManager(t) // no expects: must not be reached

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newUngatedRouter(h), http.MethodGet, "/api/v1/trips", "", nil)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}

// --- Get ------------------------------------------------------------------

func TestUnitHandleGet_OwnedTrip_Returns200WithTripEnvelope(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	want := &trip.Trip{ID: handlerTripID, CreatorID: handlerUserID}
	svc.EXPECT().Get(mock.Anything, handlerUserID, handlerTripID).Return(want, nil).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodGet, "/api/v1/trips/"+handlerTripID, "", nil)

	require.Equal(t, http.StatusOK, rec.Code)
	body := decodeBody(t, rec)
	tripBody := body["trip"].(map[string]any)
	assert.Equal(t, handlerTripID, tripBody["id"])
}

func TestUnitHandleGet_TripNotFound_Returns404(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	svc.EXPECT().Get(mock.Anything, handlerUserID, handlerTripID).
		Return(nil, domainerrors.NotFound("trip", handlerTripID)).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodGet, "/api/v1/trips/"+handlerTripID, "", nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "not_found", decodeBody(t, rec)["error"])
}

// --- Update -----------------------------------------------------------

func TestUnitHandleUpdate_ValidRequest_Returns200WithTripEnvelope(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	want := &trip.Trip{ID: handlerTripID, CreatorID: handlerUserID, Title: "Updated", Version: 4}
	svc.EXPECT().
		Update(mock.Anything, handlerUserID, handlerTripID, 3, "Updated", (*string)(nil), trip.TripStatusPublished).
		Return(want, nil).
		Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPut, "/api/v1/trips/"+handlerTripID,
		`{"title":"Updated","status":"published"}`, map[string]string{"If-Match": "3"})

	require.Equal(t, http.StatusOK, rec.Code)
	body := decodeBody(t, rec)
	tripBody := body["trip"].(map[string]any)
	assert.Equal(t, float64(4), tripBody["version"])
}

func TestUnitHandleUpdate_MissingIfMatch_Returns400InvalidRequest(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	h := trip.NewHandler(svc, adminLookup())

	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPut, "/api/v1/trips/"+handlerTripID,
		`{"title":"Updated","status":"published"}`, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "invalid_request", decodeBody(t, rec)["error"])
	svc.AssertNotCalled(t, "Update")
}

func TestUnitHandleUpdate_NonIntegerIfMatch_Returns400InvalidRequest(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	h := trip.NewHandler(svc, adminLookup())

	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPut, "/api/v1/trips/"+handlerTripID,
		`{"title":"Updated","status":"published"}`, map[string]string{"If-Match": "not-a-number"})

	require.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertNotCalled(t, "Update")
}

func TestUnitHandleUpdate_MalformedJSON_Returns400InvalidRequest(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	h := trip.NewHandler(svc, adminLookup())

	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPut, "/api/v1/trips/"+handlerTripID,
		`{not json`, map[string]string{"If-Match": "3"})

	require.Equal(t, http.StatusBadRequest, rec.Code)
	svc.AssertNotCalled(t, "Update")
}

func TestUnitHandleUpdate_ConcurrentModification_Returns409Conflict(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	svc.EXPECT().Update(mock.Anything, handlerUserID, handlerTripID, 1, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domainerrors.Conflict("trip was modified concurrently")).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPut, "/api/v1/trips/"+handlerTripID,
		`{"title":"Updated","status":"draft"}`, map[string]string{"If-Match": "1"})

	require.Equal(t, http.StatusConflict, rec.Code)
	assert.Equal(t, "conflict", decodeBody(t, rec)["error"])
}

func TestUnitHandleUpdate_NotOwnerOrNoSubscription_Returns403Forbidden(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	svc.EXPECT().Update(mock.Anything, handlerUserID, handlerTripID, 1, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, domainerrors.Forbidden("An active subscription is required to update this trip")).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPut, "/api/v1/trips/"+handlerTripID,
		`{"title":"Updated","status":"draft"}`, map[string]string{"If-Match": "1"})

	require.Equal(t, http.StatusForbidden, rec.Code)
}

// --- Delete -----------------------------------------------------------

func TestUnitHandleDelete_ActiveSubscriptionOwner_Returns204(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	svc.EXPECT().Delete(mock.Anything, handlerUserID, handlerTripID).Return(nil).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodDelete, "/api/v1/trips/"+handlerTripID, "", nil)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.Bytes())
}

func TestUnitHandleDelete_NotOwner_Returns403Forbidden(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	svc.EXPECT().Delete(mock.Anything, handlerUserID, handlerTripID).
		Return(domainerrors.Forbidden("Only the trip creator can delete this trip")).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodDelete, "/api/v1/trips/"+handlerTripID, "", nil)

	require.Equal(t, http.StatusForbidden, rec.Code)
}

func TestUnitHandleDelete_TripNotFound_Returns404(t *testing.T) {
	svc := tripmocks.NewMockManager(t)
	svc.EXPECT().Delete(mock.Anything, handlerUserID, handlerTripID).
		Return(domainerrors.NotFound("trip", handlerTripID)).Once()

	h := trip.NewHandler(svc, adminLookup())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodDelete, "/api/v1/trips/"+handlerTripID, "", nil)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
