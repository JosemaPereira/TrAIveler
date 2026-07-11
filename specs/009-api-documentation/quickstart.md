# Quickstart: Validating API Documentation via OpenAPI/Swagger

Prerequisites: local backend dev environment already set up (`backend/README.md`), `swag` CLI
installed (`go install github.com/swaggo/swag/cmd/swag@latest`).

## Scenario 1 — Contract regenerates from code

```bash
cd backend
make swagger              # wraps `swag init -g cmd/api/docs.go -o docs`
git status                # backend/docs/{docs.go,swagger.json,swagger.yaml} should be unchanged
                           # on a clean checkout, or show a diff after adding/changing an
                           # annotated handler
```

Expected: running `make swagger` twice in a row with no code changes produces zero diff
(idempotent generation) — validates FR-002.

## Scenario 2 — Interactive UI, real request

```bash
docker compose up -d          # or: go run ./cmd/api
open http://localhost:8080/swagger/index.html
```

1. Expand the `examples` tag → `POST /api/v1/examples`.
2. Click "Try it out", fill the request body, click "Execute".
3. Confirm the displayed response status/body matches what `curl` would return for the same
   request.

Expected: real response observed in-browser, not a static example — validates FR-004/US2.

## Scenario 3 — Drift check catches an un-regenerated change

```bash
# Simulate a developer forgetting to regenerate:
# 1. Edit a swag-annotated handler's @Success response type.
# 2. Do NOT run `make swagger`.
git add -A && git commit -m "test: simulate drift"
make swagger
git diff --exit-code -- backend/docs   # should be NON-zero (diff present) → simulates CI failure
git checkout -- backend/docs           # revert simulation
```

Expected: `git diff --exit-code` exits non-zero, which is exactly the condition the new
`backend-ci.yml` step gates on — validates FR-008/US3.

## Scenario 4 — Excluded endpoint stays excluded

```bash
curl -s http://localhost:8080/swagger/doc.json | grep -c '"/healthz"'
```

Expected: `0` — `/healthz` has no `swag` annotations and never appears in the contract,
without any explicit exclusion list — validates FR-006.

## Scenario 5 — Auth-gated in every environment (forward-looking, until Sprint 5)

Not independently testable until Sprint 5's JWT middleware exists (see `research.md`'s "Auth
gating ahead of Sprint 5" decision). Once it lands, re-run this quickstart's Scenario 2 without
a bearer token and confirm a 401, matching every other `/api/v1` endpoint — validates FR-010.
