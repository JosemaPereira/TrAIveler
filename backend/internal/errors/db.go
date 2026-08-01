package errors

import "github.com/jackc/pgx/v5/pgconn"

// ServiceUnavailableFromDB classifies a database-layer failure, returning a
// "service_unavailable" DomainError (via ServiceUnavailable) when err is a
// connection-level failure — refused, timed out, DNS failure, and other
// cases guaranteed to have occurred before any data reached the server, per
// pgx's pgconn.SafeToRetry — and nil for anything else (a normal query
// error: not-found, constraint violation, syntax error, ...), so the caller
// falls through to its usual handling. Introduced for the refresh endpoint
// (issue #192 Task A, specs/008-auth-collaboration-ux/contracts/api.md's
// "503 Service Unavailable: Database unavailable"), but deliberately generic
// so other DB-backed callers can adopt it without rework:
//
//	if svcErr := errors.ServiceUnavailableFromDB(err, 30); svcErr != nil {
//	    return svcErr
//	}
//	return fmt.Errorf("...: %w", err)
//
// retryAfterSeconds is carried in the returned DomainError's Details, same as
// ServiceUnavailable/RateLimited, so writeErrorResponse (handler.go) can echo
// it back as the response's Retry-After header.
func ServiceUnavailableFromDB(err error, retryAfterSeconds int) error {
	if err == nil || !pgconn.SafeToRetry(err) {
		return nil
	}
	return ServiceUnavailable(retryAfterSeconds)
}
