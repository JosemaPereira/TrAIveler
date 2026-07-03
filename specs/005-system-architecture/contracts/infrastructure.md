# Infrastructure Contracts

**Feature**: System Architecture and Technology Stack  
**Created**: 2026-07-03  
**Layer**: Infrastructure (Terraform + CI/CD)

## Overview

This document defines integration contracts for Terraform modules, environment configuration, remote state management, and CI/CD workflows. Contracts ensure idempotent infrastructure provisioning, environment isolation, and zero-downtime deployments.

---

## Terraform Module Contracts

### Module Input/Output Convention

**Rule**: All modules MUST expose standardized inputs and outputs.

**Required Inputs** (all modules):
- `environment`: "staging" or "production"
- `project`: "traveler" (for resource tagging)

**Required Outputs** (where applicable):
- Resource identifiers (IDs, ARNs)
- Connectivity details (endpoints, URLs)

**Tagging Convention** (all AWS resources):
```hcl
tags = {
  Environment = var.environment
  ManagedBy   = "terraform"
  Project     = var.project
}
```

---

### VPC Module Contract

**Purpose**: Create isolated VPC with public/private subnets across 2 AZs.

**Inputs**:
```hcl
variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16" # staging: 10.0.0.0/16, production: 10.1.0.0/16
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "enable_nat_gateway" {
  description = "Use NAT Gateway (production) vs NAT Instance (staging)"
  type        = bool
  default     = false # staging: false, production: true
}
```

**Outputs**:
```hcl
output "vpc_id" {
  description = "VPC identifier"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "List of public subnet IDs for ALB"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "List of private subnet IDs for ECS, RDS"
  value       = aws_subnet.private[*].id
}
```

**Resources Created**:
- 1 VPC with DNS enabled
- 2 public subnets (for ALB) across 2 AZs
- 2 private subnets (for ECS, RDS) across 2 AZs
- 1 Internet Gateway for public subnet routing
- 1 NAT Gateway (production) or 1 NAT Instance (staging) for private subnet outbound
- Route tables associating subnets with gateways

**Cost Optimization**:
- Staging: NAT Instance (~$3/month vs $32/month for NAT Gateway)
- Production: NAT Gateway (higher availability, no instance management)

---

### ECS Module Contract

**Purpose**: Create ECS Fargate cluster, task definition, and service with auto-scaling.

**Inputs**:
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID (from VPC module)"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for ECS tasks"
  type        = list(string)
}

variable "task_cpu" {
  description = "CPU units (256 = 0.25 vCPU, 1024 = 1 vCPU)"
  type        = string
  default     = "256" # staging: 256, production: 1024
}

variable "task_memory" {
  description = "Memory in MB (512 = 0.5 GB, 2048 = 2 GB)"
  type        = string
  default     = "512" # staging: 512, production: 2048
}

variable "min_tasks" {
  description = "Minimum task count for auto-scaling"
  type        = number
  default     = 1 # staging: 1, production: 2
}

variable "max_tasks" {
  description = "Maximum task count for auto-scaling"
  type        = number
  default     = 2 # staging: 2, production: 20
}

variable "log_retention_days" {
  description = "CloudWatch log retention period"
  type        = number
  default     = 7 # staging: 7, production: 30
}

variable "ecr_repository_url" {
  description = "ECR repository URL for container image"
  type        = string
}
```

**Outputs**:
```hcl
output "cluster_name" {
  description = "ECS cluster name for CI/CD"
  value       = aws_ecs_cluster.main.name
}

output "service_name" {
  description = "ECS service name for CI/CD"
  value       = aws_ecs_service.main.name
}

output "task_definition_family" {
  description = "Task definition family for task def ARN construction"
  value       = aws_ecs_task_definition.main.family
}

output "log_group_name" {
  description = "CloudWatch log group name"
  value       = aws_cloudwatch_log_group.ecs.name
}
```

**Resources Created**:
- ECS cluster
- Task definition (ARM64, env-specific CPU/memory)
- ECS service with ALB target group integration
- Auto-scaling policy (target tracking on CPU at 70%)
- CloudWatch log group with retention
- Security group (allow ALB → ECS on port 8080)

**Health Check Configuration**:
- Health check path: `/healthz`
- Health check interval: 30s
- Health check timeout: 5s
- Healthy threshold: 2 consecutive successes
- Unhealthy threshold: 3 consecutive failures
- Grace period: 60s (allows migrations to complete)

---

### RDS Module Contract

**Purpose**: Create PostgreSQL RDS instance with automated backups.

**Inputs**:
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID (from VPC module)"
  type        = string
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for RDS"
  type        = list(string)
}

variable "instance_class" {
  description = "RDS instance type"
  type        = string
  default     = "db.t4g.micro" # staging: db.t4g.micro, production: db.t4g.small
}

variable "multi_az" {
  description = "Enable Multi-AZ for high availability"
  type        = bool
  default     = false # staging: false, production: true
}

variable "backup_retention_days" {
  description = "Automated backup retention period"
  type        = number
  default     = 1 # staging: 1, production: 30
}

variable "db_name" {
  description = "Initial database name"
  type        = string
  default     = "traveler"
}
```

**Outputs**:
```hcl
output "db_endpoint" {
  description = "RDS instance endpoint (host:port)"
  value       = aws_db_instance.main.endpoint
}

output "db_name" {
  description = "Database name"
  value       = aws_db_instance.main.db_name
}

output "db_secret_arn" {
  description = "ARN of AWS Secrets Manager secret with DB credentials"
  value       = aws_secretsmanager_secret.db_credentials.arn
  sensitive   = true
}
```

**Resources Created**:
- RDS PostgreSQL 15.4 instance
- DB subnet group (private subnets)
- Security group (allow ECS security group on port 5432)
- Parameter group (PostgreSQL 15 optimized settings)
- AWS Secrets Manager secret with master username/password

**Backup Configuration**:
- Automated daily backups at 03:00 UTC (low-traffic window)
- Backup retention: 1 day (staging), 30 days (production)
- Backup window: 03:00-04:00 UTC
- Maintenance window: Sunday 04:00-05:00 UTC

---

### ALB Module Contract

**Purpose**: Create Application Load Balancer with HTTPS termination.

**Inputs**:
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "vpc_id" {
  description = "VPC ID (from VPC module)"
  type        = string
}

variable "public_subnet_ids" {
  description = "Public subnet IDs for ALB"
  type        = list(string)
}

variable "certificate_arn" {
  description = "ACM certificate ARN for HTTPS"
  type        = string
}

variable "ecs_security_group_id" {
  description = "ECS security group ID (for target group)"
  type        = string
}
```

**Outputs**:
```hcl
output "alb_dns_name" {
  description = "ALB DNS name for CNAME record"
  value       = aws_lb.main.dns_name
}

output "alb_zone_id" {
  description = "ALB Route 53 hosted zone ID"
  value       = aws_lb.main.zone_id
}

output "target_group_arn" {
  description = "Target group ARN for ECS service"
  value       = aws_lb_target_group.ecs.arn
}
```

**Resources Created**:
- Application Load Balancer (internet-facing, public subnets)
- Target group (HTTP 8080, health check `/healthz`)
- HTTPS listener (443) with SSL certificate
- HTTP listener (80) redirecting to HTTPS

**Security Configuration**:
- ALB security group: Allow 443 from 0.0.0.0/0, allow 80 from 0.0.0.0/0
- ECS target group: Deregistration delay 30s (for graceful shutdown)

---

### CloudFront Module Contract

**Purpose**: Create CloudFront distribution for React SPA delivery.

**Inputs**:
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}

variable "domain_name" {
  description = "Custom domain for CloudFront"
  type        = string
}

variable "certificate_arn" {
  description = "ACM certificate ARN for custom domain"
  type        = string
}
```

**Outputs**:
```hcl
output "s3_bucket_name" {
  description = "S3 bucket name for frontend builds"
  value       = aws_s3_bucket.frontend.bucket
}

output "cloudfront_distribution_id" {
  description = "CloudFront distribution ID for cache invalidation"
  value       = aws_cloudfront_distribution.main.id
}

output "cloudfront_domain_name" {
  description = "CloudFront distribution domain name"
  value       = aws_cloudfront_distribution.main.domain_name
}
```

**Resources Created**:
- S3 bucket for React build artifacts (private, no public access)
- CloudFront Origin Access Identity (OAI) for S3 access
- CloudFront distribution (HTTPS-only, TLS 1.2+)
- Cache behavior (SPA routing: all paths → index.html)

**Cache Configuration**:
- Default TTL: 1 hour
- Max TTL: 24 hours
- Invalidation on deployment: `/*` (all paths)

---

### Secrets Module Contract

**Purpose**: Create AWS Secrets Manager secrets for application.

**Inputs**:
```hcl
variable "environment" {
  description = "Environment name"
  type        = string
}
```

**Outputs**:
```hcl
output "db_secret_arn" {
  description = "ARN for database credentials secret"
  value       = aws_secretsmanager_secret.db_credentials.arn
  sensitive   = true
}

output "ai_api_key_secret_arn" {
  description = "ARN for Anthropic API key secret"
  value       = aws_secretsmanager_secret.ai_api_key.arn
  sensitive   = true
}

output "jwt_signing_key_secret_arn" {
  description = "ARN for JWT RS256 private key secret"
  value       = aws_secretsmanager_secret.jwt_signing_key.arn
  sensitive   = true
}
```

**Resources Created**:
- Secret for database credentials (username, password)
- Secret for Anthropic API key
- Secret for JWT RS256 private key (multi-key rotation support)

**Secret Population**:
- Secrets created empty by Terraform
- Values populated manually via AWS Console or AWS CLI after `terraform apply`
- ECS task IAM role granted `secretsmanager:GetSecretValue` permission

---

## Environment Configuration Contract

### .tfvars File Structure

**Staging** (`infra/environments/staging.tfvars`):
```hcl
environment = "staging"
vpc_cidr = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]
enable_nat_gateway = false  # NAT Instance for cost

# ECS configuration
task_cpu = "256"    # 0.25 vCPU
task_memory = "512" # 0.5 GB RAM
min_tasks = 1
max_tasks = 2
log_retention_days = 7

# RDS configuration
instance_class = "db.t4g.micro"
multi_az = false
backup_retention_days = 1

# Domain configuration
backend_domain = "api.staging.traveler.example.com"
frontend_domain = "staging.traveler.example.com"
```

**Production** (`infra/environments/production.tfvars`):
```hcl
environment = "production"
vpc_cidr = "10.1.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]
enable_nat_gateway = true  # NAT Gateway for reliability

# ECS configuration
task_cpu = "1024"    # 1 vCPU
task_memory = "2048" # 2 GB RAM
min_tasks = 2
max_tasks = 20
log_retention_days = 30

# RDS configuration
instance_class = "db.t4g.small"
multi_az = true
backup_retention_days = 30

# Domain configuration
backend_domain = "api.traveler.example.com"
frontend_domain = "traveler.example.com"
```

---

## Remote State Management Contract

### S3 Backend Configuration

**File**: `infra/backend.tf`

```hcl
terraform {
  backend "s3" {
    bucket         = "traveler-terraform-state"
    key            = "infra/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "traveler-terraform-locks"
    encrypt        = true
  }
}
```

**Resources** (created manually before first `terraform init`):
- S3 bucket `traveler-terraform-state` with versioning enabled
- DynamoDB table `traveler-terraform-locks` with primary key `LockID` (String)

**State Locking**:
- `terraform plan` acquires read lock
- `terraform apply` acquires write lock
- Concurrent operations blocked by DynamoDB lock
- Lock automatically released on command completion

---

## CI/CD Integration Contracts

### Backend CI/CD Workflow

**File**: `.github/workflows/backend-ci.yml`

**Triggers**:
- Pull request to `main` (paths: `backend/**`)
- Push to `main` (paths: `backend/**`)

**Jobs**:
1. **Lint**: Run `golangci-lint`
2. **Test**: Run `go test -cover`
3. **Build**: Build Docker image, tag with commit SHA
4. **Push** (main only): Push image to ECR
5. **Deploy** (main only): Update ECS service with new task definition

**Outputs** (for ECS deployment):
```yaml
- name: Deploy to ECS
  env:
    AWS_REGION: us-east-1
    ECR_REPOSITORY: ${{ secrets.ECR_REPOSITORY_URL }}
    ECS_CLUSTER: ${{ secrets.ECS_CLUSTER_NAME_STAGING }}
    ECS_SERVICE: ${{ secrets.ECS_SERVICE_NAME_STAGING }}
  run: |
    aws ecs update-service \
      --cluster $ECS_CLUSTER \
      --service $ECS_SERVICE \
      --force-new-deployment
```

---

### Frontend CI/CD Workflow

**File**: `.github/workflows/frontend-ci.yml`

**Triggers**:
- Pull request to `main` (paths: `frontend/**`)
- Push to `main` (paths: `frontend/**`)

**Jobs**:
1. **Lint**: Run ESLint, Prettier check
2. **Test**: Run Vitest unit/component tests
3. **Build**: Build production bundle with Vite
4. **Accessibility**: Run Lighthouse CI audit
5. **Deploy** (main only): Sync build to S3, invalidate CloudFront cache

**Outputs** (for S3/CloudFront deployment):
```yaml
- name: Deploy to S3
  env:
    S3_BUCKET: ${{ secrets.S3_BUCKET_NAME_STAGING }}
    CLOUDFRONT_DISTRIBUTION_ID: ${{ secrets.CLOUDFRONT_DISTRIBUTION_ID_STAGING }}
  run: |
    aws s3 sync dist/ s3://$S3_BUCKET --delete
    aws cloudfront create-invalidation \
      --distribution-id $CLOUDFRONT_DISTRIBUTION_ID \
      --paths "/*"
```

---

### Infrastructure CI/CD Workflow

**File**: `.github/workflows/infra-plan.yml` (on PR)

**Triggers**:
- Pull request to `main` (paths: `infra/**`)

**Jobs**:
1. **Validate**: Run `terraform validate`
2. **Format Check**: Run `terraform fmt -check`
3. **Plan Staging**: Run `terraform plan -var-file=environments/staging.tfvars`
4. **Plan Production**: Run `terraform plan -var-file=environments/production.tfvars`
5. **Comment Plan**: Post plan output as PR comment

**File**: `.github/workflows/infra-apply.yml` (on main merge)

**Triggers**:
- Push to `main` (paths: `infra/**`)

**Jobs**:
1. **Apply Staging**: Run `terraform apply -var-file=environments/staging.tfvars -auto-approve`
2. **Output Export**: Export Terraform outputs to GitHub Secrets for CI/CD workflows

**Manual Production Deploy**:
- Production apply requires manual workflow dispatch
- Workflow requires approval from Infrastructure Engineer role
- Runs `terraform apply -var-file=environments/production.tfvars`

---

### OIDC Authentication Contract

**GitHub Actions → AWS OIDC Federation**:

**IAM Role Trust Policy**:
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:ORG/REPO:ref:refs/heads/main"
        }
      }
    }
  ]
}
```

**Workflow Configuration**:
```yaml
permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    steps:
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1
```

**No Long-Lived Credentials**: No AWS access keys stored in GitHub Secrets.

---

## Summary

Infrastructure contracts enforce:
- **Module composability**: Standardized inputs/outputs, dependency injection via outputs
- **Environment isolation**: Separate .tfvars files, non-overlapping VPC CIDR blocks
- **State management**: Remote state in S3 with DynamoDB locking (prevents concurrent modifications)
- **CI/CD automation**: Staging auto-deploys on main merge, production requires manual approval
- **Security**: OIDC authentication (no long-lived credentials), secrets in AWS Secrets Manager
- **Cost optimization**: Environment-specific resource sizing (staging: minimum viable, production: validated via load testing)

All contracts align with constitution principles: secure configuration (secrets management), simplicity (modular Terraform), infrastructure as code (version control, idempotency).
