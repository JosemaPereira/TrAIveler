# Research: Product Vision and Scope

**Date**: 2026-07-02 | **Plan**: [plan.md](plan.md)

All `NEEDS CLARIFICATION` items from the Technical Context are resolved below.

---

## Decision 1 — AI Provider

**Decision**: Anthropic Claude (`github.com/anthropics/anthropic-sdk-go`)

**Rationale**:
- Go 1.24+ is its explicit version target; SDK is actively maintained with daily releases (v1.56.0
  at time of research).
- First-class multi-turn support via `BetaMessageUtil` and `NextStreaming()` — aligns directly with
  FR-012 (conversational, multi-turn generation flow).
- Structured output via tool-use (function calling) enables typed itinerary output (days,
  activities, transfers) without fragile text parsing.
- Streaming responses via `NewStreaming()` allow the frontend to show progressive generation for
  AI calls that take 30–90 s.
- All methods accept `context.Context` with proper cancellation — idiomatic Go.

**Alternatives considered**:
- OpenAI (`github.com/openai/openai-go/v3`): Near-identical capability (v3.41.0, same streaming and
  tool-use support). Either is valid; Claude is preferred for its Go 1.24+ native targeting and
  daily maintenance cadence. Switch to OpenAI requires only provider wiring changes, not domain
  logic changes.
- Google Gemini: The prior `cloud.google.com/go/vertexai/genai` SDK was sunset June 24, 2026;
  the replacement `google.golang.org/genai` has minimal Go adoption. Excluded for stability risk.

**Usage pattern**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
defer cancel()

stream := client.Beta.Messages.NewStreaming(ctx, anthropic.BetaMessageNewParams{
    MaxTokens: 4096,
    Model:     anthropic.ModelClaudeOpus4_5,
    Messages:  conversationHistory, // multi-turn
})
defer stream.Close()
```

---

## Decision 2 — Go HTTP Router

**Decision**: `github.com/go-chi/chi/v5`

**Rationale**:
- Built on `http.Handler` / `http.ResponseWriter` — no custom context types. Handlers are directly
  testable with `net/http/httptest` without any framework adapter, satisfying Constitution III.
- Lightweight middleware composition via `router.Use()` (auth, logging, CORS, request-ID).
- Actively maintained by the same team as Goose (selected for migrations) — consistent ecosystem.
- No reflection or code generation — aligns with KISS (Constitution II).

**Alternatives considered**:
- Gin: Popular and fast but wraps `net/http` in a custom `*gin.Context`, increasing testability
  friction and introducing framework lock-in inconsistent with idiomatic Go.
- stdlib `net/http` alone: Viable but adds significant boilerplate for routing and middleware
  at MVP scale; Chi adds negligible overhead with clear gains.

---

## Decision 3 — PostgreSQL Driver

**Decision**: `github.com/jackc/pgx/v5` with `pgxpool` for connection pooling

**Rationale**:
- Context-first API (`ctx context.Context` on every operation) — idiomatic Go 1.24, supports
  graceful shutdown and request-scoped cancellation.
- Built-in connection pooling via `pgxpool.New(ctx, connStr)` — no additional wrapper needed.
- Parameterized queries by default — eliminates SQL injection risk (OWASP A3).
- Better performance than `lib/pq` (which is in maintenance mode) for this Go version.
- No ORM — explicit SQL keeps queries readable, debuggable, and consistent with KISS.

**Alternatives considered**:
- `jmoiron/sqlx` + `lib/pq`: `lib/pq` is no longer actively developed. `sqlx` adds an abstraction
  layer that buys little at MVP scale while hiding query semantics.
- `gorm`: Full ORM; introduces speculative abstraction (YAGNI) and makes query debugging harder.
  Excluded per Constitution II.

---

## Decision 4 — Database Migrations

**Decision**: `github.com/pressly/goose/v3`

**Rationale**:
- SQL-first: migrations are plain `.sql` files committed to `backend/migrations/` — readable by
  any developer without Go knowledge.
- CLI tool (`goose ... up/down/status`) is CI-friendly and integrates cleanly with Docker entrypoint.
- Programmatic Go API (`goose.Up(db, dir)`) allows test setups to spin up a fresh schema via a
  single call — critical for integration tests (Constitution I).
- Maintained by the same team as Chi — consistent tooling culture.

**Alternatives considered**:
- `golang-migrate/migrate`: Also solid; slightly heavier dependency graph and more configuration
  overhead. Marginal preference for Goose at MVP scale.

---

## Decision 5 — Authentication

**Decision**: `github.com/golang-jwt/jwt/v5` with JWT stored in HTTP-only cookies

**Rationale**:
- HTTP-only cookies prevent JavaScript access, blocking XSS-based token theft (OWASP A7).
- `SameSite: Lax` mitigates CSRF for same-site form submissions.
- `Secure: true` ensures cookies are only transmitted over HTTPS.
- Stateless — no server-side session store required at MVP scale.
- v5 is the current actively maintained major version.

**Cookie configuration**:
```go
http.SetCookie(w, &http.Cookie{
    Name:     "auth_token",
    Value:    signedToken,
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteLaxMode,
    Path:     "/",
    MaxAge:   3600, // 1 hour
})
```

---

## Decision 6 — Mock Payment Stub Interface

**Decision**: Go `interface` in `internal/subscription/payment/provider.go`; stub implementation
in `stub.go`; real implementation added post-MVP by satisfying the same interface.

**Rationale**: FR-013 mandates that the stub satisfy the same interface as a real payment provider
so post-MVP integration requires no domain logic changes. The Go interface pattern is the idiomatic
mechanism for this contract.

**Interface contract**:
```go
// PaymentProvider is the single contract all payment implementations must satisfy.
// The StubProvider satisfies it for the POC; a real Stripe/etc. adapter satisfies
// it post-MVP without touching subscription domain logic.
type PaymentProvider interface {
    // CreateSubscription initiates a subscription and returns provider-assigned IDs.
    CreateSubscription(ctx context.Context, req CreateSubscriptionRequest) (*SubscriptionResult, error)
    // CancelSubscription halts recurring charges.
    CancelSubscription(ctx context.Context, subscriptionID string) error
    // GetSubscription retrieves current status from the provider.
    GetSubscription(ctx context.Context, subscriptionID string) (*SubscriptionStatus, error)
}
```

**Stub behaviour**: always returns `status: "succeeded"`, logs to stdout with `[STUB]` prefix,
generates a deterministic `stub-tx-<unix-epoch>` transaction ID.

---

## Decision 7 — Frontend State Management

**Decision**: TanStack Query v5 (server state) + Zustand (client-only state)

**Rationale**:
- TanStack Query manages all API data: caching, background refetch, optimistic updates, loading and
  error states — satisfying Constitution IV (every data-dependent component MUST handle Loading,
  Error, and Empty states) without boilerplate.
- Zustand handles lightweight client state (current user session, conversation panel open/closed)
  where server sync is not needed.
- No Redux — YAGNI at this scale (Constitution II).

---

## Decision 8 — Frontend Routing

**Decision**: React Router v7

**Rationale**:
- React 19 compatible; data-routing API (loaders/actions) aligns with TanStack Query for
  pre-fetching trip data on navigation.
- Well-known, minimal learning curve, no speculative alternatives needed at MVP scale.
