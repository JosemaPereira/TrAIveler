# Session Notes

Historical summaries of completed development sessions. Committed to git as a record.

## Template

### Session: <name>
- **Date**: <YYYY-MM-DD>
- **What was accomplished**: <features built/fixed>
- **Key findings and decisions**: <important learnings, trade-offs>
- **Outcomes**: <what now works in the application>

---

## Example

### Session: Project Bootstrap
- **Date**: 2026-01-01
- **What was accomplished**: Initialized repo, context docs, and SpecKit.
- **Key findings and decisions**: Chose spec-driven flow; English-only artifacts.
- **Outcomes**: Project scaffold ready; first feature spec pending.

---

### Session: Product Vision, Planning, and Spec Artifacts
- **Date**: 2026-07-02
- **Branch**: `feature/001-product-vision-scope` → PR #4
- **What was accomplished**:
  - Created `specs/001-product-vision-scope/spec.md` — full product vision with 4 personas, 4 user stories, MVP scope, out-of-scope list, 7 success criteria, and a Clarifications section.
  - Ran two full `/speckit.clarify` sessions (6 questions total) resolving: subscription model, admin/partner roles, suggest-then-approve collaboration workflow, suggestion history preservation, conversational multi-turn generation (free-form input), mock payment stub strategy, stub checkout UX (visible, always-succeeds), and MVP scale target (< 50 concurrent users).
  - Created `specs/001-product-vision-scope/plan.md` — implementation plan, constitution check (5/5 pass), project structure (Option 2: backend/ + frontend/ + e2e/).
  - Created `specs/001-product-vision-scope/research.md` — 8 technology decisions: Anthropic Claude (AI), Chi (HTTP router), pgx/v5 (PostgreSQL), Goose (migrations), golang-jwt/jwt v5 (auth, HTTP-only cookies), Go PaymentProvider interface pattern (mock stub), TanStack Query v5 + Zustand (frontend state), React Router v7.
  - Created `specs/001-product-vision-scope/data-model.md` — PostgreSQL schema for 11 entities with indexes and edge case handling.
  - Created `specs/001-product-vision-scope/contracts/api.md` — REST API contract for all domains.
  - Created `specs/001-product-vision-scope/quickstart.md` — 7 end-to-end validation scenarios.
  - Created `specs/001-product-vision-scope/tasks.md` — 81 dependency-ordered tasks across 7 phases (38 parallelisable).
  - Added `.github/agents/technical-writer.agent.md` and `.github/prompts/technical-writer.prompt.md`.
  - Rewrote `README.md` with product identity, MVP scope, tech stack, planned structure, and doc index.
- **Key findings and decisions**:
  - Platform is subscription-based (future paid); for POC the payment checkout is a **visible stub** that always succeeds. The Go `PaymentProvider` interface pattern ensures the stub is swappable post-MVP without touching domain logic.
  - Collaboration model: **suggest-then-approve** only — partners submit suggestions; admin is the sole approver. Suggestions are never hard-deleted (FR-011).
  - AI generation flow is **conversational and multi-turn** — the system asks follow-up questions before producing the itinerary. Free-form natural language input supports mixed-interest groups.
  - Anthropic Claude chosen over OpenAI (same capability, Go 1.24+ native targeting, daily maintenance) and over Google Gemini (prior SDK sunset June 2026).
  - MVP scale: < 50 concurrent users, single instance, best-effort uptime — no formal SLA.
- **Outcomes**: All planning artifacts committed and pushed. PR #4 open against `main`. Implementation can begin with Phase 1 (T001–T008) immediately.

---

### Session: NFR System Constraints, Roadmap, and Documentation
- **Date**: 2026-07-02
- **Branch**: `feature/002-nfr-system-constraints` → PR #5
- **What was accomplished**:
  - Created `specs/002-nfr-system-constraints/spec.md` — 26 measurable NFRs across 6 quality attributes (performance, scalability, availability, accessibility, security, maintainability).
  - Created `specs/002-nfr-system-constraints/plan.md` — NFR implementation plan with observability primitives (RequestID, Logger, /healthz), PromptValidator, OutputSanitizer, and CI security gates.
  - Created `specs/002-nfr-system-constraints/research.md` — tool decisions: log/slog (Go stdlib), bluemonday (sanitization), k6 (load testing), gitleaks (secret scanning), axe-core (a11y), gosec (SAST), Lighthouse CI.
  - Created `specs/002-nfr-system-constraints/data-model.md` — operational schemas: StructuredLogEntry, HealthCheckResponse, PromptValidationRule.
  - Created `specs/002-nfr-system-constraints/contracts/api.md` — NFR API contracts: GET /healthz, X-Request-ID header convention, 400 PromptRejectionResponse.
  - Created `specs/002-nfr-system-constraints/quickstart.md` — 12 validation scenarios covering all NFR categories.
  - Created `specs/002-nfr-system-constraints/tasks.md` — 46 dependency-ordered NFR tasks across 8 phases.
  - Built consolidated project roadmap at `docs/roadmap.md` — 127 tasks reconciled from both specs with Group and Sprint columns.
  - Applied 29 group assignments bundling 75 tasks (59%) into cohesive work items; reduces GitHub backlog from 127 to 81 issues.
  - Rewrote `README.md` from technical reference to presentation letter — user-centric narrative, eliminated 6 large tables, casual language with emojis, 3 key links for deep dive.
- **Key findings and decisions**:
  - **Task grouping strategy**: Grouped related small tasks by shared context (e.g. G-SETUP-INIT, G-BACKEND-MIDDLEWARE, G-US1-COMPONENTS) to reduce issue fragmentation. 29 groups form issues with checklists; 52 standalone tasks remain. Total: 81 issues instead of 127.
  - **Observability core as NFR foundation**: RequestID → Logger → /healthz middleware chain must be implemented before any load tests, security scans, or user story work can be validated. This is Phase 2 of spec 002 and gates all subsequent NFR tasks.
  - **Prompt validation is a CI blocker**: PromptValidator (002-T025–T026) and OutputSanitizer (002-T027–T028) must be wired into the itinerary handler (002-T029–T030) before shipping any AI-powered feature. Pattern-based deny-list approach chosen over LLM-based validation to avoid AI cost per rejection and enforce deterministic blocking.
  - **README as pitch, not docs**: New README targets non-technical readers and potential contributors. Technical depth lives in specs/ and docs/ — README is now a 2-minute read that tells the story, not the architecture.
- **Outcomes**: All NFR planning artifacts committed. Roadmap file with 29 groups ready for `/sync-issues` execution. README suitable for GitHub profile or project showcases. Memory system validated and documented.

---

### Session: Roadmap Grouping and README Revision
- **Date**: 2026-07-02
- **Branch**: `feature/nfr-system-constraints` → PR #5
- **What was accomplished**:
  - Created `specs/002-nfr-system-constraints/spec.md` — 26 measurable NFRs across 6 quality attributes (performance, scalability, availability/reliability, accessibility, security, maintainability) plus 3 data-privacy NFRs (GDPR-aware design). Each NFR has a concrete measurable target and a named validation method.
  - Ran `/speckit.clarify` (5 questions): GDPR posture → GDPR-aware MVP (right-to-deletion, minimal PII, privacy policy, no formal DPO); RPO → ≤ 24 h daily backup; AI rate-limit strategy → fast-fail 503 + added AI input security NFRs (prompt injection, output sanitisation); 500 VU basis → conservative POC engineering baseline; i18n → English-only, no i18n infrastructure at MVP.
  - Created `specs/002-nfr-system-constraints/plan.md` — implementation plan for NFR infrastructure: slog middleware, `/healthz`, `PromptValidator`, `OutputSanitizer`, privacy policy page, GitHub Actions CI gates.
  - Created `specs/002-nfr-system-constraints/research.md` — 5 decisions: `log/slog` (stdlib, zero deps) over zap/zerolog; `bluemonday` (server-side before storage + React JSX escaping); pattern-based deny-list for prompt injection (zero AI cost per rejection, OWASP LLM Top 10); GitHub Actions (3 separate workflows for isolated failure attribution); `@axe-core/playwright` + `@lhci/cli` (Playwright-native, no extra test runner).
  - Created `specs/002-nfr-system-constraints/data-model.md` — 4 operational schemas: `StructuredLogEntry`, `HealthCheckResponse`, `PromptValidationRule`, `PromptRejectionResponse`.
  - Created `specs/002-nfr-system-constraints/contracts/api.md` — `GET /healthz`, `X-Request-ID` convention, prompt rejection (400), AI rate-limit fast-fail (503), accessibility CI contract.
  - Created `specs/002-nfr-system-constraints/quickstart.md` — 12 runnable validation scenarios and a CI gate reference table.
  - Created `specs/002-nfr-system-constraints/tasks.md` — 46 tasks across 8 phases. Phase 2 (observability core) is foundational; US3 security is the minimum viable NFR baseline. 7 tasks extend 001 CI workflows (T076–T079) rather than duplicating them.
  - Created `docs/roadmap.md` — consolidated 127-task project roadmap with stable global IDs (`001-T001` … `002-T046`), dependency-ordered tables, priority/status/parallel columns, and a critical-path diagram.
  - Updated `README.md` — added spec 002 project status rows, NFR summary section, expanded Key Design Decisions table, updated repository structure, extended documentation index.
  - Created `backend/README.md`, `frontend/README.md`, `e2e/README.md` — area-specific orientation READMEs (tech stack, annotated directory trees, setup/test/lint commands, env vars, spec cross-references).
- **Key findings and decisions**:
  - **AI security boundary**: user input sent to the AI must be validated server-side (prompt injection deny-list) and AI output must be sanitised before storage (bluemonday). This is a non-negotiable P1 security requirement; the system prompt is treated as a secret and never exposed client-side.
  - **NFR ordering for implementation**: US4 (observability middleware) must be built first despite being P2 in the spec, because it is a prerequisite for validating all other NFRs. Phase 2 in tasks.md reflects this.
  - **Cross-spec coordination**: 002 CI extension tasks (T034–T036, T043–T044) modify files created by 001 tasks (T076–T079). Must coordinate or sequence these in sprint planning.
  - **GDPR-aware design**: chosen over full GDPR compliance program for POC. Minimum required: right-to-deletion endpoint (`DELETE /users/me`), data minimisation, privacy policy page linked from registration.
  - **Roadmap is idempotent**: `/build-roadmap` can be re-run whenever a new spec's `tasks.md` is added or modified; it adds/updates/archives without destroying human-owned fields (Priority, Status, Issue).
- **Outcomes**: All spec 002 + roadmap + docs artifacts committed and pushed. PR #5 open against `main`. Implementation sequence: Phase 1 (T001–T005) → Phase 2 (T006–T012: observability core) → P1 stories (Performance, Accessibility, Security) in parallel.
