// Package errors defines the backend's domain error types and the HTTP
// handler that maps them to the project's standard JSON error envelope
// (docs/api-design-standards.md §7). Every HTTP handler in this codebase
// should return errors built from this package's constructors rather than
// writing error responses ad hoc, so the wire format stays consistent.
package errors

import "fmt"

// DomainError represents a business-rule failure that a handler can turn
// into a standard HTTP error response via HandleError. Err carries an
// optional wrapped cause (e.g. a lower-layer error) for errors.Is/errors.As
// and for server-side logging, without exposing that cause to the client.
type DomainError struct {
	// Code is the machine-readable wire error code (docs/api-design-standards.md
	// §7, e.g. "not_found"); HandleError also uses it to pick the HTTP status.
	Code    string
	Message string
	// Fields holds per-field validation failures, serialized as the
	// envelope's optional "fields" array; empty/nil for non-validation errors.
	Fields []ValidationError
	// Details is optional free-form machine-readable context for the client,
	// serialized as the envelope's optional "details" object.
	Details map[string]any
	Err     error
}

// ValidationError is a single field-level validation failure, matching the
// "fields" entries of the standard error envelope (json tags "field"/"error").
type ValidationError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// Error implements the error interface, including the wrapped cause (if any)
// for server-side logs and debugging — it is never sent to clients verbatim.
func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the wrapped cause so errors.Is/errors.As can see through it.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NotFound builds a "not_found" domain error for a missing resource.
func NotFound(resource, id string) *DomainError {
	return &DomainError{
		Code:    "not_found",
		Message: fmt.Sprintf("%s with ID %q not found", resource, id),
	}
}

// Validation builds a "validation_failed" domain error, optionally carrying
// field-level details for the response's "fields" array.
func Validation(message string, fields ...ValidationError) *DomainError {
	return &DomainError{
		Code:    "validation_failed",
		Message: message,
		Fields:  fields,
	}
}

// Unauthorized builds an "authentication_required" domain error. The
// constructor name reflects the caller's intent (missing/invalid
// credentials); the wire error code stays "authentication_required" per
// docs/api-design-standards.md §7.
func Unauthorized(message string) *DomainError {
	return &DomainError{
		Code:    "authentication_required",
		Message: message,
	}
}

// Forbidden builds a "forbidden" domain error for an authenticated caller
// lacking sufficient permissions.
func Forbidden(message string) *DomainError {
	return &DomainError{
		Code:    "forbidden",
		Message: message,
	}
}

// Conflict builds a "conflict" domain error, e.g. an optimistic-locking
// version mismatch.
func Conflict(message string) *DomainError {
	return &DomainError{
		Code:    "conflict",
		Message: message,
	}
}
