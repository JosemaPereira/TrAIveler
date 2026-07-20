package auth

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JosemaPereira/TrAIveler/backend/internal/subscription"
)

func TestUnitUserValidate_ValidUser_ReturnsNil(t *testing.T) {
	u := &User{Email: "ada@example.com", Role: RoleAdmin}

	assert.Nil(t, u.Validate())
}

func TestUnitUserValidate_MissingEmail_ReturnsValidationError(t *testing.T) {
	u := &User{Email: "  ", Role: RolePartner}

	err := u.Validate()

	require.NotNil(t, err)
	assert.Equal(t, "validation_failed", err.Code)
	require.Len(t, err.Fields, 1)
	assert.Equal(t, "email", err.Fields[0].Field)
}

func TestUnitUserValidate_InvalidRole_ReturnsValidationError(t *testing.T) {
	u := &User{Email: "ada@example.com", Role: "superuser"}

	err := u.Validate()

	require.NotNil(t, err)
	require.Len(t, err.Fields, 1)
	assert.Equal(t, "role", err.Fields[0].Field)
}

// TestUnitUserMarshalJSON_ExcludesSensitiveAndInternalFields locks in the
// contract that a User serialized into a response never leaks the password
// hash, failed-login counters, verification flag, role, updated_at, or the
// optimistic-lock version (specs/008-auth-collaboration-ux/contracts/api.md).
func TestUnitUserMarshalJSON_ExcludesSensitiveAndInternalFields(t *testing.T) {
	lastFailed := time.Now()
	u := &User{
		ID:                  "user-1",
		Email:               "ada@example.com",
		PasswordHash:        "$2a$12$supersecrethash",
		Role:                RoleAdmin,
		FullName:            "Ada Lovelace",
		HasSubscription:     true,
		FailedLoginAttempts: 3,
		LastFailedLoginAt:   &lastFailed,
		EmailVerified:       true,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
		Version:             7,
	}

	raw, err := json.Marshal(u)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))

	// Present, contract fields.
	for _, k := range []string{"id", "email", "full_name", "has_subscription", "created_at"} {
		assert.Contains(t, out, k, "expected field %q in user JSON", k)
	}
	// Absent, sensitive/internal fields.
	for _, k := range []string{
		"password_hash", "role", "failed_login_attempts", "last_failed_login_at",
		"email_verified", "updated_at", "version",
	} {
		assert.NotContains(t, out, k, "field %q must not appear in user JSON", k)
	}
	assert.NotContains(t, string(raw), "supersecrethash")
}

func TestUnitRegisterRequestValidate_ValidRequest_ReturnsNil(t *testing.T) {
	req := &RegisterRequest{
		Email:    "ada@example.com",
		Password: "SecureP@ss123",
		FullName: "Ada Lovelace",
	}

	assert.Nil(t, req.Validate())
}

func TestUnitRegisterRequestValidate_InvalidFields_ReturnsFieldErrors(t *testing.T) {
	tests := []struct {
		name      string
		req       RegisterRequest
		wantField string
	}{
		{
			name:      "when the email is malformed it should reject it",
			req:       RegisterRequest{Email: "not-an-email", Password: "SecureP@ss123", FullName: "Ada"},
			wantField: "email",
		},
		{
			name:      "when the email is empty it should reject it",
			req:       RegisterRequest{Email: "", Password: "SecureP@ss123", FullName: "Ada"},
			wantField: "email",
		},
		{
			name:      "when the full name is blank it should reject it",
			req:       RegisterRequest{Email: "ada@example.com", Password: "SecureP@ss123", FullName: "   "},
			wantField: "full_name",
		},
		{
			name:      "when the full name exceeds 100 chars it should reject it",
			req:       RegisterRequest{Email: "ada@example.com", Password: "SecureP@ss123", FullName: strings.Repeat("a", 101)},
			wantField: "full_name",
		},
		{
			name:      "when the password is weak it should reject it",
			req:       RegisterRequest{Email: "ada@example.com", Password: "weak", FullName: "Ada"},
			wantField: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			require.NotNil(t, err)
			assert.Equal(t, "validation_failed", err.Code)
			require.NotEmpty(t, err.Fields)
			assert.Equal(t, tt.wantField, err.Fields[0].Field)
		})
	}
}

func TestUnitRegisterRequestValidate_MultipleFailures_ReportsAll(t *testing.T) {
	req := &RegisterRequest{Email: "bad", Password: "weak", FullName: ""}

	err := req.Validate()

	require.NotNil(t, err)
	// email + full_name + at least one password rule.
	assert.GreaterOrEqual(t, len(err.Fields), 3)
}

func TestUnitLoginRequestValidate_ValidRequest_ReturnsNil(t *testing.T) {
	req := &LoginRequest{Email: "ada@example.com", Password: "anything"}

	assert.Nil(t, req.Validate())
}

func TestUnitLoginRequestValidate_InvalidFields_ReturnsFieldErrors(t *testing.T) {
	tests := []struct {
		name      string
		req       LoginRequest
		wantField string
	}{
		{
			name:      "when the email is malformed it should reject it",
			req:       LoginRequest{Email: "nope", Password: "anything"},
			wantField: "email",
		},
		{
			name:      "when the password is empty it should reject it",
			req:       LoginRequest{Email: "ada@example.com", Password: ""},
			wantField: "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()

			require.NotNil(t, err)
			require.NotEmpty(t, err.Fields)
			assert.Equal(t, tt.wantField, err.Fields[0].Field)
		})
	}
}

// TestUnitLoginRequestValidate_DoesNotEnforcePasswordStrength confirms login
// only checks presence, not strength — a short/simple password must still be
// accepted so credential checking happens against the stored hash.
func TestUnitLoginRequestValidate_DoesNotEnforcePasswordStrength(t *testing.T) {
	req := &LoginRequest{Email: "ada@example.com", Password: "x"}

	assert.Nil(t, req.Validate())
}

func TestUnitRegisterResponseMarshalJSON_OmitsSubscriptionWhenNil(t *testing.T) {
	resp := RegisterResponse{User: User{ID: "user-1", Email: "ada@example.com"}}

	raw, err := json.Marshal(resp)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.NotContains(t, out, "subscription")
}

func TestUnitRegisterResponseMarshalJSON_IncludesSubscriptionWhenPresent(t *testing.T) {
	resp := RegisterResponse{
		User:         User{ID: "user-1", Email: "ada@example.com"},
		Subscription: &subscription.Subscription{ID: "sub-1", Status: subscription.StatusActive},
	}

	raw, err := json.Marshal(resp)
	require.NoError(t, err)

	var out map[string]any
	require.NoError(t, json.Unmarshal(raw, &out))
	assert.Contains(t, out, "subscription")
}
