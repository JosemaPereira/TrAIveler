// Package example is the canonical reference implementation of this
// backend's layered pattern (model -> repository -> service -> handler).
// It is not a production feature: future domain packages (Trip, User, ...)
// should copy this package's structure rather than import it.
package example

import (
	"strings"
	"time"

	domainerrors "github.com/JosemaPereira/TrAIveler/backend/internal/errors"
)

// Status values accepted by Example.Status. Stored as a VARCHAR with a CHECK
// constraint (see migrations/*_create_examples_table.sql), mirrored here so
// callers don't need to hardcode the wire strings.
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Example is the reference domain entity. Field tags demonstrate this
// project's convention: snake_case "json" tags for the API wire format
// (docs/api-design-standards.md §5) and matching "db" tags for repository
// scanning.
type Example struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Status    string    `json:"status" db:"status"`
	Count     int       `json:"count" db:"count"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Version   int       `json:"version" db:"version"`
}

// Validate performs business-rule validation, collecting every failing
// field rather than stopping at the first one so a client can fix all
// issues in a single round trip. Returns nil when ex is valid.
func (e *Example) Validate() *domainerrors.DomainError {
	var fields []domainerrors.ValidationError

	if strings.TrimSpace(e.Name) == "" {
		fields = append(fields, domainerrors.ValidationError{
			Field: "name",
			Error: "Name is required",
		})
	}

	if strings.TrimSpace(e.Email) == "" {
		fields = append(fields, domainerrors.ValidationError{
			Field: "email",
			Error: "Email is required",
		})
	}

	if e.Status != StatusActive && e.Status != StatusInactive {
		fields = append(fields, domainerrors.ValidationError{
			Field: "status",
			Error: "Status must be 'active' or 'inactive'",
		})
	}

	if len(fields) > 0 {
		return domainerrors.Validation("One or more fields failed validation", fields...)
	}

	return nil
}
