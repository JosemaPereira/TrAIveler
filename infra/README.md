# infra

> Terraform infrastructure as code for TrAIveler — AWS resource provisioning and environment management.

This directory contains Terraform modules, environment configurations, and deployment automation
for the TrAIveler application infrastructure on AWS.

[← Back to root README](../README.md) | [Cloud Strategy](../docs/cloud-and-environments.md) | [Architecture](../docs/architecture.md)

---

> **Implementation Status** (module detail: [Project Structure](#project-structure); CI/CD detail:
> [CI/CD Integration](#cicd-integration)):
>
> - ✅ Foundation — backend config, version constraints, directory structure, CI workflow (Sprint 1, 2026-07-08)
> - ✅ VPC module (Sprint 3, 2026-07-11)
> - ✅ RDS module (Sprint 3, 2026-07-11)
> - ✅ ALB module (Sprint 3, 2026-07-11)
> - ✅ CloudFront module (Sprint 3, 2026-07-12)
> - ✅ Secrets module (Sprint 3, 2026-07-12) — see [Secrets Population](#secrets-population)
> - ✅ Remote-state bootstrap script (Sprint 3, 2026-07-12) — authored and test-verified only, not
>   run against a real AWS account (AWS-cost-avoidance policy; see [AWS Account Setup](#aws-account-setup-one-time))
> - ✅ ECS module (Sprint 3, 2026-07-12)
> - ✅ Root module wiring (Sprint 3, 2026-07-12) — `terraform validate` passes for both
>   environments; `plan`/`apply` still pending the remote-state backend (`003-T009`) and real
>   ECR/ACM values in the `.tfvars` files
> - ✅ Infra CI/CD workflows (Sprint 3, 2026-07-12) — `infra-plan.yml` + `infra-apply.yml`
>   implemented; every AWS-touching step is gated behind `AWS_ROLE_ARN` (not yet set), so both are
>   safe no-ops for now — see [CI/CD Integration](#cicd-integration)
> - 🔲 OIDC IAM role (Sprint 3, 2026-07-12) — documented and reproducible, deliberately not
>   provisioned yet (one-time manual AWS step, AWS-cost-avoidance policy) — see
>   [OIDC Authentication](#oidc-authentication)

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
│   │   ├── main.tf                   # ECS cluster (ARM64/Graviton2), task definition, service,
│   │   │                             # CPU-based target-tracking autoscaling, CloudWatch log group,
│   │   │                             # ALB-only security group on port 8080
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── rds/
│   │   ├── main.tf                   # RDS instance, subnet group, security group, Secrets Manager DB credentials
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── alb/
│   │   ├── main.tf                   # ALB, /healthz-checked target group, HTTP→HTTPS redirect, SSL
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── cloudfront/
│   │   ├── main.tf                   # CloudFront distribution (HTTPS-only viewer traffic, 404→
│   │   │                             # /index.html SPA rewrite), S3 origin bucket private except via OAI
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
├── main.tf                           # Root module: wires all 6 modules via outputs
│                                     # (vpc → alb → ecs → rds; cloudfront, secrets independent)
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

**Staging** (also automated by `infra-apply.yml` on merge to `main`, AWS-gated — see
[CI/CD Integration](#cicd-integration) below):

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

> **AWS_ROLE_ARN gating** (stated once here; referenced, not repeated, elsewhere in this file):
> every AWS-touching step in `infra-plan.yml`'s `plan` job and in `infra-apply.yml` checks
> `env.AWS_ROLE_ARN != ''`, mirrored from the `AWS_ROLE_ARN` repository secret that
> [OIDC Authentication](#oidc-authentication) step 4 creates. Until that secret is set for real —
> and the S3/DynamoDB remote-state backend (`003-T009`) exists — both workflows still run their
> non-AWS steps (validate, format-check, plan/apply authoring) but are safe no-ops on anything that
> would actually touch AWS. This is the repo's AWS-cost-avoidance policy in practice.

### GitHub Actions Workflows

Infrastructure changes are (or will be) deployed via GitHub Actions:

- **`.github/workflows/infra-plan.yml`** (`005-T107`, issue #91) — the `validate` and
  `format-check` jobs are fully real and run unconditionally on every pull request:
  `terraform init -backend=false` + `terraform validate` (needs no AWS credentials — it never
  touches the remote backend) and `terraform fmt -check -recursive`. The `plan` job runs real
  `terraform init` + `terraform plan -var-file=environments/{staging,production}.tfvars` and posts
  both plans as a PR comment (AWS-gated — see the note above). Job names
  (`Validate Terraform Configuration`, `Check Terraform Formatting`, `Terraform Plan (Staging)`) are
  pinned as required status checks in the branch-protection ruleset and must not be renamed.
- **`.github/workflows/infra-apply.yml`** (`005-T108`, issue #91) — triggers only on `push` to
  `main` touching `infra/**` (never on `pull_request` — this workflow is not a required PR status
  check, so path-filtering its trigger is safe). Runs `terraform apply -auto-approve
  -var-file=environments/staging.tfvars` for staging only — production is never auto-applied by this
  workflow, per the "Production" section above and the constitution's Environment Strategy — then
  exports `terraform output` values (`ecr_repository_url`, `ecs_cluster_name`, `ecs_service_name`,
  `s3_bucket_name`, `cloudfront_distribution_id`) to GitHub repository secrets via `gh secret set`
  for `backend-ci.yml`/`frontend-ci.yml` to consume in later sprints (AWS-gated — see the note above).

### OIDC Authentication

GitHub Actions authenticates to AWS using OIDC federation (no long-lived access keys).

**Status (`005-T109`, issue #92): documented, one-time manual AWS IAM step, deliberately NOT
Terraform-managed** — provisioning the OIDC provider/role via Terraform would create a
chicken-and-egg bootstrap problem (Terraform needs AWS credentials to create the very credentials
it would then use). None of the `aws iam` commands below have been run against a real AWS account
(AWS-cost-avoidance policy) — they are authored here so the step is reproducible whenever the user
decides to lift that constraint (same posture as `infra/scripts/bootstrap-state.sh`,
issue #96/`003-T009`: authored and documented, not executed). Step 4 below produces the
`AWS_ROLE_ARN` secret that both workflows key off via `aws-actions/configure-aws-credentials@v4` —
see the [gating note](#cicd-integration) at the top of CI/CD Integration.

#### 1. Create the IAM OIDC identity provider

Trusts GitHub's OIDC token issuer. One provider per AWS account, shared by every repository:

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

(The thumbprint above is GitHub's current OIDC root CA thumbprint. AWS's `configure-aws-credentials`
GitHub Action ecosystem generally recommends re-verifying it at setup time — AWS's own OIDC provider
support can also validate GitHub's certificate chain automatically, making an explicit thumbprint
unnecessary on recent AWS provider versions, but it's included here since it's still accepted and
explicit.)

#### 2. Create the IAM role with a repo-scoped trust policy

Scope the trust policy's `Condition` to this repository specifically — never a wildcard covering
every GitHub Actions run in the account. Narrowing to `main` (rather than `repo:...:*`) is preferred
since only `infra-apply.yml` (push-to-`main`-only) and PR-sourced `infra-plan.yml` runs need to
assume this role:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::<ACCOUNT_ID>:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": [
            "repo:JosemaPereira/TrAIveler:ref:refs/heads/main",
            "repo:JosemaPereira/TrAIveler:pull_request"
          ]
        }
      }
    }
  ]
}
```

```bash
aws iam create-role \
  --role-name traveler-github-actions-terraform \
  --assume-role-policy-document file://trust-policy.json \
  --description "OIDC role assumed by GitHub Actions (infra-plan.yml, infra-apply.yml) to run Terraform against AWS"
```

#### 3. Attach least-privilege permissions

Attach permissions scoped to the resource types Terraform actually manages in this repo — VPC, ECS,
RDS, ALB, CloudFront, Secrets Manager, S3, and DynamoDB (the last two for the remote-state backend
itself) — never `AdministratorAccess`. A reasonable starting sketch (review and tighten before real
use — this is deliberately broad-strokes, not a final least-privilege audit):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "TerraformManagedResources",
      "Effect": "Allow",
      "Action": [
        "ec2:*",
        "ecs:*",
        "rds:*",
        "elasticloadbalancing:*",
        "cloudfront:*",
        "secretsmanager:*",
        "logs:*",
        "application-autoscaling:*",
        "iam:GetRole",
        "iam:PassRole",
        "iam:CreateServiceLinkedRole"
      ],
      "Resource": "*"
    },
    {
      "Sid": "TerraformRemoteState",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:ListBucket",
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:DeleteItem"
      ],
      "Resource": [
        "arn:aws:s3:::traveler-terraform-state",
        "arn:aws:s3:::traveler-terraform-state/*",
        "arn:aws:dynamodb:us-east-1:<ACCOUNT_ID>:table/traveler-terraform-locks"
      ]
    }
  ]
}
```

```bash
aws iam put-role-policy \
  --role-name traveler-github-actions-terraform \
  --policy-name terraform-managed-resources \
  --policy-document file://terraform-permissions.json
```

#### 4. Store the role ARN as a GitHub repository secret

```bash
gh secret set AWS_ROLE_ARN --body "arn:aws:iam::<ACCOUNT_ID>:role/traveler-github-actions-terraform"
```

This is exactly the secret `infra-plan.yml`'s `plan` job and `infra-apply.yml`'s `apply` job check
for (`env.AWS_ROLE_ARN != ''`) before running any AWS-touching step.

#### 5. This is a one-time manual step, intentionally not performed in this session

None of the `aws iam` commands above have been executed against a real AWS account, per the
AWS-cost-avoidance policy noted throughout this file (`docs/roadmap.md`, `005-T108`'s row). Steps
1-4 are written to be reproducible whenever the user lifts that constraint, and double as the
reference for recreating the role if it's ever lost or needs rotating.

See [Cloud Strategy](../docs/cloud-and-environments.md) for the higher-level OIDC federation
decision record.

### Terraform Outputs for CI/CD

`infra/outputs.tf` now exists (`005-T104` under `G-SPRINT3-INFRA-ROOT-WIRING`, issue #90) and
`terraform validate` passes against it — but it has never been run against real infrastructure (no
`terraform apply` has happened yet), so there is no live state to query `terraform output` against.

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

`.github/workflows/infra-apply.yml` (`005-T108`) now exists and its "Export Terraform outputs to
GitHub Secrets" step runs exactly this `terraform output -raw <name>` + `gh secret set` sequence for
each value above (AWS-gated — see the [gating note](#cicd-integration) at the top of CI/CD
Integration).

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
