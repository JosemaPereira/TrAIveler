package errors_test

import (
	"errors"
	"fmt"
	"testing"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeConnectivityError stands in for a pgx connection-level failure (dial
// refused, timeout, DNS failure, ...). pgx's own such errors are unexported
// (pgconn.pgconnError et al.) and only constructible by actually dialing, so
// this minimal type satisfies the same duck-typed `interface{ SafeToRetry()
// bool }` that pgconn.SafeToRetry checks for via errors.As, without a real
// network failure.
type fakeConnectivityError struct {
	cause error
}

func (e *fakeConnectivityError) Error() string     { return fmt.Sprintf("dial: %s", e.cause) }
func (e *fakeConnectivityError) Unwrap() error     { return e.cause }
func (e *fakeConnectivityError) SafeToRetry() bool { return true }

func TestServiceUnavailableFromDB_ConnectivityFailure_ReturnsServiceUnavailableDomainError(t *testing.T) {
	connErr := &fakeConnectivityError{cause: errors.New("connection refused")}

	got := domainerrors.ServiceUnavailableFromDB(connErr, 30)

	require.NotNil(t, got)
	var domainErr *domainerrors.DomainError
	require.True(t, errors.As(got, &domainErr))
	assert.Equal(t, "service_unavailable", domainErr.Code)
	assert.Equal(t, 30, domainErr.Details["retry_after_seconds"])
}

func TestServiceUnavailableFromDB_QueryLevelFailure_ReturnsNil(t *testing.T) {
	got := domainerrors.ServiceUnavailableFromDB(errors.New("syntax error at or near"), 30)

	assert.Nil(t, got)
}

func TestServiceUnavailableFromDB_NilError_ReturnsNil(t *testing.T) {
	got := domainerrors.ServiceUnavailableFromDB(nil, 30)

	assert.Nil(t, got)
}
