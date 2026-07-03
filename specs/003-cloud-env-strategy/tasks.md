# Tasks: Cloud & Environments Strategy — TrAIveler Infrastructure

**Input**: Design documents from `specs/003-cloud-env-strategy/`

**Prerequisites**: [plan.md](plan.md) · [spec.md](spec.md) · [research.md](research.md) · [data-model.md](data-model.md) · [contracts/infrastructure.md](contracts/infrastructure.md) · [quickstart.md](quickstart.md)

**Tests**: Not explicitly requested in spec — infrastructure validation via manual scenarios in quickstart.md.
Terraform validation (`terraform validate`, `terraform plan`) serves as the "test" layer per Constitution I.

**Cross-spec awareness**: This spec provisions infrastructure for application code defined in specs 001-002.
No tasks duplicate work from previous specs. Backend Dockerfile (from 001-T002 context) will be relocated.

**Format**: `- [ ] [ID] [P?] [Story?] Description — file path`
- `[P]` = parallelizable (no incomplete dependencies, different files)
- `[US#]` = user story label (Phase 3+ only)

---

## Phase 1: Setup

**Purpose**: Initialize infrastructure codebase, install tooling, create directory structure.
No AWS resources created in this phase — only local repository setup.

- [ ] T001 Create top-level `infra/` directory with subdirectories: `terraform/`, `docker/`, `iam/`, `scripts/` — repository root
- [ ] T002 [P] Create Terraform root module structure: `main.tf`, `variables.tf`, `outputs.tf`, `terraform.tf` — `infra/terraform/`
- [ ] T003 [P] Create Terraform workspace-specific variable files: `staging.tfvars`, `production.tfvars` — `infra/terraform/`
- [ ] T004 [P] Create module directories: `modules/vpc/`, `modules/ecs/`, `modules/rds/`, `modules/alb/`, `modules/s3-cloudfront/`, `modules/iam/` — `infra/terraform/modules/`
- [ ] T005 [P] Install Terraform 1.5+ locally and verify version: `terraform version` — local development environment
- [ ] T006 [P] Install AWS CLI 2.x and configure default region (us-east-1) — local development environment
- [ ] T007 [P] Install Docker 24.x for backend container builds — local development environment
- [ ] T008 [P] Add `infra/terraform/.gitignore` excluding: `*.tfstate`, `*.tfstate.backup`, `.terraform/`, `*.tfplan`, `*.tfvars` (workspace files are tracked, but local overrides are not) — `infra/terraform/.gitignore`

---

## Phase 2: Foundational — Terraform State & OIDC Setup

**Purpose**: Bootstrap remote state management and GitHub Actions authentication.
These are prerequisites for ALL subsequent infrastructure work — no Terraform modules can be applied without remote state.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T009 Create Terraform state bootstrap script: provisions S3 bucket (`trAIveler-terraform-state`) with versioning and encryption, DynamoDB table (`trAIveler-terraform-locks`) with `LockID` hash key — `infra/terraform/scripts/bootstrap-state.sh`
- [ ] T010 Create Terraform backend configuration using S3 bucket and DynamoDB table from T009; backend key pattern: `env/${terraform.workspace}/terraform.tfstate` — `infra/terraform/terraform.tf`
- [ ] T011 [P] Create AWS provider configuration with default tags (Project, Environment, ManagedBy, CostCenter); region: us-east-1 — `infra/terraform/terraform.tf`
- [ ] T012 [P] Create GitHub Actions OIDC trust policy JSON templates for staging and production environments — `infra/iam/github-actions-trust-policy-staging.json`, `infra/iam/github-actions-trust-policy-production.json`
- [ ] T013 Create IAM OIDC identity provider Terraform module: configures GitHub OIDC provider in AWS, creates IAM roles for staging and production with trust policies from T012 — `infra/terraform/modules/iam/github-oidc.tf`

**Checkpoint**: Run `bootstrap-state.sh`, then `terraform init` succeeds with remote backend configured. OIDC roles created manually or via IAM module for GitHub Actions authentication.

---

## Phase 3: User Story 1 — Infrastructure Provisioning & Environment Setup (Priority: P1) 🎯 MVP

**Goal**: Provision complete staging environment (VPC, ECS, RDS, ALB, S3, CloudFront) using Terraform.
Production environment is IaC-defined but not applied (remains dormant until alpha release).

**Independent Test**: Quickstart Scenario 3 — `terraform apply` for staging succeeds, all 45+ resources created, outputs include ALB DNS, ECS cluster name, RDS endpoint.

### Terraform Modules

- [ ] T014 [P] [US1] Create VPC module: provisions VPC with CIDR from tfvars, 2 public subnets (ALB), 2 private subnets (ECS, RDS), internet gateway, route tables — `infra/terraform/modules/vpc/main.tf`
- [ ] T015 [P] [US1] Create VPC security groups: `alb-sg` (ingress 80/443 from 0.0.0.0/0), `ecs-sg` (ingress 8080 from alb-sg), `rds-sg` (ingress 5432 from ecs-sg) — `infra/terraform/modules/vpc/security_groups.tf`
- [ ] T016 [P] [US1] Create VPC NAT resource: conditional NAT instance (staging) or NAT Gateway (production) based on tfvars variable — `infra/terraform/modules/vpc/nat.tf`
- [ ] T017 [P] [US1] Define VPC module variables: `environment`, `vpc_cidr`, `availability_zones`, `nat_gateway_type` — `infra/terraform/modules/vpc/variables.tf`
- [ ] T018 [P] [US1] Define VPC module outputs: `vpc_id`, `public_subnet_ids`, `private_subnet_ids`, `alb_security_group_id`, `ecs_security_group_id`, `rds_security_group_id` — `infra/terraform/modules/vpc/outputs.tf`
- [ ] T019 [P] [US1] Create ALB module: provisions Application Load Balancer in public subnets, HTTP listener (port 80 → redirect to 443), HTTPS listener (port 443 → target group), target group for ECS (port 8080, health check `/healthz`) — `infra/terraform/modules/alb/main.tf`
- [ ] T020 [P] [US1] Define ALB module variables: `environment`, `vpc_id`, `public_subnet_ids`, `alb_security_group_id`, `certificate_arn` (optional, for HTTPS) — `infra/terraform/modules/alb/variables.tf`
- [ ] T021 [P] [US1] Define ALB module outputs: `alb_arn`, `alb_dns_name`, `target_group_arn` — `infra/terraform/modules/alb/outputs.tf`
- [ ] T022 [P] [US1] Create RDS module: provisions PostgreSQL 15.4 instance with instance class, Multi-AZ flag, backup retention, maintenance window, security group, subnet group (private subnets) — `infra/terraform/modules/rds/main.tf`
- [ ] T023 [P] [US1] Define RDS module variables: `environment`, `instance_identifier`, `instance_class`, `vpc_id`, `private_subnet_ids`, `rds_security_group_id`, `multi_az`, `backup_retention_days`, `database_name`, `master_username`, `master_password_secret_arn` — `infra/terraform/modules/rds/variables.tf`
- [ ] T024 [P] [US1] Define RDS module outputs: `endpoint`, `instance_id`, `database_name` — `infra/terraform/modules/rds/outputs.tf`
- [ ] T025 [P] [US1] Create ECS module: provisions ECS cluster, task definition (Go backend container with environment variables and secrets from Secrets Manager), ECS service with ALB target group integration, auto-scaling (target tracking on CPU 70%) — `infra/terraform/modules/ecs/main.tf`
- [ ] T026 [P] [US1] Create ECS auto-scaling configuration: target tracking scaling policy (CPU 70%), scale-out cooldown 60s, scale-in cooldown 300s, min/max capacity from tfvars — `infra/terraform/modules/ecs/autoscaling.tf`
- [ ] T027 [P] [US1] Define ECS module variables: `environment`, `cluster_name`, `vpc_id`, `private_subnet_ids`, `alb_target_group_arn`, `ecs_security_group_id`, `task_cpu`, `task_memory`, `min_capacity`, `max_capacity`, `ecr_image_uri`, `database_url_secret_arn`, `anthropic_api_key_secret_arn`, `jwt_secret_secret_arn` — `infra/terraform/modules/ecs/variables.tf`
- [ ] T028 [P] [US1] Define ECS module outputs: `ecs_cluster_arn`, `ecs_service_name`, `task_definition_arn`, `service_arn` — `infra/terraform/modules/ecs/outputs.tf`
- [ ] T029 [P] [US1] Create S3+CloudFront module: provisions S3 bucket for frontend static assets with website hosting, CloudFront distribution with S3 origin, HTTPS only, caching headers — `infra/terraform/modules/s3-cloudfront/main.tf`
- [ ] T030 [P] [US1] Define S3+CloudFront module variables: `environment`, `bucket_name`, `cloudfront_price_class` (staging: PriceClass_100, production: PriceClass_200) — `infra/terraform/modules/s3-cloudfront/variables.tf`
- [ ] T031 [P] [US1] Define S3+CloudFront module outputs: `s3_bucket_name`, `cloudfront_distribution_id`, `cloudfront_domain_name` — `infra/terraform/modules/s3-cloudfront/outputs.tf`
- [ ] T032 [P] [US1] Create IAM module: provisions ECS task execution role (ECR pull, CloudWatch Logs write), ECS task role (Secrets Manager read, S3 access for application), GitHub Actions roles (from T012-T013) — `infra/terraform/modules/iam/main.tf`
- [ ] T033 [P] [US1] Define IAM module variables: `environment`, `ecr_repository_arn`, `secrets_manager_arns`, `s3_bucket_arns` — `infra/terraform/modules/iam/variables.tf`
- [ ] T034 [P] [US1] Define IAM module outputs: `ecs_task_execution_role_arn`, `ecs_task_role_arn`, `github_actions_role_arn` — `infra/terraform/modules/iam/outputs.tf`

### Root Module Integration

- [ ] T035 [US1] Wire VPC module in root `main.tf`: call `modules/vpc` with staging/production-specific CIDR blocks and NAT type — `infra/terraform/main.tf`
- [ ] T036 [US1] Wire ALB module in root `main.tf`: call `modules/alb` with VPC outputs (public subnets, security group) — `infra/terraform/main.tf`
- [ ] T037 [US1] Wire RDS module in root `main.tf`: call `modules/rds` with VPC outputs (private subnets, security group) and environment-specific instance class, Multi-AZ flag — `infra/terraform/main.tf`
- [ ] T038 [US1] Wire ECS module in root `main.tf`: call `modules/ecs` with VPC outputs, ALB target group ARN, environment-specific task sizing and capacity — `infra/terraform/main.tf`
- [ ] T039 [US1] Wire S3+CloudFront module in root `main.tf`: call `modules/s3-cloudfront` with environment-specific bucket name and CloudFront price class — `infra/terraform/main.tf`
- [ ] T040 [US1] Wire IAM module in root `main.tf`: call `modules/iam` with resource ARNs (ECR, Secrets Manager, S3) from other modules — `infra/terraform/main.tf`
- [ ] T041 [US1] Define root module variables: `environment`, `aws_region`, `vpc_cidr`, `availability_zones`, `ecs_task_cpu`, `ecs_task_memory`, `ecs_min_capacity`, `ecs_max_capacity`, `rds_instance_class`, `rds_multi_az`, `rds_backup_retention_days`, `nat_gateway_type`, `cost_budget_usd`, `tags` — `infra/terraform/variables.tf`
- [ ] T042 [US1] Define root module outputs: all outputs from modules (VPC, ALB, RDS, ECS, S3+CloudFront, IAM) — `infra/terraform/outputs.tf`
- [ ] T043 [US1] Populate `staging.tfvars` with staging-specific values: `vpc_cidr = "10.0.0.0/16"`, `ecs_task_cpu = "512"`, `ecs_task_memory = "1024"`, `ecs_min_capacity = 1`, `ecs_max_capacity = 5`, `rds_instance_class = "db.t4g.micro"`, `rds_multi_az = false`, `nat_gateway_type = "instance"`, `cost_budget_usd = 200` — `infra/terraform/staging.tfvars`
- [ ] T044 [US1] Populate `production.tfvars` with production-specific values: `vpc_cidr = "10.1.0.0/16"`, `ecs_task_cpu = "1024"`, `ecs_task_memory = "2048"`, `ecs_min_capacity = 2`, `ecs_max_capacity = 20`, `rds_instance_class = "db.t4g.small"`, `rds_multi_az = true`, `nat_gateway_type = "gateway"`, `cost_budget_usd = 400` — `infra/terraform/production.tfvars`

**Checkpoint**: `terraform workspace select staging && terraform plan -var-file=staging.tfvars` succeeds with ~45 resources to create. Production plan also succeeds but is NOT applied (dormant).

---

## Phase 4: User Story 2 — Secrets & Configuration Management (Priority: P1)

**Goal**: Initialize secrets in AWS Secrets Manager for database credentials, Anthropic API key, and JWT secret.
Terraform module references ensure ECS tasks can retrieve these at runtime.

**Independent Test**: Quickstart Scenario 5 step 2 — secrets exist in Secrets Manager, ECS tasks can read them (verified by successful backend deployment with database connection).

- [ ] T045 [P] [US2] Create secrets initialization script: uses AWS CLI to create secrets in Secrets Manager with naming pattern `${environment}/${service}/${secret_name}` — `infra/terraform/scripts/init-secrets.sh`
- [ ] T046 [US2] Create Terraform data sources for Secrets Manager: reference existing secrets for database URL, Anthropic API key, JWT secret by name — `infra/terraform/secrets.tf`
- [ ] T047 [US2] Update ECS task definition in T025 to reference secret ARNs from T046 in `secrets` block (not `environment` block) — `infra/terraform/modules/ecs/main.tf`

**Checkpoint**: After running `init-secrets.sh`, secrets exist in Secrets Manager. ECS task definition includes `secrets` configuration (not yet deployed).

---

## Phase 5: User Story 3 — CI/CD Pipeline & Deployment Promotion (Priority: P1)

**Goal**: Automate Terraform validation, backend Docker builds, frontend static builds, and deployment to staging.
Production deployment is manual-only (workflow_dispatch).

**Independent Test**: Quickstart Scenario 5 — push to main triggers backend deployment, ECS service updates with new task revision, health checks pass.

### Terraform CI/CD Workflows

- [ ] T048 [P] [US3] Create `terraform-plan.yml` GitHub Actions workflow: triggers on PR with paths `infra/terraform/**`, checks out code, configures AWS credentials via OIDC (read-only role), runs `terraform fmt -check`, `terraform init`, `terraform validate`, `terraform workspace select staging`, `terraform plan -var-file=staging.tfvars`, posts plan output as PR comment — `.github/workflows/terraform-plan.yml`
- [ ] T049 [P] [US3] Create `terraform-apply.yml` GitHub Actions workflow: triggers on push to main with paths `infra/terraform/**` (auto-deploy staging) OR workflow_dispatch (manual production deploy), configures AWS credentials via OIDC (read-write role), runs `terraform init`, `terraform workspace select ${{ inputs.environment }}`, `terraform plan -var-file=${{ inputs.environment }}.tfvars`, `terraform apply -auto-approve` (with manual approval step for production), exports outputs — `.github/workflows/terraform-apply.yml`

### Backend Deployment Workflow

- [ ] T050 [P] [US3] Create multi-stage Dockerfile for Go backend: stage 1 builds Go binary targeting `linux/arm64` with `CGO_ENABLED=0`, stage 2 uses `public.ecr.aws/docker/library/golang:1.24-alpine` base, copies binary, exposes port 8080 — `infra/docker/backend/Dockerfile`
- [ ] T051 [P] [US3] Create `.dockerignore` for backend: excludes `*.md`, `tests/`, `.git/`, `.env*` — `infra/docker/backend/.dockerignore`
- [ ] T052 [US3] Create `backend-deploy.yml` GitHub Actions workflow: triggers on push to main with paths `backend/**` (auto-deploy staging) OR workflow_dispatch (manual production deploy), configures AWS credentials via OIDC, logs in to ECR, builds Docker image with platform `linux/arm64` and tag `v<version>-<sha>`, pushes to ECR, updates ECS task definition with new image URI, registers new task definition, updates ECS service, waits for service stability — `.github/workflows/backend-deploy.yml`

### Frontend Deployment Workflow

- [ ] T053 [P] [US3] Create `frontend-deploy.yml` GitHub Actions workflow: triggers on push to main with paths `frontend/**` (auto-deploy staging) OR workflow_dispatch (manual production deploy), sets up Node.js 20, installs dependencies (`npm ci`), builds production bundle (`VITE_API_URL=${{ vars.API_URL_STAGING }} npm run build`), configures AWS credentials via OIDC, syncs `dist/` to S3 (`aws s3 sync dist/ s3://${{ vars.S3_BUCKET_STAGING }}/ --delete`), invalidates CloudFront cache (`aws cloudfront create-invalidation --distribution-id ${{ vars.CLOUDFRONT_DISTRIBUTION_ID_STAGING }} --paths "/*"`), waits for invalidation — `.github/workflows/frontend-deploy.yml`

### Deployment Rollback

- [ ] T054 [P] [US3] Update ECS service configuration in T025 to enable deployment circuit breaker: `deployment_circuit_breaker { enable = true, rollback = true }`, health check grace period 60 seconds — `infra/terraform/modules/ecs/main.tf`

**Checkpoint**: GitHub Actions workflows exist. Merging to main triggers staging deployment. Manual workflow_dispatch enables production deployment (not automatic).

---

## Phase 6: User Story 4 — Cost Monitoring & Budget Controls (Priority: P2)

**Goal**: Set up AWS Cost Explorer tags, budget alerts, and CloudWatch cost metrics.
Cost tracking is passive (observability only); no active cost-control mechanisms beyond tagging and alerts.

**Independent Test**: Quickstart Scenario 10 — AWS Cost Explorer shows spending by Environment tag, budget alert configured for $200 staging threshold.

- [ ] T055 [P] [US4] Create AWS Budget Terraform resource for staging: monthly budget $200, alerts at 80% ($160) and 100% ($200), SNS topic for notifications — `infra/terraform/budgets.tf`
- [ ] T056 [P] [US4] Create AWS Budget Terraform resource for production: monthly budget $400, alerts at 80% ($320) and 100% ($400), SNS topic for notifications — `infra/terraform/budgets.tf`
- [ ] T057 [P] [US4] Verify all Terraform modules use default tags from provider (T011): validate tags propagate to all resources (Project, Environment, ManagedBy, CostCenter) — verify in `infra/terraform/modules/**/main.tf`
- [ ] T058 [US4] Create AWS Config rule to enforce required tags: rule triggers on resource creation, fails if missing Project, Environment, ManagedBy tags — `infra/terraform/config-rules.tf`

**Checkpoint**: Budgets created in AWS. Cost Explorer can filter by Environment and CostCenter tags. Alerts send to SNS topic (requires subscription setup).

---

## Phase 7: User Story 5 — Infrastructure Observability & Health Monitoring (Priority: P2)

**Goal**: Create CloudWatch dashboards, log groups, and metric alarms for infrastructure health.
Application-level observability (from spec 002) is separate; this phase focuses on infrastructure metrics only.

**Independent Test**: Quickstart Scenario 4 — CloudWatch dashboard exists with ECS CPU/memory, ALB request count, RDS connections visible.

- [ ] T059 [P] [US5] Create CloudWatch log groups for ECS tasks: `/ecs/trAIveler-backend-staging` (retention 7 days), `/ecs/trAIveler-backend-production` (retention 30 days) — `infra/terraform/cloudwatch.tf`
- [ ] T060 [P] [US5] Create CloudWatch dashboard for staging: includes widgets for ECS task count, ECS CPU utilization, ECS memory utilization, ALB request count, ALB target response time, RDS CPU utilization, RDS database connections — `infra/terraform/cloudwatch-dashboard-staging.tf`
- [ ] T061 [P] [US5] Create CloudWatch dashboard for production (when active): same widgets as staging but for production resources — `infra/terraform/cloudwatch-dashboard-production.tf`
- [ ] T062 [P] [US5] Create CloudWatch metric alarm for ECS high CPU: triggers when staging ECS service CPU > 80% for 5 minutes, sends SNS notification — `infra/terraform/cloudwatch-alarms.tf`
- [ ] T063 [P] [US5] Create CloudWatch metric alarm for RDS high connections: triggers when staging RDS connections > 80 for 5 minutes, sends SNS notification — `infra/terraform/cloudwatch-alarms.tf`
- [ ] T064 [P] [US5] Create CloudWatch metric alarm for ALB unhealthy targets: triggers when staging ALB has 0 healthy targets for 2 minutes, sends SNS notification — `infra/terraform/cloudwatch-alarms.tf`

**Checkpoint**: CloudWatch dashboards exist (visible in AWS Console). Alarms configured but not yet triggering (no infrastructure under load).

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, linting, security scanning, and integration with existing CI workflows from spec 001.

- [ ] T065 [P] Add `tflint` configuration file: enables AWS plugin, sets minimum Terraform version 1.5.0 — `infra/terraform/.tflint.hcl`
- [ ] T066 [P] Add `tfsec` configuration file: enables all HIGH severity checks, ignores known false positives (if any) — `infra/terraform/.tfsec.yml`
- [ ] T067 [P] Extend `backend-lint.yml` workflow (from spec 001-T076) to include Docker linting: runs `hadolint infra/docker/backend/Dockerfile` — `.github/workflows/backend-lint.yml`
- [ ] T068 [P] Create infrastructure documentation README: overview of Terraform modules, quick start guide (references quickstart.md), links to contracts and research — `infra/README.md`
- [ ] T069 [P] Create Terraform module documentation: auto-generate module docs using `terraform-docs` for each module (vpc, ecs, rds, alb, s3-cloudfront, iam) — `infra/terraform/modules/**/README.md`
- [ ] T070 [P] Move backend Dockerfile from `backend/Dockerfile` (if exists from spec 001-T002) to `infra/docker/backend/Dockerfile` — relocate file, update references in GitHub Actions
- [ ] T071 [P] Create `.editorconfig` for Terraform files: indent 2 spaces, trim trailing whitespace — `infra/terraform/.editorconfig`
- [ ] T072 [P] Add pre-commit hook configuration: runs `terraform fmt` on staged `.tf` files — `.pre-commit-config.yaml` (if not already exists)

**Checkpoint**: All infrastructure code linted, formatted, and documented. CI workflows integrated with existing project workflows.

---

## Dependencies & Execution Order

### Critical Path (Must Complete Sequentially)

```
Phase 1 (T001–T008)            → repository structure created
         ↓
Phase 2 (T009–T013)            → Terraform remote state and OIDC roles ready
         ↓
Phase 3 (T014–T044)            → Terraform modules defined, staging provisioned
         ↓
Phase 4 (T045–T047)            → Secrets initialized, ECS tasks can retrieve them
         ↓
Phase 5 (T048–T054)            → CI/CD workflows enable deployment
         ↓
Phase 6 (T055–T058)            → Cost monitoring active (optional, can run after Phase 5)
         ↓
Phase 7 (T059–T064)            → Observability dashboards created (optional, can run after Phase 5)
         ↓
Phase 8 (T065–T072)            → Linting, docs, polish (can run in parallel with Phase 6-7)
```

### Parallel Execution Opportunities (Within Each Phase)

- **Phase 1**: All tasks parallelizable (T002–T008 can run concurrently)
- **Phase 2**: T012–T013 can run in parallel with T009–T011
- **Phase 3**: Module development highly parallel:
  - T014–T018 (VPC module) — independent
  - T019–T021 (ALB module) — independent
  - T022–T024 (RDS module) — independent
  - T025–T028 (ECS module) — independent
  - T029–T031 (S3+CloudFront module) — independent
  - T032–T034 (IAM module) — independent
  - After modules complete, T035–T044 (root module integration) runs sequentially
- **Phase 4**: T045–T046 parallel
- **Phase 5**: T048–T051 parallel (workflows + Dockerfile), T052–T053 parallel (deploy workflows), T054 sequential
- **Phase 6**: All tasks parallelizable
- **Phase 7**: All tasks parallelizable
- **Phase 8**: All tasks parallelizable

---

## Implementation Strategy

### MVP-First Approach

**Minimum Viable Infrastructure** (deliver first):
1. Phase 1–2: Bootstrap (T001–T013) — ~2 days
2. Phase 3: Staging environment only (T014–T043 excluding production tfvars) — ~5 days
3. Phase 4: Secrets for staging (T045–T047) — ~1 day
4. Phase 5: Staging CI/CD only (T048–T054 excluding production workflows) — ~3 days

**Total MVP**: ~11 days → Staging infrastructure operational, deployments automated

**Post-MVP Enhancements**:
- Production environment definition (T044 + production-specific workflow conditions) — ~1 day
- Cost monitoring (Phase 6) — ~2 days
- Observability dashboards (Phase 7) — ~2 days
- Polish (Phase 8) — ~2 days

### Incremental Delivery

Each phase delivers independently testable infrastructure:
- After Phase 2: Can run `terraform init` with remote state
- After Phase 3: Can provision staging environment (Quickstart Scenario 3)
- After Phase 4: Can deploy services with secrets (Quickstart Scenario 5)
- After Phase 5: Can merge code and auto-deploy (Quickstart Scenario 5-6)
- After Phase 6: Can track costs (Quickstart Scenario 10)
- After Phase 7: Can monitor infrastructure health (Quickstart Scenario 4)

---

## Task Summary

- **Total Tasks**: 72
- **Phase 1 (Setup)**: 8 tasks (all parallelizable)
- **Phase 2 (Foundational)**: 5 tasks (2 sequential, 3 parallel)
- **Phase 3 (US1 — Infrastructure)**: 31 tasks (26 parallel module development, 5 sequential integration)
- **Phase 4 (US2 — Secrets)**: 3 tasks (2 parallel)
- **Phase 5 (US3 — CI/CD)**: 7 tasks (5 parallel workflows, 2 sequential)
- **Phase 6 (US4 — Cost)**: 4 tasks (all parallel)
- **Phase 7 (US5 — Observability)**: 6 tasks (all parallel)
- **Phase 8 (Polish)**: 8 tasks (all parallel)

**Parallelizable Tasks**: 58 (80.6%)

**Estimated Effort**: 20-25 developer-days (with parallel execution, can complete in ~3 weeks with 2 engineers)

---

## Done When

- [ ] All 72 tasks completed
- [ ] Staging environment provisioned and operational (Quickstart Scenario 3 passes)
- [ ] Secrets initialized in AWS Secrets Manager (Quickstart Scenario 5 step 2 passes)
- [ ] Backend deployed to ECS Fargate via GitHub Actions (Quickstart Scenario 5 passes)
- [ ] Frontend deployed to S3+CloudFront via GitHub Actions (Quickstart Scenario 6 passes)
- [ ] Cost tracking enabled with budget alerts (Quickstart Scenario 10 passes)
- [ ] CloudWatch dashboards display infrastructure metrics (Quickstart Scenario 4 passes)
- [ ] Production infrastructure is IaC-defined (terraform plan succeeds) but NOT provisioned (Quickstart Scenario 9 passes)
- [ ] All Terraform code passes `terraform fmt -check`, `terraform validate`, `tflint`, and `tfsec`
- [ ] All GitHub Actions workflows execute successfully on test branches
- [ ] Infrastructure documentation complete (`infra/README.md` and module READMEs)
