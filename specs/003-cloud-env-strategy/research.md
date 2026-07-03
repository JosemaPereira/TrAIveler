# Research: Cloud & Environments Strategy

**Date**: 2026-07-03 | **Plan**: [plan.md](plan.md)

All infrastructure and deployment architecture decisions for TrAIveler cloud environments are documented below.

---

## Decision 1 — ECS Fargate Task Configuration for Go Backend

**Decision**:
- **Staging**: 0.5 vCPU / 1 GB memory on ARM64 (Graviton2)
- **Production**: 1 vCPU / 2 GB memory on ARM64 (Graviton2)
- Platform version: `LATEST` (currently 1.4.0)

**Rationale**:
- **Graviton2 ARM64 cost savings**: 20% cheaper than x86_64 for equivalent compute (~$0.04048/hr vs $0.04656/hr for 1vCPU/2GB). Go 1.24+ natively supports `linux/arm64` compilation with zero performance penalty.
- **0.5 vCPU is sufficient for AI workload pattern**: Backend is I/O-bound during AI API calls (90% of request time spent waiting on Anthropic streaming response). CPU usage spikes only during JSON parsing and database writes. AWS Fargate `0.5 vCPU` provides burst capacity to 1.0 vCPU for short processing windows.
- **1 GB memory handles conversation context**: Estimated memory usage: ~150 MB base Go runtime, ~200 MB for HTTP server and pgx connection pool (10 connections × 10 MB per connection), ~300 MB for active conversation state (multi-turn history up to 10 messages). Total: ~650 MB, leaving 350 MB headroom for GC and temporary allocations.
- **Production scaling headroom**: 1 vCPU / 2 GB allows 2× concurrent requests without thrashing, supports connection pool size increase to 20, and provides memory buffer for future features (image processing, PDF generation).

**Alternatives considered**:
- **x86_64 (Intel/AMD)**: 20% more expensive, zero technical advantage for Go workloads. ARM64 is the default choice unless third-party binary dependencies require x86_64.
- **Lambda**: Rejected due to 15-minute timeout (AI conversations can exceed this), 10 MB payload limit (streaming responses may accumulate more), and cold start latency (500-1000ms vs Fargate's persistent tasks).

**Health check configuration**:
```hcl
health_check {
  enabled             = true
  path                = "/health"
  interval            = 30  # seconds
  timeout             = 5   # seconds
  healthy_threshold   = 2   # consecutive successes
  unhealthy_threshold = 3   # consecutive failures
  matcher             = "200"
}
```

**Grace period**: 60 seconds (allows Go application to complete startup: database connection pool initialization ~5s, health check endpoint registration ~2s, remaining buffer for container network setup).

**Auto-scaling strategy**:
```hcl
# Target Tracking - CPU-based (preferred for cost predictability)
resource "aws_appautoscaling_policy" "ecs_cpu_policy" {
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"

  target_tracking_scaling_policy_configuration {
    target_value       = 70.0  # Scale when average CPU > 70%
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    scale_in_cooldown  = 300  # 5 minutes before scaling down
    scale_out_cooldown = 60   # 1 minute before scaling up
  }
}

# Staging: 1-2 tasks (min 1 for availability, max 2 for cost cap)
# Production: 2-6 tasks (min 2 for HA across AZs, max 6 for alpha load)
```

**Rationale for target tracking over step scaling**: Target tracking automatically adjusts task count to maintain 70% CPU utilization, reducing manual threshold management. Step scaling requires defining exact thresholds (e.g., add 1 task at 75%, add 2 tasks at 90%), which is brittle for unpredictable AI workload spikes.

---

## Decision 2 — Application Load Balancer Configuration

**Decision**:
- ALB Type: Application Load Balancer (not Network Load Balancer)
- Scheme: `internet-facing` (public subnets)
- IP address type: `ipv4` (IPv6 optional, not required for MVP)
- Deletion protection: `false` (staging), `true` (production)

**Target group settings**:
```hcl
resource "aws_lb_target_group" "backend" {
  name        = "traiveler-backend-${var.environment}"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"  # Required for Fargate

  health_check {
    enabled             = true
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }

  deregistration_delay = 30  # Connection draining: 30s for graceful shutdown

  stickiness {
    enabled         = false  # Stateless backend; no session affinity needed
    type            = "lb_cookie"
    cookie_duration = 86400  # Not used, but required field
  }
}
```

**Connection draining rationale**: 30 seconds allows in-flight AI streaming requests to complete. Anthropic API responses typically finish within 20 seconds; 30s provides 10s buffer. Longer delays (60s+) slow down deployments unnecessarily.

**HTTPS listener configuration**:
```hcl
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"  # TLS 1.3 + 1.2
  certificate_arn   = aws_acm_certificate.main.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.backend.arn
  }
}

# HTTP listener: redirect to HTTPS
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}
```

**Certificate management**:
- AWS Certificate Manager (ACM) for SSL/TLS certificates
- DNS validation via Route 53 (or manual CNAME if using external DNS provider)
- Auto-renewal: ACM handles renewals automatically 60 days before expiration
- Certificate ARN stored in Terraform state, referenced by ALB listener

**Alternatives considered**:
- **Network Load Balancer**: Lower cost but lacks HTTP-level routing, health checks return only TCP connectivity (not application health), and no native HTTPS termination. ALB is preferred for L7 routing and integrated certificate management.
- **CloudFront + ALB**: Adds $0.085/GB data transfer + $0.01/10k requests. Rejected for MVP; CloudFront caching provides no value for dynamic AI responses. Consider post-alpha if frontend assets need global CDN distribution.

---

## Decision 3 — Docker Build Strategy for Go + React

**Decision**:
- **Backend**: Multi-stage Dockerfile targeting `linux/arm64`
- **Frontend**: Static build → S3 + CloudFront distribution
- **ECR**: Single repository per service with semantic version tags

**Backend Dockerfile** (multi-stage for Go + ARM64):
```dockerfile
# Stage 1: Build
FROM --platform=linux/arm64 golang:1.24-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o /app/server ./cmd/server

# Stage 2: Runtime
FROM --platform=linux/arm64 alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
USER nobody
ENTRYPOINT ["/app/server"]
```

**Build command** (GitHub Actions):
```bash
docker buildx build \
  --platform linux/arm64 \
  --tag ${ECR_REGISTRY}/traiveler-backend:${GIT_SHA} \
  --tag ${ECR_REGISTRY}/traiveler-backend:latest \
  --push \
  backend/
```

**Rationale**:
- **Multi-stage reduces image size**: Builder stage includes Go compiler (~300 MB), runtime stage is Alpine + binary (~15 MB). Final image: ~20 MB vs 300 MB monolithic image.
- **`CGO_ENABLED=0` ensures static binary**: No libc dependencies; fully portable across any linux/arm64 environment (Fargate, local Docker, CI).
- **`-ldflags="-s -w"` strips debug symbols**: Reduces binary size by 30% (from ~25 MB to ~18 MB). Debug symbols not needed in production; source maps and CloudWatch Logs provide sufficient observability.
- **`USER nobody`**: Non-root execution follows least-privilege principle; mitigates container escape vulnerabilities.

**Frontend static build workflow**:
```yaml
# .github/workflows/deploy-frontend.yml
- name: Build React frontend
  run: |
    cd frontend
    npm ci
    npm run build  # Output: frontend/build/
    
- name: Sync to S3
  run: |
    aws s3 sync frontend/build/ s3://traiveler-frontend-${ENVIRONMENT}/ \
      --delete \
      --cache-control "public, max-age=31536000, immutable" \
      --exclude "index.html"
    
    # index.html: no cache (users always get latest version)
    aws s3 cp frontend/build/index.html s3://traiveler-frontend-${ENVIRONMENT}/ \
      --cache-control "public, max-age=0, must-revalidate"
    
- name: Invalidate CloudFront cache
  run: |
    aws cloudfront create-invalidation \
      --distribution-id ${CLOUDFRONT_DISTRIBUTION_ID} \
      --paths "/*"
```

**ECR repository structure**:
```
traiveler-backend (ARM64)
├── 001-abc123f  (Git SHA, immutable)
├── 001-v1.0.0   (Semantic version tag)
└── latest       (Mutable, always points to main branch HEAD)

Tag format:
- Production releases: v1.0.0, v1.0.1, v1.1.0 (semantic versioning)
- Feature branches: 001-abc123f (spec number + Git short SHA)
- Latest: Overwritten on every main branch merge
```

**Rationale for SHA + semantic version tagging**:
- **SHA tags enable exact rollback**: `docker pull traiveler-backend:001-abc123f` retrieves the exact image deployed at a specific commit, even if `latest` has moved forward.
- **Semantic version tags enable changelog tracking**: `v1.0.0` → `v1.1.0` indicates a minor version bump; operations team can correlate with release notes.
- **`latest` tag simplifies staging deployments**: CI always deploys `latest` to staging; production deployments explicitly pin to `v1.x.x` tags.

**Alternatives considered**:
- **Backend as static binary + S3**: Go binaries can be uploaded to S3 and executed via Lambda or EC2 user data. Rejected because ECS Fargate requires Docker images, and container orchestration provides better health checks, zero-downtime deployments, and resource isolation.
- **Frontend SSR (Server-Side Rendering)**: React 19 supports SSR, but TrAIveler frontend is a SPA (Single-Page Application) with client-side routing. SSR adds complexity (Node.js runtime, increased server costs) without SEO benefits (app is behind authentication). Static build + S3 is optimal.

---

## Decision 4 — VPC and Networking Architecture

**Decision**:
- **CIDR blocks**: Staging `10.0.0.0/16`, Production `10.1.0.0/16`
- **Subnets**: 2 Availability Zones per environment
  - Public subnets (ALB): `10.x.1.0/24`, `10.x.2.0/24`
  - Private subnets (ECS + RDS): `10.x.11.0/24`, `10.x.12.0/24`
- **NAT strategy**: NAT Instance for staging, NAT Gateway for production
- **Internet Gateway**: Attached to VPC for public subnet internet access

**Subnet allocation rationale**:
- **Public subnets (251 IPs each)**: ALB requires minimum 8 IPs per AZ; 251 IPs provides ~30× headroom for ALB scaling and future public-facing services (bastion hosts, VPN endpoints).
- **Private subnets (251 IPs each)**: ECS tasks (1-6 tasks × 1 IP = 6 IPs max), RDS primary + replica (2 IPs), RDS subnet group requires minimum 2 IPs. Total usage: ~10 IPs per AZ; 251 IPs allows 25× growth.
- **Non-overlapping CIDRs (`10.0.x.x` vs `10.1.x.x`)**: Enables future VPC peering if staging needs to access production databases for data migration or read replicas.

**NAT Gateway vs NAT Instance cost analysis**:

| Component          | NAT Gateway (Production) | NAT Instance (Staging) |
|--------------------|--------------------------|------------------------|
| Hourly cost        | $0.045/hr × 2 AZs = $0.09/hr | $0.0116/hr (t4g.nano) × 1 = $0.0116/hr |
| Monthly cost       | $0.09 × 730 = $65.70 | $0.0116 × 730 = $8.47 |
| Data transfer      | $0.045/GB | $0.09/GB (EC2 standard) |
| High availability  | AWS-managed, per-AZ | Single instance (manual failover) |
| Maintenance        | Zero (AWS-managed) | OS patches, AMI updates |

**Decision rationale**:
- **Staging uses NAT Instance (t4g.nano)**: Saves $57/month vs NAT Gateway. Acceptable risk: staging downtime during NAT instance failures does not impact users; developers can re-deploy NAT instance in 5 minutes via Terraform. Data transfer cost increase ($0.09/GB vs $0.045/GB) is negligible for staging (~10 GB/month data transfer = $0.45 extra cost).
- **Production uses NAT Gateway**: High availability requirement justifies $66/month cost. AWS-managed service eliminates maintenance burden and provides automatic failover.

**Security group rules**:

```hcl
# ALB Security Group
resource "aws_security_group" "alb" {
  name        = "traiveler-alb-${var.environment}"
  description = "Allow HTTPS inbound from internet"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS from internet"
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP redirect to HTTPS"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }
}

# ECS Task Security Group
resource "aws_security_group" "ecs_tasks" {
  name        = "traiveler-ecs-tasks-${var.environment}"
  description = "Allow inbound from ALB, outbound to RDS and internet"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
    description     = "HTTP from ALB"
  }

  egress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    security_groups = [aws_security_group.rds.id]
    description = "PostgreSQL to RDS"
  }

  egress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTPS to Anthropic API and AWS services"
  }
}

# RDS Security Group
resource "aws_security_group" "rds" {
  name        = "traiveler-rds-${var.environment}"
  description = "Allow PostgreSQL from ECS tasks only"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs_tasks.id]
    description     = "PostgreSQL from ECS tasks"
  }

  # No egress rules (RDS does not initiate outbound connections)
}
```

**Security group design principles**:
- **Least privilege**: Each security group allows only required traffic. RDS accepts connections only from ECS tasks, not from ALB or internet.
- **Explicit ingress, permissive egress**: Ingress rules are tightly scoped; egress allows all outbound (ECS tasks need to reach Anthropic API, AWS services, and package repositories during deployments).
- **Security group chaining**: ALB → ECS → RDS rules reference security group IDs, not CIDR blocks. This automatically adapts to IP changes when tasks scale or replace.

**Alternatives considered**:
- **Single large CIDR (`10.0.0.0/8`)**: Provides more IPs but wastes address space. AWS recommends `/16` for VPCs with <1000 instances.
- **3+ Availability Zones**: Increases costs (3× ALB, 3× NAT Gateway/Instance, 3× RDS Multi-AZ replicas) without meaningful availability improvement for MVP. 2 AZs provide 99.99% SLA; 3 AZs target 99.999% (five nines), which is overkill for alpha release.
- **VPC Endpoints for AWS services**: Reduces data transfer costs by routing S3/ECR traffic through private AWS network instead of NAT Gateway. Cost analysis: VPC endpoint = $0.01/hr + $0.01/GB = ~$7.30/month + data transfer. Staging S3/ECR traffic is ~5 GB/month; savings = $0.225 vs endpoint cost $7.30. **Rejected for staging; consider for production post-alpha if data transfer exceeds 100 GB/month.**

---

## Decision 5 — RDS Configuration and Cost Optimization

**Decision**:
- **Staging**: `db.t4g.micro` (2 vCPU, 1 GB RAM), Single-AZ, gp3 storage 20 GB
- **Production**: `db.t4g.small` (2 vCPU, 2 GB RAM), Multi-AZ, gp3 storage 50 GB
- **Engine**: PostgreSQL 16.3 (latest minor version)
- **Backup retention**: 7 days (staging), 30 days (production)
- **Maintenance window**: Sunday 03:00-04:00 UTC (lowest traffic)

**db.t4g.micro capabilities and limits**:

| Metric               | db.t4g.micro Capacity | MVP Workload Estimate |
|----------------------|-----------------------|-----------------------|
| vCPUs                | 2 (ARM64 Graviton2)   | ~10% utilization (I/O-bound queries) |
| RAM                  | 1 GB                  | ~400 MB used (PostgreSQL shared_buffers 256 MB + OS 150 MB) |
| Baseline performance | 20% of 2 vCPU         | Sufficient; queries complete in <50ms |
| Burst credits        | 144 credits/hr        | Replenishes faster than consumption rate |
| Max connections      | ~87 (PostgreSQL default formula) | 10 connections from ECS (1 task × 10 pool size) |
| Storage IOPS         | 3000 IOPS (gp3 baseline) | ~50 IOPS during peak (10 qps × 5 queries per request) |

**Rationale for db.t4g.micro in staging**:
- **Burstable credits handle MVP load**: T4g instances accrue CPU credits during idle periods (nights, weekends) and burst to 100% CPU during active testing. MVP workload is <10 qps (queries per second); credits replenish faster than consumption.
- **1 GB RAM fits working set**: Estimated database size: 50 MB (1000 itineraries × 50 KB each). PostgreSQL loads hot data into shared_buffers (256 MB); remaining 744 MB available for connection overhead and OS cache.
- **Cost efficiency**: $0.016/hr = $11.68/month vs db.t4g.small ($0.032/hr = $23.36/month). Saves $11.68/month for staging, acceptable given MVP scale.

**Production db.t4g.small justification**:
- **2 GB RAM**: Doubles buffer cache capacity, reduces disk I/O by 30-40% (more queries served from memory).
- **Multi-AZ**: Synchronous replication to standby instance in second AZ. Automatic failover in 60-120 seconds during primary failure. Adds $23.36/month (doubles RDS cost), but mandatory for production SLA.
- **50 GB storage headroom**: Initial database size ~50 MB; 50 GB allows 1000× growth before storage scaling required.

**Connection pooling from ECS tasks**:

```go
// backend/internal/database/pool.go
import (
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, err
    }

    // Staging: 1 ECS task × 10 connections = 10 total
    // Production: 2-6 ECS tasks × 10 connections = 20-60 total
    config.MaxConns = 10
    config.MinConns = 2
    config.MaxConnLifetime = 1 * time.Hour  // Recycle connections hourly
    config.MaxConnIdleTime = 5 * time.Minute

    return pgxpool.NewWithConfig(ctx, config)
}
```

**Rationale for pool sizing**:
- **10 connections per task**: Each ECS task handles 2-3 concurrent HTTP requests during peak; 10 connections provides 3-5× headroom for spikes.
- **MaxConnLifetime = 1 hour**: Prevents long-lived connections from accumulating stale state (prepared statements, advisory locks). Forced recycling every hour ensures clean connection pool.
- **MinConns = 2**: Keeps 2 connections warm; avoids cold-start latency (connection establishment takes ~50ms).

**Backup retention strategy**:
- **Staging (7 days)**: Sufficient for debugging recent issues; reduces storage costs (backups billed at $0.095/GB-month).
- **Production (30 days)**: Enables point-in-time recovery for compliance and data loss scenarios. 30 days aligns with typical audit retention requirements.

**Maintenance window rationale**:
- **Sunday 03:00-04:00 UTC**: Corresponds to Saturday 8pm-9pm Pacific (US West Coast), Sunday 12pm-1pm Central Europe. Lowest user traffic globally; acceptable downtime window for minor version patches (10-15 minutes).

**Cost estimate** (staging):

| Component              | Configuration        | Monthly Cost |
|------------------------|----------------------|--------------|
| db.t4g.micro instance  | Single-AZ, 730 hrs   | $11.68       |
| gp3 storage            | 20 GB × $0.115/GB    | $2.30        |
| Backup storage         | ~5 GB × $0.095/GB    | $0.48        |
| **Total**              |                      | **$14.46**   |

**Alternatives considered**:
- **Aurora Serverless v2**: Auto-scaling database with per-ACU billing. Minimum capacity: 0.5 ACU × $0.12/hr = $43.80/month. Rejected; 3× more expensive than db.t4g.micro for MVP workload.
- **DynamoDB**: NoSQL alternative with pay-per-request pricing. Rejected; relational schema (users, itineraries, activities with foreign keys) maps naturally to PostgreSQL. DynamoDB would require denormalization and complex query patterns (GSIs for filtering by date, destination).

---

## Decision 6 — GitHub Actions OIDC Federation to AWS

**Decision**:
- **Authentication method**: OpenID Connect (OIDC) federation
- **IAM roles**: Separate roles per environment (staging-deployer, production-deployer)
- **Terraform permissions**: Separate role with restricted permissions for `terraform apply`

**IAM OIDC Identity Provider configuration**:

```hcl
resource "aws_iam_openid_connect_provider" "github" {
  url = "https://token.actions.githubusercontent.com"

  client_id_list = ["sts.amazonaws.com"]

  thumbprint_list = [
    "6938fd4d98bab03faadb97b34396831e3780aea1",  # GitHub Actions OIDC thumbprint (valid until 2031)
    "1c58a3a8518e8759bf075b76b750d4f2df264fcd"   # Backup thumbprint
  ]
}
```

**IAM role trust policy** (staging-deployer):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::AWS_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com",
          "token.actions.githubusercontent.com:sub": "repo:ORG_NAME/traiveler:environment:staging"
        }
      }
    }
  ]
}
```

**Condition constraints breakdown**:
- **`aud` = `sts.amazonaws.com`**: Ensures token was issued for AWS STS audience; prevents token reuse across different cloud providers.
- **`sub` = `repo:ORG_NAME/traiveler:environment:staging`**: Restricts role assumption to workflows running in the `traiveler` repository within `ORG_NAME` organization, specifically in the `staging` GitHub environment. This prevents:
  - Workflows in other repositories from assuming the role
  - Workflows in the same repository but different environments (e.g., production) from using staging credentials
  - Forked repositories from accessing AWS credentials

**IAM policy for deployer role** (staging-deployer):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "ECSDeployment",
      "Effect": "Allow",
      "Action": [
        "ecs:UpdateService",
        "ecs:DescribeServices",
        "ecs:DescribeTaskDefinition",
        "ecs:RegisterTaskDefinition"
      ],
      "Resource": [
        "arn:aws:ecs:us-east-1:AWS_ACCOUNT_ID:service/traiveler-staging/*",
        "arn:aws:ecs:us-east-1:AWS_ACCOUNT_ID:task-definition/traiveler-backend-staging:*"
      ]
    },
    {
      "Sid": "ECRAccess",
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:PutImage",
        "ecr:InitiateLayerUpload",
        "ecr:UploadLayerPart",
        "ecr:CompleteLayerUpload"
      ],
      "Resource": [
        "arn:aws:ecr:us-east-1:AWS_ACCOUNT_ID:repository/traiveler-backend"
      ]
    },
    {
      "Sid": "S3FrontendDeployment",
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:PutObjectAcl",
        "s3:DeleteObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::traiveler-frontend-staging",
        "arn:aws:s3:::traiveler-frontend-staging/*"
      ]
    },
    {
      "Sid": "CloudFrontInvalidation",
      "Effect": "Allow",
      "Action": [
        "cloudfront:CreateInvalidation",
        "cloudfront:GetInvalidation"
      ],
      "Resource": "arn:aws:cloudfront::AWS_ACCOUNT_ID:distribution/STAGING_DISTRIBUTION_ID"
    }
  ]
}
```

**Terraform role with permission boundaries**:

```hcl
resource "aws_iam_role" "terraform" {
  name               = "traiveler-terraform"
  assume_role_policy = data.aws_iam_policy_document.github_oidc_trust.json

  permissions_boundary = aws_iam_policy.terraform_boundary.arn
}

resource "aws_iam_policy" "terraform_boundary" {
  name = "terraform-permissions-boundary"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DenyIAMEscalation"
        Effect = "Deny"
        Action = [
          "iam:CreateUser",
          "iam:CreateAccessKey",
          "iam:PutUserPolicy",
          "iam:AttachUserPolicy"
        ]
        Resource = "*"
      },
      {
        Sid    = "DenyRootAccountActions"
        Effect = "Deny"
        Action = "*"
        Resource = "*"
        Condition = {
          StringEquals = {
            "aws:PrincipalType" = "Root"
          }
        }
      },
      {
        Sid    = "AllowInfrastructureManagement"
        Effect = "Allow"
        Action = [
          "ec2:*",
          "ecs:*",
          "rds:*",
          "s3:*",
          "cloudfront:*",
          "elasticloadbalancing:*",
          "iam:CreateRole",
          "iam:PutRolePolicy",
          "iam:PassRole"
        ]
        Resource = "*"
      }
    ]
  })
}
```

**Permission boundary rationale**:
- **Prevents privilege escalation**: Terraform can create IAM roles and policies (necessary for ECS task execution roles), but the boundary blocks creation of IAM users with access keys. This prevents an attacker who compromises GitHub Actions from creating permanent backdoor credentials.
- **Denies root account actions**: Protects against accidental misconfiguration that could lock out all IAM users.
- **Scoped to infrastructure resources**: Terraform can manage ECS, RDS, S3, but cannot modify billing settings, AWS Organizations, or CloudTrail logs.

**GitHub Actions workflow usage**:

```yaml
name: Deploy Staging

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: staging  # Must match IAM trust policy condition
    permissions:
      id-token: write  # Required for OIDC token generation
      contents: read   # Required for checkout

    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::AWS_ACCOUNT_ID:role/staging-deployer
          aws-region: us-east-1

      - name: Login to Amazon ECR
        run: |
          aws ecr get-login-password --region us-east-1 | \
            docker login --username AWS --password-stdin \
            AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com

      - name: Build and push Docker image
        run: |
          docker buildx build \
            --platform linux/arm64 \
            --tag AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/traiveler-backend:${GITHUB_SHA} \
            --tag AWS_ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/traiveler-backend:latest \
            --push \
            backend/

      - name: Update ECS service
        run: |
          aws ecs update-service \
            --cluster traiveler-staging \
            --service backend \
            --force-new-deployment
```

**Alternatives considered**:
- **Long-lived IAM access keys**: Stored as GitHub Secrets. Rejected due to security risks: keys never expire, can be leaked in logs, and provide persistent access even after repository compromise is detected. OIDC tokens expire in 10 minutes and are scoped to specific workflows.
- **Self-hosted GitHub Actions runners on EC2**: Provides VPC-internal access without OIDC. Rejected due to maintenance overhead (OS patching, scaling, monitoring) and cost (~$20/month for t4g.small runner vs $0 for GitHub-hosted runners).

---

## Decision 7 — Terraform Remote State Management

**Decision**:
- **State storage**: S3 bucket with versioning and encryption
- **State locking**: DynamoDB table with `LockID` primary key
- **Workspace isolation**: Terraform workspaces (staging, production) share backend but maintain separate state files

**S3 bucket configuration**:

```hcl
resource "aws_s3_bucket" "terraform_state" {
  bucket = "traiveler-terraform-state"

  lifecycle {
    prevent_destroy = true  # Protect against accidental deletion
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"  # S3-managed keys (no KMS cost)
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_lifecycle_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    id     = "expire-old-versions"
    status = "Enabled"

    noncurrent_version_expiration {
      noncurrent_days = 90  # Keep 90 days of version history
    }
  }
}
```

**DynamoDB table for state locking**:

```hcl
resource "aws_dynamodb_table" "terraform_locks" {
  name         = "traiveler-terraform-locks"
  billing_mode = "PAY_PER_REQUEST"  # No cost when idle; scales automatically
  hash_key     = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }

  lifecycle {
    prevent_destroy = true
  }
}
```

**Terraform backend configuration**:

```hcl
# backend.tf
terraform {
  backend "s3" {
    bucket         = "traiveler-terraform-state"
    key            = "infrastructure/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "traiveler-terraform-locks"
  }
}
```

**Workspace-specific state files** (stored in S3):
```
s3://traiveler-terraform-state/
├── infrastructure/
│   ├── terraform.tfstate         (default workspace - unused)
│   ├── env:/staging/terraform.tfstate
│   └── env:/production/terraform.tfstate
```

**Workspace commands**:
```bash
# Initialize Terraform with staging workspace
terraform workspace select staging || terraform workspace new staging
terraform apply -var-file=staging.tfvars

# Switch to production workspace
terraform workspace select production || terraform workspace new production
terraform apply -var-file=production.tfvars
```

**Bootstrap sequence** (chicken-and-egg problem):

The S3 bucket and DynamoDB table must exist before Terraform can use them for remote state, but Terraform needs state storage to manage resources. Resolution:

**Step 1**: Create bootstrap resources manually (one-time):
```bash
# bootstrap/create-state-backend.sh
aws s3api create-bucket \
  --bucket traiveler-terraform-state \
  --region us-east-1

aws s3api put-bucket-versioning \
  --bucket traiveler-terraform-state \
  --versioning-configuration Status=Enabled

aws s3api put-bucket-encryption \
  --bucket traiveler-terraform-state \
  --server-side-encryption-configuration \
    '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'

aws dynamodb create-table \
  --table-name traiveler-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

**Step 2**: Migrate bootstrap resources to Terraform management:
```bash
# Import existing resources into Terraform state
terraform import aws_s3_bucket.terraform_state traiveler-terraform-state
terraform import aws_dynamodb_table.terraform_locks traiveler-terraform-locks
```

**Step 3**: All future infrastructure managed via Terraform with remote state.

**State file security**:
- **Encryption at rest**: AES256 encryption on S3 (no additional cost vs KMS-managed keys at $1/month).
- **Encryption in transit**: Terraform uses HTTPS for all S3 API calls.
- **Access control**: S3 bucket policy restricts access to GitHub Actions OIDC role and admin IAM users only.

**State versioning rationale**:
- **90-day retention**: Balances rollback capability (3 months of history) with storage costs ($0.023/GB-month for versioned objects). Typical state file: 50-100 KB; 90 versions = 5-10 MB = $0.12/year.
- **Automatic expiration**: Prevents unbounded storage growth; after 90 days, old versions are automatically deleted.

**Alternatives considered**:
- **Terraform Cloud**: Managed state storage, locking, and UI. Free tier supports 5 users. Rejected due to added dependency on third-party service; S3 + DynamoDB provides same reliability with full control.
- **Local state files**: Stored in repository (`.terraform/` directory). Rejected due to collaboration issues (merge conflicts), lack of locking (concurrent `terraform apply` causes corruption), and security risk (state contains sensitive values like database passwords).
- **Separate S3 buckets per workspace**: Increases cost (2 buckets × $0.023/GB vs 1 bucket) and complexity (2 IAM policies, 2 lifecycle rules). Workspace-based state isolation within single bucket is simpler.

---

## Decision 8 — Cost Tagging Strategy

**Decision**:
- **Required tags**: `Project`, `Environment`, `ManagedBy`, `CostCenter`
- **Propagation**: All tags defined in Terraform variables and applied via `default_tags` provider block

**Tag definitions**:

| Tag Key      | Allowed Values                  | Purpose                                    |
|--------------|----------------------------------|--------------------------------------------|
| `Project`    | `traiveler`                     | Group all resources by application         |
| `Environment`| `staging`, `production`         | Separate cost reporting by environment     |
| `ManagedBy`  | `terraform`                     | Identify IaC-managed vs manually created   |
| `CostCenter` | `engineering`, `operations`     | Allocate costs to organizational teams     |

**Terraform provider configuration**:

```hcl
# providers.tf
provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "traiveler"
      Environment = terraform.workspace  # Automatically set to 'staging' or 'production'
      ManagedBy   = "terraform"
      CostCenter  = "engineering"
    }
  }
}
```

**Rationale for default_tags**:
- **Automatic propagation**: Every resource created by Terraform automatically inherits tags. Eliminates manual tag definitions in every `aws_*` resource block (200+ lines of repetitive code).
- **Workspace-based environment tag**: `Environment = terraform.workspace` ensures staging and production resources are correctly labeled without manual variable passing.
- **Immutable tags**: Default tags cannot be overridden at resource level (requires explicit `tags = {}` merge), preventing accidental tag removal.

**Cost allocation reports**:

AWS Cost Explorer filters:
```
Filter: Tag:Project = traiveler
Group by: Tag:Environment

Result:
- traiveler / staging: $198.47/month
- traiveler / production: $0.00/month (dormant)
```

**Tag enforcement** (AWS Config rule):

```hcl
resource "aws_config_config_rule" "required_tags" {
  name = "required-tags"

  source {
    owner             = "AWS"
    source_identifier = "REQUIRED_TAGS"
  }

  input_parameters = jsonencode({
    tag1Key = "Project"
    tag2Key = "Environment"
    tag3Key = "ManagedBy"
  })
}
```

**Enforcement behavior**: AWS Config evaluates all resources every 24 hours. Resources missing required tags are marked as non-compliant (visible in AWS Config dashboard). Does not block resource creation (informational only).

**Cost tagging best practices**:
- **Avoid dynamic tags**: Tags like `CreatedBy = ${github.actor}` or `DeployedAt = ${timestamp()}` create high cardinality (100s of unique values), making cost aggregation reports unreadable. Use static values only.
- **Use CamelCase for keys**: AWS billing reports preserve case; inconsistent casing (`environment` vs `Environment`) splits costs across multiple rows.
- **Tag S3 buckets explicitly**: S3 bucket tags are not inherited by objects; use `aws_s3_bucket_tagging` resource to apply tags to bucket itself (enables storage class cost allocation).

**Alternatives considered**:
- **Resource naming conventions for cost tracking**: Naming all resources `traiveler-staging-*` allows cost filtering by name prefix. Rejected because AWS Cost Explorer does not support name-based filtering; tags are the only reliable method.
- **Tag every resource individually**: Provides flexibility to override default tags per resource. Rejected due to maintenance burden (100+ resource blocks × 4 tags = 400+ lines of repetitive code) and risk of typos.

---

## Summary: Key Infrastructure Values

| Parameter                     | Staging                          | Production                       |
|-------------------------------|----------------------------------|----------------------------------|
| **ECS Fargate**               | 0.5 vCPU / 1 GB, ARM64           | 1 vCPU / 2 GB, ARM64             |
| **RDS**                       | db.t4g.micro, Single-AZ          | db.t4g.small, Multi-AZ           |
| **VPC CIDR**                  | `10.0.0.0/16`                    | `10.1.0.0/16`                    |
| **Public subnets**            | `10.0.1.0/24`, `10.0.2.0/24`     | `10.1.1.0/24`, `10.1.2.0/24`     |
| **Private subnets**           | `10.0.11.0/24`, `10.0.12.0/24`   | `10.1.11.0/24`, `10.1.12.0/24`   |
| **NAT**                       | NAT Instance (t4g.nano)          | NAT Gateway                      |
| **Auto-scaling**              | 1-2 tasks                        | 2-6 tasks                        |
| **Backup retention**          | 7 days                           | 30 days                          |
| **Monthly cost (active)**     | ~$200                            | ~$350 (when deployed)            |

---

## Next Steps

1. **Implement Terraform modules**: Create reusable modules for VPC, ECS, RDS, ALB based on decisions above.
2. **Document workspace-specific tfvars**: Define `staging.tfvars` and `production.tfvars` with environment-specific values (instance sizes, subnet CIDRs, scaling limits).
3. **Set up GitHub Actions workflows**: Implement OIDC authentication, Docker builds, and ECS deployments.
4. **Bootstrap remote state backend**: Run `create-state-backend.sh` script to provision S3 bucket and DynamoDB table.
5. **Test staging deployment end-to-end**: Verify full CI/CD pipeline (push to main → build → deploy → health checks pass).
