# Implementation Plan: Cloud & Environments Strategy

**Branch**: `003-cloud-env-strategy` | **Date**: 2026-07-03 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/003-cloud-env-strategy/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

This feature establishes the cloud infrastructure foundation for TrAIveler, defining and provisioning two AWS environments (staging and production) using Infrastructure as Code (Terraform). The staging environment serves as the active MVP deployment target with cost-optimized configuration ($200/month budget), while production remains IaC-defined but dormant until alpha release. The architecture uses ECS Fargate for containerized backend services (chosen over Lambda for AI workload requirements: unlimited execution time, no payload limits, persistent HTTP connections to Anthropic API), Application Load Balancer for HTTPS traffic, RDS PostgreSQL for database persistence, and S3+CloudFront for frontend static assets. CI/CD pipelines via GitHub Actions with OIDC federation automate deployments without long-lived credentials. All infrastructure is version-controlled, workspace-isolated (Terraform workspaces), and tagged for cost allocation by environment and service.

## Technical Context

**Infrastructure as Code (IaC)**: Terraform 1.5+ (HCL), AWS provider ~> 5.0

**Primary Dependencies**:
- **Terraform modules**: VPC, ECS, RDS, ALB, S3, CloudFront, IAM, Secrets Manager
- **AWS CLI**: 2.x for manual operations and troubleshooting
- **Docker**: 24.x for backend container builds (multi-stage, linux/arm64 Graviton2)
- **GitHub Actions**: CI/CD platform with OIDC federation for AWS authentication

**Storage**:
- **Terraform State**: S3 remote backend with versioning and encryption (`s3://trAIveler-terraform-state`)
- **State Locking**: DynamoDB table (`trAIveler-terraform-locks`)
- **Secrets**: AWS Secrets Manager for database credentials, API keys, JWT secrets
- **Configuration**: AWS Parameter Store for non-sensitive environment-specific values
- **Frontend Assets**: S3 bucket with CloudFront CDN
- **Container Images**: Amazon ECR private repository

**Testing**:
- **IaC Validation**: `terraform validate`, `terraform plan` (syntax and logic checks)
- **Infrastructure Tests**: Manual validation scenarios in [quickstart.md](quickstart.md)
- **Deployment Tests**: GitHub Actions workflows with plan-on-PR, apply-on-merge
- **Health Checks**: ALB target group health checks against `/healthz` endpoint
- **Load Testing**: k6 scripts (from spec 002) for performance validation

**Target Platform**: AWS us-east-1 region (single region for MVP)

**Project Type**: Infrastructure provisioning and CI/CD automation (not application code)

**Performance Goals**:
- Infrastructure provisioning: < 30 minutes from `terraform apply` to all resources available
- Deployment latency: < 15 minutes from git push to ECS service stable (staging)
- ALB health check: < 100ms response time from `/healthz`
- RDS connection establishment: < 500ms from ECS task

**Constraints**:
- **Cost**: Staging $200/month, Production $300-400/month (when provisioned)
- **Availability**: Staging single-AZ (no HA requirement), Production Multi-AZ (99.9% uptime)
- **Security**: No long-lived AWS credentials; OIDC federation only
- **Compliance**: GDPR-aware design (right-to-deletion, data minimization, US-based acceptable for MVP per spec 002)
- **Network**: VPC CIDR blocks must not overlap (staging: 10.0.0.0/16, production: 10.1.0.0/16)

**Scale/Scope**:
- **Environments**: 2 (staging active, production dormant)
- **Infrastructure Resources**: ~45 AWS resources per environment (VPC, subnets, security groups, ECS, RDS, ALB, S3, CloudFront)
- **ECS Tasks**: 1-5 concurrent tasks (staging), 2-20 (production when active)
- **Database**: < 50 concurrent users (from spec 001), db.t4g.micro sufficient for staging
- **Traffic**: < 100 RPS sustained (MVP load profile), auto-scaling handles bursts

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Test-First Development ✅ ADAPTED

**Status**: **PASS** (adapted for infrastructure)

**Rationale**: Infrastructure-as-code requires validation-first approach analogous to TDD:
- **"Red"**: `terraform plan` shows intended changes before applying
- **"Green"**: `terraform apply` provisions resources matching plan
- **"Refactor"**: Iterative improvement of module structure and variable design

Infrastructure tests manifest as:
- **Unit-level**: Terraform `validate` checks syntax and logic
- **Integration-level**: Manual validation scenarios in [quickstart.md](quickstart.md) verify end-to-end provisioning
- **E2E-level**: GitHub Actions workflows test full CI/CD pipeline (plan-on-PR, apply-on-merge)

**Adaptation**: Traditional TDD (write test, write code, refactor) becomes **Plan-Apply-Validate** for infrastructure:
1. Write Terraform configuration
2. Run `terraform plan` (equivalent to "failing test" — shows what will change)
3. Review plan output (peer review required)
4. Run `terraform apply` (equivalent to "make test pass")
5. Validate with [quickstart.md](quickstart.md) scenarios (integration tests)
6. Refactor modules for reusability

**Commitment**: All Terraform changes MUST be peer-reviewed via PR with plan output visible before merge.

---

### II. Simplicity — KISS & DRY ✅ PASS

**Status**: **PASS**

**Rationale**:
- Terraform modules encapsulate reusable infrastructure patterns (VPC, ECS, RDS, ALB)
- Shared configuration values (tags, naming conventions) defined once in `variables.tf` and referenced everywhere
- No speculative abstraction: each module solves a concrete need from spec 003
- Module interfaces follow HashiCorp best practices (documented in [contracts/infrastructure.md](contracts/infrastructure.md))
- Workspace-based environment isolation avoids duplicating Terraform code across staging/production

**Evidence**: Module structure in [research.md](research.md) shows shared modules called with environment-specific tfvars, not separate codebases per environment.

---

### III. Code Quality & Consistency ✅ PASS

**Status**: **PASS**

**Rationale**:
- **Terraform**: `terraform fmt` enforces formatting; `terraform validate` enforces syntax
- **HCL linting**: `tflint` will be added to CI workflow (part of implementation tasks)
- **Security scanning**: `tfsec` and `checkov` will scan for misconfigurations (hardcoded secrets, overly permissive security groups)
- **Naming conventions**: All resources follow `trAIveler-<component>-<environment>` pattern
- **Tagging**: All resources tagged with `Project`, `Environment`, `ManagedBy`, `CostCenter` (enforced via AWS Config)

**Commit gates**: GitHub Actions workflows fail PR if:
- `terraform fmt -check` finds unformatted files
- `terraform validate` reports errors
- `tfsec` or `checkov` find HIGH severity issues

---

### IV. Accessible & Token-Driven UI ⚠️ NOT APPLICABLE

**Status**: **N/A** (no user-facing UI in infrastructure code)

**Rationale**: This principle governs React components and frontend design. Infrastructure code provisions cloud resources, not UI components. Accessibility requirements apply to application code (specs 001-002), not Terraform modules.

**Verification**: UI accessibility is validated in frontend deployment (spec 001, spec 002 Phase 4), not in infrastructure provisioning.

---

### V. Secure Configuration ✅ PASS

**Status**: **PASS**

**Rationale**:
- **No hardcoded secrets**: All sensitive values (database passwords, API keys, JWT secrets) stored in AWS Secrets Manager
- **No credentials in code**: GitHub Actions uses OIDC federation; no long-lived AWS access keys
- **Environment variables**: Non-secret config (API URLs, environment names) passed via Terraform variables
- **Secrets rotation**: Secrets Manager ARNs referenced in ECS task definitions; secret values can be rotated without code changes
- **Audit logging**: Terraform state stored in S3 with versioning; all `terraform apply` executions logged via CloudWatch

**Evidence**: [contracts/infrastructure.md](contracts/infrastructure.md) documents Secrets Manager naming convention and IAM policies for secret access.

---

## Re-Check After Phase 1 Design

All constitution gates remain **PASS** after design phase. No changes required.

**Summary**:
- ✅ Test-First Development (adapted to Plan-Apply-Validate workflow)
- ✅ Simplicity — KISS & DRY (modular Terraform, shared variables)
- ✅ Code Quality & Consistency (fmt, validate, tflint, tfsec, checkov)
- ⚠️ Accessible & Token-Driven UI (N/A for infrastructure)
- ✅ Secure Configuration (Secrets Manager, OIDC, no hardcoded credentials)

## Project Structure

### Documentation (this feature)

```text
specs/003-cloud-env-strategy/
├── spec.md              # Feature specification (user requirements)
├── plan.md              # This file (implementation plan)
├── research.md          # Phase 0 output (technology decisions)
├── data-model.md        # Phase 1 output (infrastructure configuration schemas)
├── quickstart.md        # Phase 1 output (validation scenarios)
├── contracts/           # Phase 1 output (Terraform module and CI/CD interfaces)
│   └── infrastructure.md
└── checklists/          # Quality validation
    └── requirements.md
```

### Source Code (repository root)

**Infrastructure as Code** (new top-level directory):

```text
infra/
├── terraform/
│   ├── main.tf                 # Root module entry point
│   ├── variables.tf            # Input variables
│   ├── outputs.tf              # Output values
│   ├── terraform.tf            # Provider and backend configuration
│   ├── staging.tfvars          # Staging environment variables
│   ├── production.tfvars       # Production environment variables
│   ├── modules/
│   │   ├── vpc/
│   │   │   ├── main.tf         # VPC, subnets, NAT, route tables
│   │   │   ├── variables.tf
│   │   │   ├── outputs.tf
│   │   │   └── security_groups.tf
│   │   ├── ecs/
│   │   │   ├── main.tf         # ECS cluster, service, task definition
│   │   │   ├── variables.tf
│   │   │   ├── outputs.tf
│   │   │   └── autoscaling.tf
│   │   ├── rds/
│   │   │   ├── main.tf         # RDS instance, subnet group
│   │   │   ├── variables.tf
│   │   │   └── outputs.tf
│   │   ├── alb/
│   │   │   ├── main.tf         # ALB, listeners, target groups
│   │   │   ├── variables.tf
│   │   │   └── outputs.tf
│   │   ├── s3-cloudfront/
│   │   │   ├── main.tf         # S3 bucket, CloudFront distribution
│   │   │   ├── variables.tf
│   │   │   └── outputs.tf
│   │   └── iam/
│   │       ├── main.tf         # IAM roles for ECS tasks, GitHub Actions
│   │       ├── variables.tf
│   │       └── outputs.tf
│   └── scripts/
│       ├── bootstrap-state.sh  # Create S3 backend and DynamoDB lock table
│       └── init-secrets.sh     # Initialize Secrets Manager values
├── docker/
│   └── backend/
│       └── Dockerfile          # Multi-stage Go build for linux/arm64
└── iam/
    ├── github-actions-trust-policy-staging.json
    └── github-actions-trust-policy-production.json
```

**CI/CD Workflows** (extends existing `.github/workflows/`):

```text
.github/
└── workflows/
    ├── terraform-plan.yml      # Validate Terraform on PR
    ├── terraform-apply.yml     # Deploy infrastructure on merge to main
    ├── backend-deploy.yml      # Build and deploy backend to ECS
    ├── frontend-deploy.yml     # Build and deploy frontend to S3+CloudFront
    ├── backend-lint.yml        # (from spec 001-T076) — EXTENDED with Docker linting
    └── backend-test.yml        # (from spec 001-T078) — EXTENDED with integration tests
```

**Application Code** (no changes; spec 003 provisions infrastructure for existing app structure):

```text
backend/                        # (from spec 001) — No structural changes
├── cmd/
├── internal/
├── pkg/
└── Dockerfile                  # MOVED to infra/docker/backend/Dockerfile

frontend/                       # (from spec 001) — No structural changes
├── src/
├── public/
└── vite.config.ts

e2e/                            # (from spec 001) — No structural changes
└── tests/
```

**Notes**:
- Infrastructure code (`infra/`) is a new top-level directory alongside `backend/`, `frontend/`, and `e2e/`
- Terraform modules are organized by AWS service/component for maximum reusability
- Workspace-specific variable files (`staging.tfvars`, `production.tfvars`) provide environment configuration without code duplication
- CI/CD workflows are additive to existing workflows from spec 001; some workflows are extended (e.g., backend-lint now includes Docker linting via `hadolint`)
- The `backend/Dockerfile` from spec 001-T002 will be moved to `infra/docker/backend/Dockerfile` for better organization
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
