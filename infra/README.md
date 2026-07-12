# infra

> Terraform infrastructure as code for TrAIveler — AWS resource provisioning and environment management.

This directory contains Terraform modules, environment configurations, and deployment automation
for the TrAIveler application infrastructure on AWS.

[← Back to root README](../README.md) | [Cloud Strategy](../docs/cloud-and-environments.md) | [Architecture](../docs/architecture.md)

---

> **Implementation Status**: ✅ Foundation complete (Sprint 1, 2026-07-08) — Backend configuration, version constraints, directory structure, and CI workflow established. ✅ VPC module implemented (Sprint 3, 2026-07-11) — network isolation, public/private subnets, and NAT (instance/gateway) routing. ✅ RDS module implemented (Sprint 3, 2026-07-11) — PostgreSQL 15.4 instance on private subnets, ECS-only security group, and Secrets Manager–backed credentials. ✅ ALB module implemented (Sprint 3, 2026-07-11) — public-facing Application Load Balancer with HTTPS termination, HTTP→HTTPS redirect, and a `/healthz`-checked target group for ECS. ✅ CloudFront module implemented (Sprint 3, 2026-07-12) — private S3 origin bucket accessible only via a CloudFront Origin Access Identity, and a CloudFront distribution with HTTPS-only viewer traffic and a 404→`/index.html` rewrite for SPA client-side routing. ✅ Secrets module implemented (Sprint 3, 2026-07-12) — empty `ai_api_key` and `jwt_signing_key` Secrets Manager placeholders, populated manually post-apply (DB credentials remain owned by the RDS module). ✅ Remote-state bootstrap script implemented (Sprint 3, 2026-07-12) — `infra/scripts/bootstrap-state.sh` idempotently provisions the `traveler-terraform-state` S3 bucket and `traveler-terraform-locks` DynamoDB table; authored and test-verified only (mocked `aws` CLI), not executed against a real AWS account, per this repo's AWS-cost-avoidance policy. ✅ ECS module implemented (Sprint 3, 2026-07-12) — ARM64 (Graviton2) Fargate cluster, task definition, and service attached to the ALB target group, CPU-based target-tracking autoscaling, a dedicated CloudWatch log group, and an ALB-only security group on port 8080. ✅ Root module wiring implemented (Sprint 3, 2026-07-12) — `infra/main.tf` calls all six modules with dependency injection via module outputs (`vpc` → `alb` → `ecs` → `rds`, `cloudfront`/`secrets` independent), `infra/variables.tf`/`infra/outputs.tf` define the root interface, and `infra/environments/{staging,production}.tfvars` hold the environment-specific `.tfvars`; `terraform validate` passes for both environments. `terraform plan`/`apply` remain pending until the S3/DynamoDB remote-state resources actually exist (see `003-T009`) and real ECR/ACM resources replace the placeholder values in the `.tfvars` files.

---

## Responsibility

The infrastructure layer defines and provisions all AWS resources required to run TrAIveler in
staging and production environments. Its primary jobs are:

1. **Network isolation** — VPCs, subnets, security groups, and routing for staging and production
2. **Compute orchestration** — ECS Fargate clusters, task definitions, and auto-scaling
3. **Data persistence** — RDS PostgreSQL instances with automated backups
4. **Load balancing** — Application Load Balancers with HTTPS termination
5. **Content delivery** — S3 + CloudFront for frontend static assets
6. **Secrets management** — AWS Secrets Manager integration for credentials
7. **Observability** — CloudWatch logs, metrics, and alerting rules
8. **CI/CD integration** — GitHub Actions OIDC authentication and deployment automation

---

## Tech Stack

| Concern | Tool / Service |
|---------|----------------|
| IaC tool | Terraform 1.5+ (HCL syntax) |
| Cloud provider | AWS (us-east-1 region) |
| State management | S3 backend with DynamoDB locking |
| Compute | ECS Fargate (ARM64 Graviton2) |
| Database | Amazon RDS PostgreSQL 15.4 |
| Load balancing | Application Load Balancer (ALB) |
| CDN | Amazon CloudFront |
| Storage | Amazon S3 |
| Secrets | AWS Secrets Manager |
| Logging | Amazon CloudWatch Logs |
| CI/CD | GitHub Actions with OIDC federation |

---

## Project Structure

```
infra/
├── modules/                          # Reusable Terraform modules
│   ├── vpc/
│   │   ├── main.tf                   # VPC, subnets, NAT, routing
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── ecs/
│   │   ├── main.tf                   # ECS cluster, task definition, service, auto-scaling
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── rds/
│   │   ├── main.tf                   # RDS instance, subnet group, security group, Secrets Manager DB credentials
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── alb/
│   │   ├── main.tf                   # ALB, target group, listeners, SSL
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── cloudfront/
│   │   ├── main.tf                   # CloudFront distribution, S3 bucket, OAI
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── secrets/
│       ├── main.tf                   # Secrets Manager secrets for AI API key and JWT signing keys
│       │                             # (DB credentials are created directly by the rds module)
│       ├── variables.tf
│       └── outputs.tf
├── environments/
│   ├── staging.tfvars                # Staging-specific configuration (cost-optimized)
│   └── production.tfvars             # Production configuration (high-availability)
├── scripts/
│   ├── bootstrap-state.sh            # Idempotent remote-state (S3 + DynamoDB) bootstrap
│   └── bootstrap-state.test.sh       # Mocked-aws test suite for bootstrap-state.sh
├── main.tf                           # Root module calling reusable modules
├── variables.tf                      # Input variables (environment, sizing, domains, ECR/ACM references)
├── outputs.tf                        # Output values for CI/CD (ECR URL, ECS cluster & service, S3 bucket, CloudFront ID)
├── backend.tf                        # S3 + DynamoDB remote state configuration
└── versions.tf                       # Terraform and provider version constraints
```

---

## Prerequisites

| Tool | Version | Check |
|------|---------|-------|
| Terraform | ≥ 1.5 | `terraform version` |
| AWS CLI | ≥ 2.0 | `aws --version` |
| Valid AWS credentials | — | `aws sts get-caller-identity` |

### AWS Account Setup (One-Time)

Before running Terraform for the first time, create the remote state resources
(`traveler-terraform-state` S3 bucket and `traveler-terraform-locks` DynamoDB table) that
`infra/backend.tf` expects to already exist. The recommended path is the bootstrap script:

```bash
./infra/scripts/bootstrap-state.sh
```

It idempotently provisions both resources — safe to re-run, since each one is guarded by an
existence check (`aws s3api head-bucket` / `aws dynamodb describe-table`) and skipped if already
present:

- S3 bucket `traveler-terraform-state` (`us-east-1`) — versioning enabled, AES-256 server-side
  encryption, all four block-public-access settings on.
- DynamoDB table `traveler-terraform-locks` — partition key `LockID` (String), `PAY_PER_REQUEST`
  billing mode.

Its test suite (`infra/scripts/bootstrap-state.test.sh`) exercises both branches — resources
already present vs. missing — against a mocked `aws` CLI, without ever touching a real AWS account.

**Manual fallback** (raw AWS CLI commands, if you prefer not to run the script):

```bash
# 1. Create S3 bucket for Terraform state
aws s3api create-bucket \
  --bucket traveler-terraform-state \
  --region us-east-1

# 2. Enable versioning on the state bucket
aws s3api put-bucket-versioning \
  --bucket traveler-terraform-state \
  --versioning-configuration Status=Enabled

# 3. Create DynamoDB table for state locking
aws dynamodb create-table \
  --table-name traveler-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

---

## Environment Configurations

TrAIveler has **two environments** with different resource configurations:

### Staging (Active MVP)

- **Status**: Active, running all user-facing functionality
- **Cost Target**: $200/month
- **Configuration**:
  - VPC: 10.0.0.0/16
  - ECS: 0.25 vCPU / 0.5 GB RAM ARM64, auto-scaling 1-2 tasks
  - RDS: db.t4g.micro, single-AZ, 1-day backups
  - NAT: NAT instance (cost savings)
  - CloudWatch logs: 7-day retention

### Production (Dormant)

- **Status**: IaC-defined but not provisioned until alpha release
- **Cost Target**: $300-400/month (when active)
- **Configuration**:
  - VPC: 10.1.0.0/16 (non-overlapping with staging)
  - ECS: 1 vCPU / 2 GB RAM ARM64, auto-scaling 2-20 tasks
  - RDS: db.t4g.small, Multi-AZ, 30-day backups
  - NAT: NAT Gateway (production-grade)
  - CloudWatch logs: 30-day retention

---

## Terraform Workflow

### Initialize

```bash
cd infra
terraform init
```

This downloads provider plugins and configures the S3 backend for remote state.

### Validate Configuration

```bash
# Check syntax and validate configuration
terraform validate

# Format all .tf files
terraform fmt -recursive

# Check formatting (used in CI)
terraform fmt -check -recursive
```

### Plan Changes

```bash
# Preview changes for staging environment
terraform plan -var-file=environments/staging.tfvars

# Preview changes for production environment
terraform plan -var-file=environments/production.tfvars
```

### Apply Changes

**Staging** (target design: automated in CI on main branch merge — not yet wired; see
[CI/CD Integration](#cicd-integration) below for current status):

```bash
terraform apply -var-file=environments/staging.tfvars
```

**Production** (manual approval required):

```bash
# Production apply requires explicit confirmation
terraform apply -var-file=environments/production.tfvars
```

### Destroy Resources (Caution!)

```bash
# Only for staging environment during development
terraform destroy -var-file=environments/staging.tfvars
```

**Never destroy production without explicit team approval and data backup verification.**

---

## Secrets Population

Terraform creates AWS Secrets Manager secrets with **empty values**. Populate them manually after
the first `terraform apply`:

```bash
# Database credentials (auto-generated by RDS module, retrieve from Terraform output)
aws secretsmanager put-secret-value \
  --secret-id traveler-staging-db-credentials \
  --secret-string '{"username":"traiveler_admin","password":"GENERATE_SECURE_PASSWORD"}'

# Anthropic API key
aws secretsmanager put-secret-value \
  --secret-id traveler-staging-ai-api-key \
  --secret-string 'sk-ant-...'

# JWT signing keys (RS256 private key in PEM format)
# Generate with: openssl genrsa -out private.pem 2048
aws secretsmanager put-secret-value \
  --secret-id traveler-staging-jwt-signing-key \
  --secret-string file://private.pem
```

**Never commit secret values to Git.** Store them in a secure password manager and inject via CI/CD.

---

## CI/CD Integration

### GitHub Actions Workflows

Infrastructure changes are (or will be) deployed via GitHub Actions:

- **`.github/workflows/infra-plan.yml`** — exists and runs `terraform validate`/`fmt -check` on
  every pull request. Its `terraform plan` step is currently a placeholder — the root module
  (`infra/main.tf`, `G-SPRINT3-INFRA-ROOT-WIRING`, issue #90) now exists and `terraform validate`
  passes, but wiring a real `plan` with a PR-comment summary, plus giving the workflow real AWS
  credentials, is tracked separately as `005-T107` (issue #91, Backlog) — it also needs the real
  S3/DynamoDB backend to exist first (`003-T009`).
- **`.github/workflows/infra-apply.yml`** — **not created yet.** Planned as `005-T108`
  (issue #91, Backlog) to run `terraform apply -auto-approve` for staging on `main` merge. Per this
  repo's AWS-cost-avoidance policy, it must stay inert (or unmerged) even once authored, until infra
  is mature enough to allow real `apply` runs against AWS — see `docs/roadmap.md`'s note on
  `005-T108`.

### OIDC Authentication

GitHub Actions authenticates to AWS using OIDC federation (no long-lived access keys):

1. Create an IAM OIDC provider for GitHub in AWS Console
2. Create an IAM role with trust policy allowing GitHub Actions to assume it
3. Store the role ARN in GitHub repository secrets as `AWS_ROLE_ARN`

See [Cloud Strategy](../docs/cloud-and-environments.md) for detailed OIDC setup instructions.

### Terraform Outputs for CI/CD

`infra/outputs.tf` now exists (`005-T104` under `G-SPRINT3-INFRA-ROOT-WIRING`, issue #90) and
`terraform validate` passes against it — but it has never been run against real infrastructure
(no `terraform apply` has happened; see the AWS-cost-avoidance note above), so there is no live
state to query `terraform output` against yet.

Once infrastructure is applied, Terraform will export values needed by backend and frontend CI/CD:

```bash
# Export outputs for use in deployment workflows
terraform output -json > outputs.json

# Values exported:
# - ecr_repository_url (for Docker image push)
# - ecs_cluster_name (for service updates)
# - ecs_service_name (for rolling deployments)
# - s3_bucket_name (for frontend artifact sync)
# - cloudfront_distribution_id (for cache invalidation)
```

Once built, these outputs would be automatically exported to GitHub Secrets by the (not-yet-created)
`infra-apply.yml` workflow — see [GitHub Actions Workflows](#github-actions-workflows) above.

---

## Cost Monitoring

Track AWS spending by environment and service:

```bash
# View cost breakdown by tagged environment
aws ce get-cost-and-usage \
  --time-period Start=2026-07-01,End=2026-07-31 \
  --granularity MONTHLY \
  --metrics "UnblendedCost" \
  --group-by Type=TAG,Key=Environment

# Set budget alerts (one-time setup)
aws budgets create-budget \
  --account-id $(aws sts get-caller-identity --query Account --output text) \
  --budget file://budget.json \
  --notifications-with-subscribers file://budget-notifications.json
```

Budget alerts fire at 80% and 100% of the monthly threshold ($200 staging, $300-400 production).

---

## Troubleshooting

### State Lock Conflict

If `terraform apply` fails with a state lock error:

```bash
# List active locks
aws dynamodb scan \
  --table-name traveler-terraform-locks \
  --region us-east-1

# Force-unlock (only if you're certain no other operation is running)
terraform force-unlock <LOCK_ID>
```

### ECS Task Won't Start

```bash
# View ECS task events
aws ecs describe-services \
  --cluster traveler-staging-cluster \
  --services traveler-staging-service

# Check CloudWatch logs
aws logs tail /ecs/traveler-staging --follow
```

### RDS Connection Issues

```bash
# Test database connectivity from ECS task
aws ecs execute-command \
  --cluster traveler-staging-cluster \
  --task <TASK_ID> \
  --command "pg_isready -h <RDS_ENDPOINT> -U traiveler_admin"
```

---

## Related Documentation

| Document | Relevance |
|----------|-----------|
| [Cloud Strategy](../docs/cloud-and-environments.md) | Environment topology, cost strategy, CI/CD approach |
| [Architecture](../docs/architecture.md) | Component boundaries, integration rules, security boundaries |
| [Spec 003](../specs/003-cloud-env-strategy/spec.md) | Detailed cloud and IaC requirements |
| [Spec 005](../specs/005-system-architecture/spec.md) | Infrastructure module contracts and Terraform structure |
| [Tasks 005](../specs/005-system-architecture/tasks.md) | Infrastructure implementation tasks (Phase 5, T061-T109) |

---

## Next Steps

After infrastructure is provisioned:

1. **Populate secrets** in AWS Secrets Manager (database, API keys, JWT keys)
2. **Build and push** backend Docker image to ECR
3. **Deploy backend** via GitHub Actions (triggers ECS rolling update)
4. **Build and sync** frontend to S3, invalidate CloudFront cache
5. **Verify health** at ALB DNS name: `https://<alb-dns-name>/healthz`

See [backend/README.md](../backend/README.md) and [frontend/README.md](../frontend/README.md) for
deployment commands.
