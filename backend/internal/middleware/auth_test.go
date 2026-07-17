//go:build test

package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mw "github.com/JosemaPereira/TrAIveler/backend/internal/middleware"
	mwmocks "github.com/JosemaPereira/TrAIveler/backend/internal/middleware/mocks"
)

// accessTokenCookie mirrors the unexported constant in the middleware package;
// this external test package can't reach it, so the wire name is repeated here.
const accessTokenCookie = "access_token"

func TestAuthenticate_InvalidCredentials_Returns401AndSkipsNext(t *testing.T) {
	tests := []struct {
		name      string
		cookie    *http.Cookie
		setupMock func(m *mwmocks.MockTokenValidator)
	}{
		{
			name:      "when no access_token cookie is present it should reject",
			cookie:    nil,
			setupMock: func(*mwmocks.MockTokenValidator) {}, // validator must not be consulted
		},
		{
			name:      "when the access_token cookie is empty it should reject",
			cookie:    &http.Cookie{Name: accessTokenCookie, Value: ""},
			setupMock: func(*mwmocks.MockTokenValidator) {},
		},
		{
			name:   "when the token fails validation it should reject",
			cookie: &http.Cookie{Name: accessTokenCookie, Value: "garbage.token.value"},
			setupMock: func(m *mwmocks.MockTokenValidator) {
				m.EXPECT().
					ValidateToken(mock.Anything, "garbage.token.value").
					Return(mw.AuthClaims{}, errors.New("invalid or expired token"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := mwmocks.NewMockTokenValidator(t)
			tt.setupMock(validator)

			nextRan := false
			handler := mw.Authenticate(validator)(http.HandlerFunc(
				func(http.ResponseWriter, *http.Request) { nextRan = true }))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code)
			assert.False(t, nextRan, "an unauthenticated request must never reach the handler")
			assert.Contains(t, rec.Body.String(), "authentication_required")
		})
	}
}

func TestAuthenticate_ValidToken_AttachesUserContextAndCallsNext(t *testing.T) {
	validator := mwmocks.NewMockTokenValidator(t)
	validator.EXPECT().
		ValidateToken(mock.Anything, "valid.jwt.token").
		Return(mw.AuthClaims{UserID: "user-123", HasSubscription: true}, nil)

	var (
		nextRan bool
		gotUser string
		gotSub  bool
	)
	handler := mw.Authenticate(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextRan = true
		gotUser, _ = mw.UserIDFromContext(r.Context())
		gotSub, _ = mw.HasSubscriptionFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: accessTokenCookie, Value: "valid.jwt.token"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.True(t, nextRan, "a valid token must reach the handler")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "user-123", gotUser)
	assert.True(t, gotSub)
}
