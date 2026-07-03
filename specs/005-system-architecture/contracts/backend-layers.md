# Backend Layer Contracts

**Feature**: System Architecture and Technology Stack  
**Created**: 2026-07-03  
**Layer**: Backend (Go REST API)

## Overview

This document defines the integration contracts between backend architectural layers: Handler (presentation), Service (business logic), and Repository (data access). These contracts ensure loose coupling, testability, and consistent error propagation across layers.

---

## Layer Responsibilities

### Handler Layer (Presentation)

**Responsibilities**:
- Parse HTTP request (path params, query params, request body)
- Validate request format (JSON schema, required fields)
- Call service layer methods
- Map service responses to HTTP responses (status codes, JSON body)
- Handle errors via ErrorHandler

**Must NOT**:
- Contain business logic
- Access database directly
- Call external APIs directly

**Location**: `internal/<domain>/handler.go`

---

### Service Layer (Business Logic)

**Responsibilities**:
- Validate business rules
- Orchestrate repository calls and external API calls
- Apply domain transformations
- Return domain errors (not HTTP status codes)

**Must NOT**:
- Import `net/http` package
- Return HTTP-specific types
- Log directly (logging handled by middleware)

**Location**: `internal/<domain>/service.go`

---

### Repository Layer (Data Access)

**Responsibilities**:
- Execute SQL queries (SELECT, INSERT, UPDATE, DELETE)
- Map database rows to domain models
- Wrap database errors with context

**Must NOT**:
- Contain business logic
- Return raw SQL rows
- Log query results

**Location**: `internal/<domain>/repository.go`

---

## Handler ↔ Service Contract

### Request Flow

1. Handler parses HTTP request into service request struct
2. Handler calls service method with request struct
3. Service validates business rules, calls repository/AI client
4. Service returns domain response or domain error
5. Handler maps domain response to HTTP response via ErrorHandler

### Example Contract (Trip Creation)

**Service Interface**:
```go
type TripService interface {
    CreateTrip(ctx context.Context, req CreateTripRequest) (*Trip, error)
}

type CreateTripRequest struct {
    Title        string    `json:"title"`
    StartDate    time.Time `json:"start_date"`
    EndDate      time.Time `json:"end_date"`
    UserID       string    // Injected from auth context
}
```

**Handler Implementation**:
```go
func (h *TripHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
    var req CreateTripRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.HandleError(w, r, errors.ErrInvalidJSON)
        return
    }

    // Inject user ID from auth context
    req.UserID = auth.GetUserID(r.Context())

    trip, err := h.service.CreateTrip(r.Context(), req)
    if err != nil {
        errors.HandleError(w, r, err)
        return
    }

    json.NewEncoder(w).Encode(trip)
}
```

**Domain Errors**:
- `ErrValidation`: Business rule violation (e.g., end_date before start_date)
- `ErrUnauthorized`: User not authenticated
- `ErrForbidden`: User not authorized for this action
- `ErrNotFound`: Resource not found
- `ErrConflict`: Resource already exists or version mismatch

**Error Mapping** (in `internal/errors/handler.go`):
```go
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
    var status int
    var errorCode string

    switch {
    case errors.Is(err, ErrNotFound):
        status = http.StatusNotFound
        errorCode = "resource_not_found"
    case errors.Is(err, ErrValidation):
        status = http.StatusBadRequest
        errorCode = "validation_error"
    case errors.Is(err, ErrUnauthorized):
        status = http.StatusUnauthorized
        errorCode = "unauthorized"
    case errors.Is(err, ErrForbidden):
        status = http.StatusForbidden
        errorCode = "forbidden"
    case errors.Is(err, ErrConflict):
        status = http.StatusConflict
        errorCode = "conflict"
    default:
        status = http.StatusInternalServerError
        errorCode = "internal_error"
        slog.Error("unexpected error", "error", err, "request_id", requestID)
    }

    w.WriteHeader(status)
    json.NewEncoder(w).Encode(ErrorResponse{
        Error:     errorCode,
        Message:   err.Error(),
        RequestID: requestID,
    })
}
```

---

## Service ↔ Repository Contract

### Request Flow

1. Service receives validated request from handler
2. Service applies business logic (e.g., check subscription limits)
3. Service calls repository methods with domain models
4. Repository executes SQL, returns domain models or wrapped errors
5. Service propagates errors or transforms repository results

### Example Contract (Trip Retrieval)

**Repository Interface**:
```go
type TripRepository interface {
    FindByID(ctx context.Context, tripID string) (*Trip, error)
}
```

**Service Implementation**:
```go
func (s *TripService) GetTrip(ctx context.Context, tripID, userID string) (*Trip, error) {
    trip, err := s.repo.FindByID(ctx, tripID)
    if err != nil {
        return nil, err // Repository error already wrapped
    }

    // Business logic: Check authorization
    if trip.OwnerID != userID {
        // Check if user is a collaborator
        if !s.isCollaborator(ctx, tripID, userID) {
            return nil, ErrForbidden
        }
    }

    return trip, nil
}
```

**Repository Implementation**:
```go
func (r *TripRepository) FindByID(ctx context.Context, tripID string) (*Trip, error) {
    var trip Trip
    err := r.db.QueryRow(ctx, `
        SELECT id, title, start_date, end_date, owner_id, created_at, updated_at
        FROM trips
        WHERE id = $1
    `, tripID).Scan(&trip.ID, &trip.Title, &trip.StartDate, &trip.EndDate, &trip.OwnerID, &trip.CreatedAt, &trip.UpdatedAt)

    if errors.Is(err, pgx.ErrNoRows) {
        return nil, fmt.Errorf("trip not found: %w", ErrNotFound)
    }
    if err != nil {
        return nil, fmt.Errorf("failed to query trip: %w", err)
    }

    return &trip, nil
}
```

**Error Wrapping Rules**:
- Repository MUST wrap all database errors with context: `fmt.Errorf("failed to <operation>: %w", err)`
- Service MAY transform repository errors into domain errors (e.g., `pgx.ErrNoRows` → `ErrNotFound`)
- Service MUST NOT log repository errors (logging happens in middleware for 5xx errors)

---

## Service ↔ AI Client Contract

### Request Flow

1. Service receives itinerary generation request from handler
2. Service calls PromptValidator to check for injection attacks
3. Service calls AIClient with validated prompt
4. AIClient streams response chunks or returns full itinerary
5. Service calls OutputSanitizer on AI response before persisting
6. Service returns sanitized itinerary to handler

### Example Contract (Itinerary Generation)

**AI Client Interface**:
```go
type AIClient interface {
    GenerateItinerary(ctx context.Context, prompt string) (*ItineraryResponse, error)
    StreamItinerary(ctx context.Context, prompt string) (<-chan ItineraryChunk, <-chan error)
}

type ItineraryResponse struct {
    Destinations []Destination `json:"destinations"`
    RawContent   string         `json:"raw_content"` // Unsanitized AI output
}
```

**Service Implementation**:
```go
func (s *ItineraryService) GenerateItinerary(ctx context.Context, req GenerateRequest) (*Itinerary, error) {
    // Validate prompt for injection attacks
    if ruleID, matched := s.validator.Validate(req.Prompt); matched {
        slog.Warn("prompt injection detected", "rule_id", ruleID, "request_id", requestID)
        return nil, ErrInvalidPrompt
    }

    // Call AI provider
    response, err := s.aiClient.GenerateItinerary(ctx, req.Prompt)
    if err != nil {
        return nil, err // AI client errors already wrapped
    }

    // Sanitize output before persisting
    sanitized := s.sanitizer.Sanitize(response.RawContent)

    // Persist to database
    itinerary := &Itinerary{
        TripID:  req.TripID,
        Content: sanitized,
    }
    if err := s.repo.Create(ctx, itinerary); err != nil {
        return nil, err
    }

    return itinerary, nil
}
```

**AI Client Error Handling**:
- 400 (invalid prompt): Return `ErrValidation` with AI provider message
- 429 (rate limit): Return `ErrRateLimited` with `Retry-After` duration
- 503 (AI provider unavailable): Return `ErrServiceUnavailable`
- Network errors: Retry with exponential backoff (max 3 attempts), then return `ErrServiceUnavailable`

---

## Middleware ↔ Handler Contract

### Context Propagation

Middleware injects values into `context.Context`; handlers extract values using context keys.

**Request ID Middleware**:
```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Handler Extraction**:
```go
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}
```

**Authenticated User Middleware**:
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID, err := extractUserFromJWT(r)
        if err != nil {
            errors.HandleError(w, r, ErrUnauthorized)
            return
        }
        ctx := context.WithValue(r.Context(), UserIDKey, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Testing Contracts

### Service Layer Testing

**Unit Test** (service logic only, repository mocked):
```go
func TestTripService_CreateTrip(t *testing.T) {
    mockRepo := &MockTripRepository{
        CreateFunc: func(ctx context.Context, trip *Trip) error {
            return nil
        },
    }

    service := NewTripService(mockRepo, nil)

    req := CreateTripRequest{
        Title:     "Paris Adventure",
        StartDate: time.Now(),
        EndDate:   time.Now().Add(7 * 24 * time.Hour),
        UserID:    "user-123",
    }

    trip, err := service.CreateTrip(context.Background(), req)
    assert.NoError(t, err)
    assert.Equal(t, "Paris Adventure", trip.Title)
}
```

### Repository Layer Testing

**Integration Test** (real database, transactions rolled back):
```go
func TestTripRepository_FindByID(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()

    repo := NewTripRepository(db)

    // Insert test data
    tripID := "trip-123"
    _, err := db.Exec(context.Background(), `
        INSERT INTO trips (id, title, owner_id)
        VALUES ($1, $2, $3)
    `, tripID, "Test Trip", "user-123")
    require.NoError(t, err)

    // Test retrieval
    trip, err := repo.FindByID(context.Background(), tripID)
    assert.NoError(t, err)
    assert.Equal(t, "Test Trip", trip.Title)
}
```

---

## Summary

Backend layer contracts enforce:
- **Separation of concerns**: Handler (HTTP), Service (business logic), Repository (data access)
- **Testability**: Services have no HTTP dependencies, repositories have no business logic
- **Error propagation**: Repositories wrap errors, services transform errors, handlers map to HTTP status
- **Interface abstraction**: AI client behind interface for mockability
- **Context propagation**: Middleware injects values (requestID, userID) via context

All contracts align with constitution principles: simplicity (KISS/DRY), quality (testability), secure configuration (no HTTP coupling in business logic).
