# Spec 005 Validation Results

**Date**: 2026-07-13
**Issue**: [#122](https://github.com/JosemaPereira/TrAIveler/issues/122) — G-ARCH-POLISH-VALIDATION
**Stable IDs**: 005-T124 – 005-T130
**Environment**: macOS (darwin/amd64), Go 1.26.4, Node 24.18.0, Terraform 1.15.8, Colima
(Docker runtime) for backend testcontainers.

This is the final validation sweep for Spec 005 (System Architecture and Technology Stack). Each
section below records the actual command output/numbers observed, not an assumed pass. Where the
current codebase has diverged from `quickstart.md`'s original (now-stale) skeleton snapshot, the
scenario's *intent* — the architectural outcome it was meant to prove — was validated against the
real, current `backend/`/`frontend/`/`infra/` trees instead of literally re-running the quickstart's
`mkdir`/file-listing steps.

---

## 005-T124: Quickstart Scenarios 1–4 (Scenario 5 covered separately under 005-T130)

### Scenario 1 — Backend Skeleton

Built the real backend (`go build ./...` — clean) and ran it directly (`go run ./cmd/api`) against
the existing local Postgres (already running via `docker-compose`, reused as-is; no new containers
created) on a scratch port (8081, to avoid colliding with a stale container already bound to 8080 —
see note below).

Observed startup log:
```
{"msg":"starting TrAIveler backend API","version":"0.1.0","environment":"development"}
{"msg":"configuration loaded successfully"}
{"msg":"database connection established","attempt":1,"min_conns":5,"max_conns":25}
{"msg":"HTTP server listening","port":8081}
```

`curl -i http://localhost:8081/healthz`:
```
HTTP/1.1 200 OK
Content-Type: application/json
X-Request-Id: c6afb436-3b79-4b50-b3da-c31b55e3f1fa
{"status":"ok","version":"0.1.0","uptime_seconds":9.007465471}
```
A second call with an incoming `X-Request-ID: test-correlation-123` header echoed the same value
back verbatim (correlation ID propagation confirmed both directions: generate-if-absent, echo-if-
present).

CORS preflight (`OPTIONS /healthz` with `Origin: http://localhost:5173`) returned `204` with
`Access-Control-Allow-Origin: http://localhost:5173` and the expected allowed headers/methods.

Sent `SIGTERM` to the running process; log confirmed graceful drain:
```
{"msg":"shutting down gracefully...","signal":"terminated"}
{"msg":"shutdown complete"}
{"msg":"database connection pool closed"}
```

**Result**: PASS. Server starts cleanly, DB pool established (min 5 / max 25 per
`internal/database/client.go`), `/healthz` matches the current `HealthCheckResponse` contract
(`status`/`version`/`uptime_seconds`, always `200` — reconciled in #110/PR #126, see
`patterns-discovered.md`), correlation ID and CORS middleware both work, graceful shutdown on
SIGTERM confirmed.

**Note (not a gap, informational)**: a leftover `traveler-api` Docker container from an earlier
manual session was still `docker ps`-listed as "healthy" on host port 8080, but the port was not
actually reachable from the host (`nc -z localhost 8080` refused), and the one response it did give
via `docker exec` showed the pre-#126 `{"status":"healthy","database":"connected"}` schema — i.e. a
stale image predating the current `/healthz` contract. Not touched (no container start/stop
performed against it) since it's unrelated to this validation-only issue's scope; whoever next
touches local Docker state should just recreate it (`docker-compose up -d --build backend`) to pick
up the current image.

### Scenario 2 — Frontend Scaffold

Started the real Vite dev server (`npm run dev`, port 5173) and loaded it in a headless Chromium
page (Playwright, via `e2e/node_modules`). Confirmed:
- Renders without error: `<h1>TrAIveler</h1>` (current `HomePage.tsx`, a documented placeholder —
  see its own doc comment — real content is future-sprint work), zero browser console errors, zero
  page errors.
- Design tokens (`src/styles/tokens.css`), Zustand store (`src/stores/auth-store.ts`), and
  `QueryClientProvider` (wired in `src/App.tsx`) all exist in the current tree and are exercised by
  the existing Vitest suite (see 005-T129/T126 below).

**Result**: PASS (against current codebase; the original quickstart's bespoke `ArchitectureTest`/
`Login`-button flow no longer exists as literal code — real auth pages are future-sprint work — so
the check was on the actual current scaffold: it renders, wires the same libraries, zero runtime
errors).

### Scenario 3 — Infrastructure Plan

Ran from `infra/`, mirroring exactly what `.github/workflows/infra-plan.yml`'s real (non-gated)
`validate`/`format-check` jobs do:
```
$ terraform init -backend=false   → "Terraform has been successfully initialized!"
$ terraform validate              → "Success! The configuration is valid."
$ terraform fmt -check -recursive → exit 0, no output (no diffs)
```
6 modules initialize cleanly (`vpc`, `rds`, `alb`, `cloudfront`, `secrets`, `ecs`), `aws` (v5.100.0)
and `random` (v3.9.0) providers resolve from the lockfile.

**`terraform plan` was deliberately not run against real AWS** — this machine does have live AWS
credentials configured (`aws sts get-caller-identity` succeeds), but running a real plan would make
live AWS API calls, which is inconsistent with this repo's standing constraint (carried since Sprint
3, still in force per `.github/memory/session-notes.md` and the personal-memory note
`no-aws-deploys-until-infra-refined`): no real `terraform apply`/`aws` calls until the user says
infra is ready. The CI workflow itself follows the same caution — its `plan` job is gated behind the
(currently absent) `AWS_ROLE_ARN` secret and has never run for real either. This is a deliberate,
consistent choice, not an oversight.

**Result**: PASS for the parts safe to run for real (`init -backend=false`, `validate`, `fmt
-check`); `plan` intentionally skipped per standing AWS-cost-avoidance policy.

### Scenario 4 — End-to-End Integration

The current codebase has no live `ArchitectureTest`-style component (per session notes: no route
makes a real API call yet — auth/trip pages are future-sprint work), so the quickstart's literal
"click a button, see the response" steps don't apply. Instead, validated the actual integration
mechanics that exist today, from inside a real browser page (so real CORS enforcement applies,
unlike a bare `curl`):

```js
await fetch('http://localhost:8081/healthz', { credentials: 'include', headers: {...} })
// → { ok: true, status: 200, body: { status: 'ok', version: '0.1.0', uptime_seconds: ... } }
```
No CORS error was thrown (the fetch resolved cleanly from the `http://localhost:5173` origin against
the backend's CORS-middleware-gated origin allowlist). Correlation ID propagation was already
confirmed at the HTTP layer in Scenario 1; `frontend/src/lib/api-client.ts` reads the ID from the
JSON error body's `request_id` field (not the response header) on error responses — by design, this
is what `internal/errors/handler.go`'s `writeErrorResponse` puts there, and matches the existing,
already-tested `useErrorHandler`/`ErrorMessage` pattern (#118/#129).

**Result**: PASS for the outcome this scenario protects — frontend↔backend reachability, CORS
configured correctly, correlation ID mechanism in place end-to-end — validated against real running
processes, not the stale bespoke test component the quickstart originally specified.

---

## 005-T125: Backend Lint (`golangci-lint`)

```
$ cd backend && golangci-lint run ./...
(no output, exit 0)
```
**Result**: PASS — zero errors/warnings. (`gofmt -l .` and `go vet ./...` also clean, run alongside
as a sanity check.)

---

## 005-T126: Frontend Lint + Format (ESLint + Prettier)

```
$ cd frontend && npm run lint
> eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0
(no output, exit 0)

$ npx prettier --check "src/**/*.{ts,tsx,css}"
Checking formatting...
All matched files use Prettier code style!
```
**Result**: PASS — zero ESLint errors/warnings, zero Prettier formatting diffs.

---

## 005-T127: Terraform Format Check

```
$ cd infra && terraform fmt -check -recursive
(exit 0, no output)
```
**Result**: PASS — zero formatting diffs across all `infra/` modules and environments.

---

## 005-T128: Backend Coverage (`internal/example` — service + repository)

Two numbers are relevant here, and the gap between them is expected and already documented in this
repo's own testing strategy (`patterns-discovered.md`, "Layered Testing Strategy with Optional
Testcontainers") — flagging it explicitly rather than reporting only the flattering number:

- **`make test-coverage`** (the actual CI target — `-short`, no testcontainers, no Docker
  dependency): `internal/example` package coverage = **64.1%**. This is expected to look low:
  `repository.go`'s DB-backed methods (`Create`/`FindByID`/`FindByEmail`/`Update`/`Delete`/`List`)
  are only exercised by `repository_integration_test.go`, which is testcontainer-gated and
  intentionally skipped in short/CI mode (`-short` flag) — by design, not a regression.
- **Full suite with Colima-backed testcontainers** (`go test -tags=test ./... -coverprofile=...`,
  no `-short`, `DOCKER_HOST`/`TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE` set per the existing Colima
  pattern): `internal/example` package coverage = **88.0%**. Isolating just `service.go` +
  `repository.go` (the two files 005-T128/the issue actually asks about) from the raw coverage
  profile: **101 statements total, 90 covered = 89.1%**.

**Result**: PASS against the ≥80% target in `docs/testing-guidelines.md` — using the full,
Colima-backed test run (the only run that actually exercises `repository.go`), `internal/example`'s
service + repository combined coverage is **89.1%**, comfortably over the floor. The `-short`/CI
number alone would look like a shortfall, but that's a known, deliberate consequence of skipping
testcontainer tests in CI's fast unit-only lane, not a real gap — no code or test changes made here.

---

## 005-T129: Frontend Primitive Tests (Button, Card, Input)

```
$ cd frontend && npm test -- Button Card Input
 Test Files  3 passed (3)
      Tests  27 passed (27)
```
Also ran the full frontend suite as a sanity check (not required by this task, but cheap and
informative): **136/136 tests passed across 15 files**, zero failures.

**Result**: PASS — all three primitives render correctly with design tokens (existing `.test.tsx`
suites, no changes needed).

---

## 005-T130: Accessibility (axe-core, ad-hoc)

**Important scope note**: `.github/workflows/accessibility.yml`'s "Run Playwright accessibility
specs" step (`npx playwright test --grep @accessibility --pass-with-no-tests`) is a deliberately
deferred, already-tracked no-op — its own inline comment states no spec is tagged `@accessibility`
yet, pending task **002-T023 (Sprint 9)**. This is not a new finding and is not re-filed here; see
`.github/memory/session-notes.md` (Sprint 3 closure) and the personal-memory note
`sprint3-lighthouse-ci-gap`. Per the issue's own explicit instruction, no new permanent
`accessibility.spec.ts` was authored — that would misrepresent a Sprint-9-owned decision as this
issue's own gap.

Instead, ran a real, disposable, one-off axe-core scan (not committed — a throwaway script using
`@axe-core/playwright`'s `AxeBuilder`, already a dependency in both `frontend/package.json` and
`e2e/node_modules`) against the current frontend dev server's `/` route (`HomePage.tsx`, currently
just `<h1>TrAIveler</h1>` — a documented placeholder, real content lands in future sprints):

```
axe-core tags scanned: ['wcag2a', 'wcag2aa']
violations: 0
passes: 6
```

**Result**: PASS — **0 real WCAG 2.1 AA violations** observed, as expected given the near-absence of
markup today. This is an honest point-in-time measurement of the current placeholder page, not a
substitute for the real Playwright-based CI gate — that gate remains in its documented
`--pass-with-no-tests` vacuous-pass state until 002-T023 tags a real `@accessibility` spec in
Sprint 9. Re-run this same ad-hoc check (or, preferably, land 002-T023) once real page content
exists — a 0-violation result on one placeholder `<h1>` says nothing about pages with real forms,
navigation, or images.

---

## Summary

| Task | Check | Result |
|---|---|---|
| 005-T124 | Quickstart Scenarios 1–4 (adjusted for current codebase) | PASS |
| 005-T125 | `golangci-lint run ./...` | PASS — 0 errors |
| 005-T126 | ESLint + Prettier | PASS — 0 errors, 0 diffs |
| 005-T127 | `terraform fmt -check -recursive` | PASS — 0 diffs |
| 005-T128 | Backend coverage, `internal/example` (service+repository) | PASS — 89.1% (full suite); CI's `-short` lane alone reads 64.1% by design (repository needs testcontainers) |
| 005-T129 | Frontend primitive tests (Button/Card/Input) | PASS — 27/27 (136/136 full suite) |
| 005-T130 | Ad-hoc axe-core scan | PASS — 0 violations (placeholder page); real CI gate still deferred to 002-T023/Sprint 9, as already tracked |

**No functional bugs or coverage shortfalls requiring new application code were found.** No follow-up
issues filed as a result of this validation sweep — the two items worth a reader's attention (the
short-mode coverage lane's expected gap on `repository.go`, and the pre-existing/known
002-T023 accessibility-CI deferral) are both already-documented, already-tracked, non-blocking
characteristics of this repo's existing testing strategy, not new gaps discovered by this issue.
