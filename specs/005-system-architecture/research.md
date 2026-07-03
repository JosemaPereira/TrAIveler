# Research: System Architecture and Technology Stack

**Feature**: System Architecture and Technology Stack  
**Created**: 2026-07-03  
**Status**: Complete

## Overview

This document consolidates technology decisions for the backend, frontend, and infrastructure layers. Many choices are mandated by the constitution (spec 001-003 decisions). Clarifications from the planning session resolved ambiguities around cost optimization, environment separation, and production validation strategy.

---

## Decision 1: Frontend HTTP Client

**Chosen**: Native fetch API with thin wrapper

**Rationale**:
- Modern browsers have excellent fetch support; no external HTTP library needed
- React 19 ecosystem integrates seamlessly with fetch
- TanStack Query works natively with fetch-based functions
- Eliminates Axios dependency (~15 KB), reducing bundle size (target: < 500 KB gzipped per SC-006)
- Thin wrapper provides base URL, default headers, and response/error normalization without library overhead

**Alternatives Considered**:
- Axios: Full-featured HTTP library with interceptors and automatic JSON parsing, but adds dependency and bundle size
- TanStack Query built-in fetch: No abstraction layer, but makes header/error handling repetitive across query functions
- Hybrid (fetch + Axios for complex cases): Unnecessary complexity for MVP; fetch handles file uploads and request cancellation via AbortController

**Implementation**:
- Create `src/lib/api-client.ts` with fetch wrapper
- Inject base URL, auth tokens (from cookies), correlation ID header
- Normalize responses (2xx → data, 4xx/5xx → throw structured error)
- Export typed fetch functions for use in TanStack Query hooks

**Source**: Clarification #3

---

## Decision 2: Terraform Environment Separation Strategy

**Chosen**: Separate .tfvars files (staging.tfvars, production.tfvars) with single default workspace

**Rationale**:
- Environment differences explicit and reviewable in version control
- Simpler for teams new to Terraform; avoids workspace-switching errors
- Single command: `terraform apply -var-file=environments/staging.tfvars`
- All environment-specific values (instance types, retention periods, CIDR blocks) visible in .tfvars diffs
- Terraform workspaces better suited for ephemeral environments (feature branches, testing), not long-lived staging/production

**Alternatives Considered**:
- Terraform workspaces: Requires `terraform workspace select` before every command; state differences hidden from version control
- Separate directories (infra/staging/, infra/production/): Duplicates module references and main.tf; harder to keep configurations in sync
- Hybrid (workspaces for compute, .tfvars for values): Unnecessary complexity; doesn't solve the discoverability problem

**Implementation**:
- `infra/environments/staging.tfvars`: ECS 0.25 vCPU/0.5 GB RAM, db.t4g.micro, 1-day backups, 7-day logs, NAT instance
- `infra/environments/production.tfvars`: ECS 1 vCPU/2 GB RAM, db.t4g.small Multi-AZ, 30-day backups, 30-day logs, NAT Gateway
- Root `main.tf` parameterized with variables
- GitHub Actions passes `-var-file=environments/staging.tfvars` for staging deploys

**Source**: Clarification #2

---

## Decision 3: Database Migration Concurrency Control

**Chosen**: goose advisory lock mechanism (lock table)

**Rationale**:
- All ECS tasks attempt migration on startup; only one proceeds via PostgreSQL advisory lock
- Industry-standard pattern (Flyway, Liquibase, goose all use advisory locks)
- Prevents migration conflicts during rolling ECS deployments (2+ tasks starting simultaneously)
- No separate pre-deployment migration job needed; keeps migrations as part of ECS task lifecycle
- Other tasks wait for lock, then skip migrations (already applied by winning task)

**Alternatives Considered**:
- First task runs migrations, others retry health check: Coordination complexity; tasks may restart before migrations complete
- Separate GitHub Actions migration job: Adds deployment step; requires database credentials in GitHub; slows deployments
- Environment variable flag (RUN_MIGRATIONS=true for task index 0): ECS task index not reliably available; brittle

**Implementation**:
- goose uses `SELECT pg_advisory_lock()` before migration execution
- Losing tasks block on lock, then detect migrations already applied and skip
- Health check (`/healthz`) only returns 200 OK after migrations complete successfully
- ALB health check grace period allows time for migrations (60s per SC-005)

**Source**: Clarification #1

---

## Decision 4: Database Connection Pool Sizing (MVP Staging)

**Chosen**: Min 5, Max 25 connections per ECS task

**Rationale**:
- Staging MVP targets < 50 concurrent users with 1-2 ECS tasks auto-scaling
- db.t4g.micro supports ~100 total connections
- 2 tasks × 25 connections = 50 max connections, leaving headroom for admin queries and monitoring
- Min 5 keeps persistent connections available, avoiding cold-start connection overhead
- Balanced for MVP scale; can tune production based on load test results

**Alternatives Considered**:
- Min 2, Max 10 (conservative): May exhaust pool under burst traffic; increases latency due to connection acquisition waits
- Min 10, Max 50 (optimized for low latency): Risks exhausting db.t4g.micro connection limit; over-provisioned for < 50 concurrent users
- Min 1, Max 100 (maximum utilization): Single task could monopolize database; no room for admin connections

**Implementation**:
- Configure pgx connection pool in `internal/database/client.go`
- Environment variables: `DB_POOL_MIN_CONNS=5`, `DB_POOL_MAX_CONNS=25`
- Monitor CloudWatch metric `DatabaseConnections` to validate pool sizing
- Production values will be adjusted based on load test observations (clarification #9)

**Source**: Clarification #4

---

## Decision 5: ECS Auto-Scaling Configuration (Staging)

**Chosen**: Min 1, Max 2 tasks with 70% CPU target

**Rationale**:
- Minimal cost for MVP staging (1 task running most of time, scales to 2 under load)
- 70% CPU target provides headroom before scaling trigger
- Staging MVP targets < 50 concurrent users; 1-2 tasks sufficient
- Cost: ~$9-18/month for ECS tasks (0.25 vCPU/0.5 GB RAM per task)
- Staging budget: $200/month total (well within target)

**Alternatives Considered**:
- Min 1, Max 3 (more headroom): Higher cost ($27/month max); unnecessary for MVP staging < 50 users
- Min 2, Max 5 (optimized for availability): Doubles baseline cost; staging doesn't require high availability
- Min 1, Max 5 with 60% CPU target (maximum flexibility): Higher cost, lower CPU threshold may trigger unnecessary scaling

**Implementation**:
- ECS service auto-scaling policy: `TargetTrackingScaling` on `ECSServiceAverageCPUUtilization`
- Staging: min 1, max 2, target 70%
- Production: min 2, max 20, target 70% (TBD after load testing per clarification #9)
- CloudWatch alarm on CPU > 80% sustained 5 min (scaling lag indicator)

**Source**: Clarification #5, user request for cheapest viable option

---

## Decision 6: ECS Task Resource Allocation (Staging)

**Chosen**: 0.25 vCPU / 0.5 GB RAM ARM64 (cheapest Fargate tier)

**Rationale**:
- Cheapest Fargate configuration: ~$0.012/hour per task (~$9/month at 100% utilization)
- Go API has minimal memory footprint; 0.5 GB sufficient for HTTP server + pgx connection pool + AI SDK
- ARM64 Graviton2 provides 20% cost savings vs x86 without code changes
- Staging MVP baseline cost: $9/month (1 task), max $18/month (2 tasks)
- Can validate sufficiency via load testing; increase if CPU/memory pressure observed

**Alternatives Considered**:
- 0.5 vCPU / 1 GB RAM: Double the cost (~$18/task/month); may be over-provisioned for Go API
- 1 vCPU / 2 GB RAM: Production-like config, but 4x cost (~$36/task/month); unnecessary for staging MVP
- Configurable per environment variable: Adds complexity; Fargate task definitions require fixed CPU/memory

**Implementation**:
- Staging ECS task definition: `cpu = "256"` (0.25 vCPU), `memory = "512"` (0.5 GB), `runtime_platform.cpu_architecture = "ARM64"`
- Production task definition: `cpu = "1024"` (1 vCPU), `memory = "2048"` (2 GB), ARM64
- Monitor CloudWatch metrics `CPUUtilization` and `MemoryUtilization` to validate sizing
- If staging exceeds 80% sustained CPU or memory, increase to 0.5 vCPU / 1 GB

**Source**: Clarification #6, user cost optimization request

---

## Decision 7: CloudWatch Log Retention (Staging)

**Chosen**: 7 days retention for staging

**Rationale**:
- Reasonable debugging window for MVP staging (sufficient to diagnose recent issues)
- Cost-balanced: ~$0.50/GB/month stored (7-day retention cheaper than 14-day or 30-day)
- Issues in staging typically caught immediately during testing; 7-day window adequate
- Production uses 30-day retention (clarification #7) for incident forensics

**Alternatives Considered**:
- 3 days retention: Cheapest option, but insufficient if engineer unavailable for long weekend
- 14 days retention: Extended debugging window, but higher cost for staging MVP
- 30 days retention: Production-like, but unnecessary storage cost for staging

**Implementation**:
- CloudWatch log group retention in `modules/ecs/main.tf`: `retention_in_days = var.log_retention_days`
- Staging .tfvars: `log_retention_days = 7`
- Production .tfvars: `log_retention_days = 30`
- All backend logs sent to `/ecs/backend-api-${environment}` log group

**Source**: Clarification #7

---

## Decision 8: RDS Automated Backup Retention (Staging)

**Chosen**: 1 day retention for staging

**Rationale**:
- Minimum viable backup window (AWS requires ≥ 1 day for automated backups)
- Cheapest backup storage cost (1-day snapshots only)
- Staging data is non-production; 1-day recovery point acceptable for MVP
- Production uses 30-day retention for extended recovery window

**Alternatives Considered**:
- 7 days retention: Matches cloud-and-environments.md, but higher storage cost for staging MVP
- 14 days retention: Extended recovery window, unnecessary for staging
- 0 days (no backups): Not allowed by AWS for automated backups; manual snapshots only

**Implementation**:
- RDS instance configuration in `modules/rds/main.tf`: `backup_retention_period = var.backup_retention_days`
- Staging .tfvars: `backup_retention_days = 1`
- Production .tfvars: `backup_retention_days = 30`
- Automated daily backups at 3 AM UTC (low-traffic window)

**Source**: Clarification #8, user cost optimization request

---

## Decision 9: Production Configuration Validation Strategy

**Chosen**: Load test staging with production-like scenarios; adjust production .tfvars based on observed metrics before provisioning

**Rationale**:
- Evidence-based approach prevents over-provisioning and provides accurate cost estimates
- Staging load tests simulate 500 concurrent VUs (production target per NFR-SCALE-001)
- Observe CPU/memory utilization, database query patterns, API latency, error rates
- Adjust production ECS resource allocation, auto-scaling bounds, RDS instance class based on actual data
- Defer production provisioning until metrics validate configuration

**Alternatives Considered**:
- Use staging config as baseline; defer tuning: Risk of under-provisioning production, harder to fix post-launch
- Provision production with conservative defaults; tune after launch: Wastes budget on over-provisioning, requires production traffic to validate
- Document assumptions in separate analysis: Analysis becomes stale; no validation loop

**Implementation**:
- Before production provisioning: Run k6 load test on staging with 500 VUs sustained 10 min
- Capture metrics: ECS CPU/memory utilization, RDS connections/IOPS, API p95 latency, error rate
- Review production .tfvars: Adjust ECS CPU/memory, auto-scaling max, RDS instance class if needed
- Update production cost estimate based on validated configuration
- Re-run load test after production .tfvars changes to confirm targets met

**Source**: Clarification #9, user request for future production analysis

---

## Summary

All technology decisions finalized. Key outcomes:
- **Cost optimization**: Staging configured at minimum viable AWS tiers (~$150-180/month total)
- **Environment separation**: Explicit .tfvars files for staging/production (no workspace confusion)
- **Production readiness**: Load testing validates configuration before provisioning (prevents over-provisioning)
- **Deployment safety**: Advisory locks prevent migration conflicts during rolling ECS updates
- **API client simplicity**: Native fetch eliminates dependency, reduces bundle size

No further research required. Ready for Phase 1 (data model, contracts, quickstart).
