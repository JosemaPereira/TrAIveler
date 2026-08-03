//go:build test

package conversation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/conversation"
	convmocks "github.com/JosemaPereira/TrAIveler/backend/internal/conversation/mocks"
	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	"github.com/JosemaPereira/TrAIveler/backend/internal/trip"
)

const (
	handlerUserID    = "9f1c1a1e-2b3d-4c5e-8f6a-1234567890ab"
	handlerTripID    = "11111111-2222-3333-4444-555555555555"
	handlerSessionID = "99999999-8888-7777-6666-555555555555"
)

// fakeClaimsValidator is a hand fake for middleware.TokenValidator, mirroring
// auth/handler_test.go's and trip/handler_test.go's identical helper.
type fakeClaimsValidator struct {
	userID string
}

func (f *fakeClaimsValidator) ValidateToken(_ context.Context, _ string) (middleware.AuthClaims, error) {
	return middleware.AuthClaims{UserID: f.userID}, nil
}

// fakeTripAccess hand-fakes conversation.TripAccess, the small consumer-owned
// port `make mocks` would otherwise auto-generate (see docs/mock-standards.md
// and .github/memory/patterns-discovered.md's "mockery --all Regenerates a
// Mock for Every Interface..." entry) — its generated mock is deliberately
// not committed.
type fakeTripAccess struct {
	tr  *trip.Trip
	err error
}

func (f *fakeTripAccess) Get(_ context.Context, _, _ string) (*trip.Trip, error) {
	return f.tr, f.err
}

func ownedTripAccess() *fakeTripAccess {
	return &fakeTripAccess{tr: &trip.Trip{ID: handlerTripID, CreatorID: handlerUserID}}
}

// conversationPath is the single URL every test in this file exercises:
// the two Conversation endpoints share one path, distinguished only by
// HTTP method.
const conversationPath = "/api/v1/trips/" + handlerTripID + "/conversation"

// newGatedRouter mounts conversation.Handler's routes behind the real
// middleware.Authenticate for handlerUserID, mirroring
// trip/handler_test.go's identical helper.
func newGatedRouter(t *testing.T, h *conversation.Handler) http.Handler {
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

func newUngatedRouter(h *conversation.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Route("/api/v1", func(r chi.Router) {
		h.RegisterRoutes(r)
	})
	return r
}

func doAuthenticatedRequest(t *testing.T, router http.Handler, method, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, conversationPath, bytes.NewBufferString(body))
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "any-non-empty-token"})
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

// --- POST (send message) -----------------------------------------------

func TestUnitHandleSendMessage_ValidRequest_ReturnsSingleSSEEvent(t *testing.T) {
	svc := convmocks.NewMockManager(t)
	svc.EXPECT().
		SendMessage(mock.Anything, handlerTripID, "I have 10 days in Japan").
		Return(&conversation.Message{
			SessionID: handlerSessionID,
			Role:      conversation.RoleAssistant,
			Content:   "Do you want joint activities?",
		}, false, nil).
		Once()

	h := conversation.NewHandler(svc, ownedTripAccess())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, `{"message":"I have 10 days in Japan"}`)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/event-stream", rec.Header().Get("Content-Type"))

	raw := rec.Body.String()
	require.True(t, strings.HasPrefix(raw, "data: "), "must be framed as a single SSE data event")
	require.True(t, strings.HasSuffix(raw, "\n\n"), "SSE frame must end with a blank line")
	require.Equal(t, 1, strings.Count(raw, "data: "), "must carry exactly one event per turn (Continue is synchronous)")

	payload := strings.TrimSuffix(strings.TrimPrefix(raw, "data: "), "\n\n")
	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(payload), &event))
	assert.Equal(t, handlerSessionID, event["session_id"])
	assert.Equal(t, "assistant", event["role"])
	assert.Equal(t, "Do you want joint activities?", event["message"])
	assert.Equal(t, false, event["itinerary_ready"])
}

func TestUnitHandleSendMessage_ItineraryReady_CarriesTrueFlag(t *testing.T) {
	svc := convmocks.NewMockManager(t)
	svc.EXPECT().SendMessage(mock.Anything, handlerTripID, mock.Anything).
		Return(&conversation.Message{SessionID: handlerSessionID, Role: conversation.RoleAssistant, Content: "Your itinerary is ready!"}, true, nil).
		Once()

	h := conversation.NewHandler(svc, ownedTripAccess())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, `{"message":"looks good"}`)

	require.Equal(t, http.StatusOK, rec.Code)
	payload := strings.TrimSuffix(strings.TrimPrefix(rec.Body.String(), "data: "), "\n\n")
	var event map[string]any
	require.NoError(t, json.Unmarshal([]byte(payload), &event))
	assert.Equal(t, true, event["itinerary_ready"])
}

func TestUnitHandleSendMessage_MalformedJSON_Returns400InvalidRequest(t *testing.T) {
	svc := convmocks.NewMockManager(t)
	h := conversation.NewHandler(svc, ownedTripAccess())

	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, `{not json`)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "invalid_request", decodeBody(t, rec)["error"])
	assert.NotEqual(t, "text/event-stream", rec.Header().Get("Content-Type"))
	svc.AssertNotCalled(t, "SendMessage")
}

func TestUnitHandleSendMessage_EmptyMessage_Returns422ValidationFailed(t *testing.T) {
	svc := convmocks.NewMockManager(t)
	h := conversation.NewHandler(svc, ownedTripAccess())

	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, `{"message":"   "}`)

	require.Equal(t, http.StatusUnprocessableEntity, rec.Code)
	svc.AssertNotCalled(t, "SendMessage")
}

// TestUnitHandleSendMessage_TripOwnedBySomeoneElse_Returns404NotFound proves
// the ownership check runs BEFORE any SSE header is committed: the handler
// must still be able to answer with the normal JSON error envelope, not a
// half-open text/event-stream response.
func TestUnitHandleSendMessage_TripOwnedBySomeoneElse_Returns404NotFound(t *testing.T) {
	svc := convmocks.NewMockManager(t) // no expects: must not be reached
	access := &fakeTripAccess{err: domainerrors.NotFound("trip", handlerTripID)}

	h := conversation.NewHandler(svc, access)
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, `{"message":"hello"}`)

	require.Equal(t, http.StatusNotFound, rec.Code)
	assert.NotEqual(t, "text/event-stream", rec.Header().Get("Content-Type"))
	svc.AssertNotCalled(t, "SendMessage")
}

func TestUnitHandleSendMessage_ServiceError_Returns503ThroughNormalErrorPath(t *testing.T) {
	svc := convmocks.NewMockManager(t)
	svc.EXPECT().SendMessage(mock.Anything, handlerTripID, mock.Anything).
		Return(nil, false, domainerrors.ServiceUnavailable(30)).Once()

	h := conversation.NewHandler(svc, ownedTripAccess())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodPost, `{"message":"hello"}`)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.NotEqual(t, "text/event-stream", rec.Header().Get("Content-Type"))
}

func TestUnitHandleSendMessage_NoUserInContext_Returns401(t *testing.T) {
	svc := convmocks.NewMockManager(t) // no expects: must not be reached

	h := conversation.NewHandler(svc, ownedTripAccess())
	rec := doAuthenticatedRequest(t, newUngatedRouter(h), http.MethodPost, `{"message":"hello"}`)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	svc.AssertNotCalled(t, "SendMessage")
}

// --- GET (history) -------------------------------------------------------

func TestUnitHandleGetHistory_ReturnsSessionAndMessagesInOrder(t *testing.T) {
	svc := convmocks.NewMockManager(t)
	messages := []*conversation.Message{
		{ID: "m1", SessionID: handlerSessionID, Role: conversation.RoleUser, Content: "hi"},
		{ID: "m2", SessionID: handlerSessionID, Role: conversation.RoleAssistant, Content: "hello"},
	}
	svc.EXPECT().GetHistory(mock.Anything, handlerTripID).
		Return(&conversation.Session{ID: handlerSessionID, Status: conversation.StatusInProgress}, messages, nil).
		Once()

	h := conversation.NewHandler(svc, ownedTripAccess())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodGet, "")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	body := decodeBody(t, rec)
	assert.Equal(t, handlerSessionID, body["session_id"])
	assert.Equal(t, "in_progress", body["status"])
	msgs, ok := body["messages"].([]any)
	require.True(t, ok)
	require.Len(t, msgs, 2)
	assert.Equal(t, "hi", msgs[0].(map[string]any)["content"])
	assert.Equal(t, "hello", msgs[1].(map[string]any)["content"])
}

func TestUnitHandleGetHistory_TripOwnedBySomeoneElse_Returns404NotFound(t *testing.T) {
	svc := convmocks.NewMockManager(t) // no expects: must not be reached
	access := &fakeTripAccess{err: domainerrors.NotFound("trip", handlerTripID)}

	h := conversation.NewHandler(svc, access)
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodGet, "")

	require.Equal(t, http.StatusNotFound, rec.Code)
	svc.AssertNotCalled(t, "GetHistory")
}

func TestUnitHandleGetHistory_NoConversationYet_Returns404(t *testing.T) {
	svc := convmocks.NewMockManager(t)
	svc.EXPECT().GetHistory(mock.Anything, handlerTripID).
		Return(nil, nil, domainerrors.NotFound("conversation session for trip", handlerTripID)).Once()

	h := conversation.NewHandler(svc, ownedTripAccess())
	rec := doAuthenticatedRequest(t, newGatedRouter(t, h), http.MethodGet, "")

	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUnitHandleGetHistory_NoUserInContext_Returns401(t *testing.T) {
	svc := convmocks.NewMockManager(t) // no expects: must not be reached

	h := conversation.NewHandler(svc, ownedTripAccess())
	rec := doAuthenticatedRequest(t, newUngatedRouter(h), http.MethodGet, "")

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
