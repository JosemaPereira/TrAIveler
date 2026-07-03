# Research: Non-Functional Requirements and System Constraints

**Date**: 2026-07-02 | **Plan**: [plan.md](plan.md)

All `NEEDS CLARIFICATION` items from the Technical Context are resolved below.

---

## Decision 1 — Go Structured Logging Library

**Decision**: `log/slog` (Go standard library, available since Go 1.21)

**Rationale**:
- Aligns with the constitution's KISS principle: zero external dependencies; ships with Go 1.24.
- Natively produces structured JSON output via `slog.NewJSONHandler(os.Stdout, nil)`.
- Context-scoped log entries via `slog.With()` and `slog.WithGroup()` allow the request-ID
  middleware to inject the correlation ID, method, path, status code, and duration into every
  log line without thread-local state or logger passing.
- Chi middleware integration is idiomatic: create a `slog.Logger` per request using
  `logger.With("request_id", r.Header.Get("X-Request-ID"))` inside the HTTP middleware chain.
- Performance is on-par with structured third-party loggers for typical REST API throughput.
  No benchmark evidence that slog is a bottleneck at the 500 VU scale required by NFR-SCALE-001.
- Go's maintenance guarantee ensures slog evolves with the language. Migration to a third-party
  logger later requires only replacing the handler, not rewriting call sites.

**Alternatives considered**:
- **`github.com/rs/zerolog`**: Lightweight and fast, but adds an external dependency for features
  slog already provides. Justified only if serialization benchmarks show a measured bottleneck.
- **`go.uber.org/zap`**: Battle-tested at high scale; adds API complexity and external dependency.
  Revisit post-MVP if load-test profiling identifies logging as a bottleneck.

---

## Decision 2 — AI Output Sanitisation

**Decision**: `github.com/microcosm-cc/bluemonday` (server-side, before storage) +
React default JSX escaping (frontend, at render time)

**Rationale**:
- NFR-SEC-008 requires sanitisation *before storage and rendering*. Bluemonday strips HTML tags,
  script blocks, and dangerous event-handler attributes from untrusted LLM output at write-time.
- Sanitising at write-time (before the DB INSERT) means every downstream consumer — REST API,
  future PDF export, email notifications, mobile clients — receives already-safe content without
  needing to re-sanitise.
- React's default JSX escaping provides defence-in-depth at the rendering layer with no extra code.
- `html.EscapeString()` (stdlib) is insufficient: it escapes text content but does not strip
  `<script>` tags or `onerror=` attributes embedded in HTML fragments returned by the LLM.
- Bluemonday is actively maintained, widely adopted in Go, and requires a five-line integration:
  create a strict `UGCPolicy`, call `p.Sanitize(aiContent)` before storage.

**Alternatives considered**:
- **`html.EscapeString()` only**: Insufficient for HTML fragments; does not handle attribute-level
  injection vectors.
- **No server-side sanitisation**: Violates NFR-SEC-008 explicitly; leaves non-React consumers
  (exports, future mobile clients) vulnerable.

---

## Decision 3 — AI Prompt Injection Detection Approach

**Decision**: Pattern-based server-side deny-list validated before forwarding to the AI provider.
Rules defined in a versioned YAML config file (`backend/config/prompt-rules.yml`).

**Rationale**:
- OWASP LLM Top 10 (LLM01 — Prompt Injection) recommends input validation as the primary
  defence layer. A deny-list of known injection patterns (instruction-override phrases, role-
  switching directives, system-prompt extraction attempts) is the simplest correct solution per
  the constitution's KISS principle.
- The validation runs synchronously in the request path before any AI call is made, so rejected
  prompts incur zero AI provider cost and near-zero latency.
- Patterns are stored in a versioned config file (not hard-coded) so they can be updated without
  redeploying the binary — the service reloads the config on startup.
- Examples of covered pattern classes: `ignore previous instructions`, `repeat your system prompt`,
  `you are now [role]`, `disregard all prior context`, excessively long repeated tokens (prompt
  flooding), non-travel-domain directive attempts detected via keyword classification.
- A scoped system prompt (server-side, never exposed to the client) further constrains Claude's
  response domain to travel planning.

**Alternatives considered**:
- **LLM-as-classifier** (a second AI call to classify whether the input is malicious): High latency
  overhead (~1–3 s), additional API cost, and complexity. Not suitable for MVP.
- **No validation layer**: Violates NFR-SEC-007 and exposes user data to exfiltration via prompt
  manipulation. Explicitly rejected.

---

## Decision 4 — CI/CD Platform and Quality Gate Pipeline

**Decision**: GitHub Actions (`.github/workflows/`) — three workflows: `ci.yml` (always-on PR
gate), `accessibility.yml` (always-on PR gate), `load-test.yml` (manual trigger only).

**Rationale**:
- GitHub Actions is already the implied platform from the constitution's `.github/workflows/`
  reference and the team's `main`-branch protection model.
- Three separate workflows allow independent failure attribution: a lint failure does not obscure
  an accessibility failure; the load test can be triggered manually without blocking PR merges.
- `load-test.yml` is manual-trigger (`workflow_dispatch`) because running a 500 VU test on every
  PR would exhaust free GitHub Actions minutes and slow feedback loops. The gate is enforced
  before each release milestone.

**Alternatives considered**:
- **Single monolithic CI workflow**: Conflates unrelated failures; slower to iterate on.
- **CircleCI / GitLab CI**: No evidence in the repo that either is adopted. Switching costs not
  justified.

---

## Decision 5 — Accessibility Automation in CI

**Decision**: `@axe-core/playwright` for automated WCAG 2.1 AA scanning integrated into the
existing Playwright E2E suite; `@lhci/cli` (Lighthouse CI) for Core Web Vitals assertions in CI.

**Rationale**:
- Playwright is already the E2E framework per the constitution. Adding `@axe-core/playwright`
  requires only two lines per test (`await checkA11y(page)`) — no new test runner or process.
- `@lhci/cli` runs Lighthouse audits headlessly against the built application and can fail CI if
  the accessibility score drops below 90 (NFR-A11Y-004) or if LCP/CLS/INP exceed thresholds
  (NFR-PERF-003). It integrates as a GitHub Actions step with the official LHCI Action.
- Running both tools in the same `accessibility.yml` workflow keeps accessibility checks isolated
  from the functional CI pipeline.

**Alternatives considered**:
- **Pa11y**: Standalone accessibility tool; adds a separate process alongside Playwright.
  Redundant given @axe-core/playwright already provides Playwright-native scanning.
- **Manual-only accessibility audits**: Insufficient; violations introduced by new components
  would go undetected until a manual review cycle.
