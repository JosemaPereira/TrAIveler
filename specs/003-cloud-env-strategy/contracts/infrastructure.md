# Infrastructure Contracts: Cloud & Environments Strategy

**Feature**: [spec.md](../spec.md) | **Plan**: [plan.md](../plan.md)

**Purpose**: Define the interface contracts for infrastructure provisioning, CI/CD pipelines, and inter-service communication. These contracts ensure consistency across environments and enable infrastructure-as-code modularity.

---

## 1. Terraform Module Contracts

### 1.1 Root Module Interface

**Purpose**: Top-level Terraform entry point for provisioning an environment (staging or production).

**Location**: `infra/terraform/`

**Usage**:
```bash
cd infra/terraform
terraform workspace select staging  # or production
terraform init
terraform plan -var-file=staging.tfvars
terraform apply -var-file=staging.tfvars
```

#### Input Variables

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `environment` | string | Yes | — | Environment name: `"staging"` or `"production"` |
| `aws_region` | string | No | `"us-east-1"` | AWS region for all resources |
| `vpc_cidr` | string | Yes | — | VPC CIDR block (must not overlap) |
| `availability_zones` | list(string) | Yes | — | List of AZs for subnets (minimum 2) |
| `ecs_task_cpu` | string | Yes | — | Fargate task CPU: `"512"` (staging) or `"1024"` (production) |
| `ecs_task_memory` | string | Yes | — | Fargate task memory: `"1024"` (staging) or `"2048"` (production) |
| `ecs_min_capacity` | number | Yes | — | Auto-scaling minimum: `1` (staging) or `2` (production) |
| `ecs_max_capacity` | number | Yes | — | Auto-scaling maximum: `5` (staging) or `20` (production) |
| `rds_instance_class` | string | Yes | — | RDS instance type: `"db.t4g.micro"` (staging) or `"db.t4g.small"` (production) |
| `rds_multi_az` | bool | Yes | — | RDS Multi-AZ: `false` (staging) or `true` (production) |
| `rds_backup_retention_days` | number | Yes | — | Backup retention: `7` (staging) or `30` (production) |
| `nat_gateway_type` | string | Yes | — | NAT type: `"instance"` (staging) or `"gateway"` (production) |
| `cost_budget_usd` | number | Yes | — | Monthly cost ceiling: `200` (staging) or `400` (production) |
| `tags` | map(string) | No | `{}` | Additional resource tags (merged with default tags) |

#### Output Values

| Output | Type | Description | Example |
|--------|------|-------------|---------|
| `vpc_id` | string | VPC identifier | `"vpc-0abc123def456789"` |
| `public_subnet_ids` | list(string) | Public subnet IDs (ALB) | `["subnet-0abc...", "subnet-0def..."]` |
| `private_subnet_ids` | list(string) | Private subnet IDs (ECS, RDS) | `["subnet-0123...", "subnet-0456..."]` |
| `alb_dns_name` | string | ALB public DNS | `"trAIveler-alb-staging-1234567890.us-east-1.elb.amazonaws.com"` |
| `alb_arn` | string | ALB ARN | `"arn:aws:elasticloadbalancing:us-east-1:..."` |
| `ecs_cluster_name` | string | ECS cluster name | `"trAIveler-staging"` |
| `ecs_cluster_arn` | string | ECS cluster ARN | `"arn:aws:ecs:us-east-1:123456789012:cluster/trAIveler-staging"` |
| `ecs_service_name` | string | ECS service name | `"trAIveler-backend-staging"` |
| `ecr_repository_url` | string | ECR repository URI | `"123456789012.dkr.ecr.us-east-1.amazonaws.com/trAIveler-backend"` |
| `rds_endpoint` | string | RDS connection endpoint | `"trAIveler-staging.abc123def456.us-east-1.rds.amazonaws.com:5432"` |
| `rds_instance_id` | string | RDS instance identifier | `"trAIveler-staging"` |
| `s3_frontend_bucket` | string | S3 bucket for frontend | `"trAIveler-frontend-staging"` |
| `cloudfront_distribution_id` | string | CloudFront distribution ID | `"E1ABC123DEF456"` |
| `cloudfront_domain_name` | string | CloudFront public domain | `"d123abc456def.cloudfront.net"` |

#### Provider Configuration

**AWS Provider** (required):
```hcl
terraform {
  required_version = ">= 1.5.0"
  
  backend "s3" {
    bucket         = "trAIveler-terraform-state"
    key            = "env/${terraform.workspace}/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "trAIveler-terraform-locks"
    encrypt        = true
  }
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Project     = "TrAIveler"
      Environment = var.environment
      ManagedBy   = "terraform"
      CostCenter  = "mvp"
    }
  }
}
```

---

### 1.2 VPC Module Interface

**Purpose**: Provision VPC, subnets, route tables, NAT, security groups.

**Location**: `infra/terraform/modules/vpc/`

#### Input Variables

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `environment` | string | Yes | Environment name |
| `vpc_cidr` | string | Yes | VPC CIDR block |
| `availability_zones` | list(string) | Yes | AZs for subnet distribution |
| `nat_gateway_type` | string | Yes | `"instance"` or `"gateway"` |

#### Output Values

| Output | Type | Description |
|--------|------|-------------|
| `vpc_id` | string | VPC ID |
| `public_subnet_ids` | list(string) | Public subnet IDs |
| `private_subnet_ids` | list(string) | Private subnet IDs |
| `alb_security_group_id` | string | ALB security group ID |
| `ecs_security_group_id` | string | ECS security group ID |
| `rds_security_group_id` | string | RDS security group ID |

#### Security Group Rules

**ALB Security Group** (`alb-sg`):
- **Ingress**:
  - Port 80 (HTTP): `0.0.0.0/0` → Redirect to 443
  - Port 443 (HTTPS): `0.0.0.0/0` → Accept
- **Egress**:
  - All traffic: `0.0.0.0/0` → Allow

**ECS Security Group** (`ecs-sg`):
- **Ingress**:
  - Port 8080 (HTTP): `source_security_group_id = alb-sg` → Accept
- **Egress**:
  - All traffic: `0.0.0.0/0` → Allow

**RDS Security Group** (`rds-sg`):
- **Ingress**:
  - Port 5432 (PostgreSQL): `source_security_group_id = ecs-sg` → Accept
- **Egress**:
  - None (RDS does not initiate outbound connections)

---

### 1.3 ECS Module Interface

**Purpose**: Provision ECS cluster, task definition, service, auto-scaling, ALB target group.

**Location**: `infra/terraform/modules/ecs/`

#### Input Variables

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `environment` | string | Yes | Environment name |
| `cluster_name` | string | Yes | ECS cluster name |
| `vpc_id` | string | Yes | VPC ID (from VPC module) |
| `private_subnet_ids` | list(string) | Yes | Private subnets for tasks |
| `alb_target_group_arn` | string | Yes | ALB target group ARN |
| `ecs_security_group_id` | string | Yes | ECS security group ID |
| `task_cpu` | string | Yes | Fargate task CPU |
| `task_memory` | string | Yes | Fargate task memory |
| `min_capacity` | number | Yes | Auto-scaling min |
| `max_capacity` | number | Yes | Auto-scaling max |
| `ecr_image_uri` | string | Yes | Docker image URI |
| `database_url_secret_arn` | string | Yes | Secrets Manager ARN for DB connection string |
| `anthropic_api_key_secret_arn` | string | Yes | Secrets Manager ARN for Anthropic API key |
| `jwt_secret_secret_arn` | string | Yes | Secrets Manager ARN for JWT secret |

#### Output Values

| Output | Type | Description |
|--------|------|-------------|
| `ecs_cluster_arn` | string | ECS cluster ARN |
| `ecs_service_name` | string | ECS service name |
| `task_definition_arn` | string | ECS task definition ARN |
| `service_arn` | string | ECS service ARN |

#### Task Definition Contract

**Family**: `trAIveler-backend-${environment}`

**Container Definition**:
```json
{
  "name": "backend",
  "image": "<ecr_image_uri>",
  "cpu": <task_cpu>,
  "memory": <task_memory>,
  "essential": true,
  "portMappings": [
    {
      "containerPort": 8080,
      "protocol": "tcp"
    }
  ],
  "environment": [
    {"name": "PORT", "value": "8080"},
    {"name": "ENVIRONMENT", "value": "<environment>"}
  ],
  "secrets": [
    {"name": "DATABASE_URL", "valueFrom": "<database_url_secret_arn>"},
    {"name": "ANTHROPIC_API_KEY", "valueFrom": "<anthropic_api_key_secret_arn>"},
    {"name": "JWT_SECRET", "valueFrom": "<jwt_secret_secret_arn>"}
  ],
  "logConfiguration": {
    "logDriver": "awslogs",
    "options": {
      "awslogs-group": "/ecs/trAIveler-backend-<environment>",
      "awslogs-region": "us-east-1",
      "awslogs-stream-prefix": "ecs"
    }
  }
}
```

**IAM Roles**:
- **Task Role** (`ecsTaskRole-${environment}`): Permissions for application runtime (Secrets Manager read, S3 access)
- **Execution Role** (`ecsTaskExecutionRole-${environment}`): Permissions for ECS agent (ECR pull, CloudWatch Logs write)

---

### 1.4 RDS Module Interface

**Purpose**: Provision RDS PostgreSQL instance with backups and security.

**Location**: `infra/terraform/modules/rds/`

#### Input Variables

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `environment` | string | Yes | Environment name |
| `instance_identifier` | string | Yes | RDS instance name |
| `instance_class` | string | Yes | RDS instance type |
| `vpc_id` | string | Yes | VPC ID |
| `private_subnet_ids` | list(string) | Yes | Private subnets for RDS |
| `rds_security_group_id` | string | Yes | RDS security group ID |
| `multi_az` | bool | Yes | Multi-AZ deployment |
| `backup_retention_days` | number | Yes | Backup retention period |
| `database_name` | string | Yes | Initial database name |
| `master_username` | string | Yes | Master username |
| `master_password_secret_arn` | string | Yes | Secrets Manager ARN for master password |

#### Output Values

| Output | Type | Description |
|--------|------|-------------|
| `endpoint` | string | RDS connection endpoint (host:port) |
| `instance_id` | string | RDS instance identifier |
| `database_name` | string | Database name |

---

### 1.5 ALB Module Interface

**Purpose**: Provision Application Load Balancer, listeners, target groups.

**Location**: `infra/terraform/modules/alb/`

#### Input Variables

| Variable | Type | Required | Description |
|----------|------|----------|-------------|
| `environment` | string | Yes | Environment name |
| `vpc_id` | string | Yes | VPC ID |
| `public_subnet_ids` | list(string) | Yes | Public subnets for ALB |
| `alb_security_group_id` | string | Yes | ALB security group ID |
| `certificate_arn` | string | No | ACM certificate ARN (optional for staging with HTTP-only) |

#### Output Values

| Output | Type | Description |
|--------|------|-------------|
| `alb_arn` | string | ALB ARN |
| `alb_dns_name` | string | ALB public DNS |
| `target_group_arn` | string | Target group ARN |

#### Target Group Contract

- **Protocol**: HTTP
- **Port**: 8080
- **Target Type**: `ip` (Fargate)
- **Health Check Path**: `/healthz`
- **Health Check Interval**: 30 seconds
- **Healthy Threshold**: 2
- **Unhealthy Threshold**: 3
- **Deregistration Delay**: 30 seconds

---

## 2. GitHub Actions Workflow Contracts

### 2.1 Terraform Plan Workflow

**Purpose**: Validate Terraform changes on every PR.

**File**: `.github/workflows/terraform-plan.yml`

**Triggers**:
- `pull_request` (paths: `infra/terraform/**`)

**Inputs**: None (reads from PR context)

**Outputs**:
- **PR Comment**: Terraform plan output with resource changes
- **Status Check**: Pass/fail based on `terraform plan` exit code

**Steps**:
1. Checkout code
2. Configure AWS credentials via OIDC (read-only role)
3. Setup Terraform CLI
4. `terraform fmt -check` (fail on formatting issues)
5. `terraform init`
6. `terraform validate`
7. `terraform workspace select staging`
8. `terraform plan -var-file=staging.tfvars -out=staging.tfplan`
9. `terraform show -no-color staging.tfplan` → Post as PR comment
10. Repeat steps 7-9 for `production` workspace

**Required Secrets/Variables**:
- `AWS_ROLE_ARN_READONLY`: IAM role for read-only Terraform plan

---

### 2.2 Terraform Apply Workflow

**Purpose**: Deploy infrastructure changes to staging automatically on merge to `main`; deploy to production manually.

**File**: `.github/workflows/terraform-apply.yml`

**Triggers**:
- `push` (branch: `main`, paths: `infra/terraform/**`) → Auto-deploy to staging
- `workflow_dispatch` (input: `environment=production`) → Manual production deploy

**Inputs** (for `workflow_dispatch`):
- `environment`: `"staging"` or `"production"` (default: `"staging"`)
- `auto_approve`: `true` or `false` (default: `false`, requires manual approval for production)

**Outputs**:
- **Deployment Status**: Success/failure with Terraform output values
- **Summary**: Updated infrastructure resources (created/modified/destroyed counts)

**Steps**:
1. Checkout code
2. Configure AWS credentials via OIDC (read-write role for `${{ inputs.environment }}`)
3. Setup Terraform CLI
4. `terraform init`
5. `terraform workspace select ${{ inputs.environment }}`
6. `terraform plan -var-file=${{ inputs.environment }}.tfvars -out=plan.tfplan`
7. **(Manual Approval for Production)**: If `environment=production` and `auto_approve=false`, pause for manual approval
8. `terraform apply plan.tfplan`
9. `terraform output -json` → Export outputs as job outputs

**Required Secrets/Variables**:
- `AWS_ROLE_ARN_STAGING`: IAM role for staging infrastructure write access
- `AWS_ROLE_ARN_PRODUCTION`: IAM role for production infrastructure write access

---

### 2.3 Backend Deploy Workflow

**Purpose**: Build, push, and deploy backend Docker image to ECS.

**File**: `.github/workflows/backend-deploy.yml`

**Triggers**:
- `push` (branch: `main`, paths: `backend/**`) → Auto-deploy to staging
- `workflow_dispatch` (input: `environment=production`) → Manual production deploy

**Inputs** (for `workflow_dispatch`):
- `environment`: `"staging"` or `"production"` (default: `"staging"`)
- `image_tag`: Optional override for image tag (default: auto-generated from git SHA)

**Outputs**:
- **Image URI**: Full ECR image URI with tag
- **ECS Service**: Updated ECS service name
- **Deployment Status**: Success/failure

**Steps**:
1. Checkout code
2. Configure AWS credentials via OIDC
3. Login to Amazon ECR
4. Extract metadata (version from git tag or SHA, build timestamp)
5. Build Docker image: `docker buildx build --platform linux/arm64 --tag <ecr_url>:v<version>-<sha> backend/`
6. Push image to ECR
7. Update ECS task definition with new image URI
8. Register new task definition revision
9. Update ECS service to use new task definition
10. Wait for service stability (health checks pass, old tasks drained)

**Required Secrets/Variables**:
- `AWS_ROLE_ARN_STAGING`: IAM role for ECS deployment (staging)
- `AWS_ROLE_ARN_PRODUCTION`: IAM role for ECS deployment (production)
- `ECR_REPOSITORY_NAME`: `"trAIveler-backend"`
- `ECS_CLUSTER_NAME_STAGING`: `"trAIveler-staging"`
- `ECS_CLUSTER_NAME_PRODUCTION`: `"trAIveler-production"`
- `ECS_SERVICE_NAME_STAGING`: `"trAIveler-backend-staging"`
- `ECS_SERVICE_NAME_PRODUCTION`: `"trAIveler-backend-production"`

---

### 2.4 Frontend Deploy Workflow

**Purpose**: Build and deploy frontend static assets to S3 + CloudFront.

**File**: `.github/workflows/frontend-deploy.yml`

**Triggers**:
- `push` (branch: `main`, paths: `frontend/**`) → Auto-deploy to staging
- `workflow_dispatch` (input: `environment=production`) → Manual production deploy

**Inputs** (for `workflow_dispatch`):
- `environment`: `"staging"` or `"production"` (default: `"staging"`)

**Outputs**:
- **S3 Bucket**: S3 bucket name
- **CloudFront URL**: CloudFront distribution domain
- **Deployment Status**: Success/failure

**Steps**:
1. Checkout code
2. Setup Node.js 20
3. Install dependencies: `npm ci`
4. Build production bundle: `VITE_API_URL=${{ vars.API_URL_STAGING }} npm run build`
5. Configure AWS credentials via OIDC
6. Sync build artifacts to S3: `aws s3 sync dist/ s3://${{ vars.S3_BUCKET_STAGING }}/ --delete`
7. Invalidate CloudFront cache: `aws cloudfront create-invalidation --distribution-id ${{ vars.CLOUDFRONT_DISTRIBUTION_ID_STAGING }} --paths "/*"`
8. Wait for invalidation completion

**Required Secrets/Variables**:
- `AWS_ROLE_ARN_STAGING`: IAM role for S3/CloudFront deployment (staging)
- `AWS_ROLE_ARN_PRODUCTION`: IAM role for S3/CloudFront deployment (production)
- `API_URL_STAGING`: Backend API URL (e.g., `https://api-staging.trAIveler.com`)
- `API_URL_PRODUCTION`: Backend API URL (e.g., `https://api.trAIveler.com`)
- `S3_BUCKET_STAGING`: `"trAIveler-frontend-staging"`
- `S3_BUCKET_PRODUCTION`: `"trAIveler-frontend-production"`
- `CLOUDFRONT_DISTRIBUTION_ID_STAGING`: CloudFront distribution ID
- `CLOUDFRONT_DISTRIBUTION_ID_PRODUCTION`: CloudFront distribution ID

---

## 3. AWS Resource Tagging Contract

**Purpose**: Ensure consistent cost allocation, environment identification, and resource management.

### Required Tags (All Resources)

| Tag Key | Tag Value | Description |
|---------|-----------|-------------|
| `Project` | `"TrAIveler"` | Project identifier |
| `Environment` | `"staging"` \| `"production"` | Environment name |
| `ManagedBy` | `"terraform"` | Infrastructure management tool |
| `CostCenter` | `"mvp"` | Cost allocation identifier |

### Optional Tags

| Tag Key | Tag Value | Description |
|---------|-----------|-------------|
| `Service` | `"backend"` \| `"database"` \| `"frontend"` | Service component |
| `Owner` | Email or team name | Responsible party |
| `Version` | Semantic version | Application version (for deployments) |

### Tag Enforcement

- **Terraform**: Applied via `default_tags` in AWS provider configuration
- **AWS Config**: Rule enforcing required tags on all resources
- **Cost Explorer**: Filters and groups by `Environment`, `Service`, and `CostCenter` tags

---

## 4. Secrets Manager Contract

**Purpose**: Define secret naming conventions and access patterns.

### Secret Naming Convention

**Pattern**: `${environment}/${service}/${secret_name}`

**Examples**:
- `staging/database/url` — Database connection string
- `staging/anthropic/api-key` — Anthropic Claude API key
- `staging/jwt/secret` — JWT signing secret
- `production/database/url` — Production database connection string

### Secret Value Format

**Database URL**:
```
postgresql://<username>:<password>@<rds_endpoint>/<database_name>
```

**Anthropic API Key**:
```
sk-ant-api03-<key>
```

**JWT Secret** (base64-encoded 32-byte random value):
```
<base64_string>
```

### IAM Access Policy

ECS Task Role must have `secretsmanager:GetSecretValue` permission for secrets matching pattern `${environment}/*`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "secretsmanager:GetSecretValue",
      "Resource": "arn:aws:secretsmanager:us-east-1:<account>:secret:${environment}/*"
    }
  ]
}
```

---

## 5. Health Check Endpoint Contract

**Purpose**: Define the health check endpoint used by ALB target groups and monitoring systems.

### Endpoint

**Path**: `GET /healthz`

**Response**:
- **Status Code**: `200 OK` (healthy) or `503 Service Unavailable` (unhealthy)
- **Content-Type**: `application/json`

**Response Body** (Healthy):
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime_seconds": 3625
}
```

**Response Body** (Unhealthy):
```json
{
  "status": "error",
  "error": "database connection failed"
}
```

### Requirements

- Must respond within **100 milliseconds**
- Must check critical dependencies (database connectivity)
- Must not require authentication
- Must not perform expensive operations (e.g., no full database scans)

### ALB Health Check Configuration

- **Path**: `/healthz`
- **Interval**: 30 seconds
- **Timeout**: 5 seconds
- **Healthy Threshold**: 2 consecutive successes
- **Unhealthy Threshold**: 3 consecutive failures
- **Matcher**: HTTP 200

---

## 6. Deployment Rollback Contract

**Purpose**: Define rollback conditions and procedures for failed deployments.

### Automatic Rollback Triggers

ECS deployment circuit breaker triggers rollback if:
1. New task fails health checks after grace period (60 seconds)
2. New task crashes immediately after start
3. CloudWatch alarm triggers (e.g., error rate > 5%)

### Manual Rollback Procedure

**Backend (ECS)**:
1. Identify last stable task definition revision: `aws ecs describe-services --cluster <cluster> --services <service>`
2. Update service to previous revision: `aws ecs update-service --cluster <cluster> --service <service> --task-definition <previous_revision>`
3. Wait for service stability: Service reaches steady state with all tasks healthy

**Frontend (S3 + CloudFront)**:
1. List S3 bucket versions: `aws s3api list-object-versions --bucket <bucket>`
2. Restore previous version: `aws s3 sync s3://<bucket>/builds/<previous_version>/ s3://<bucket>/ --delete`
3. Invalidate CloudFront: `aws cloudfront create-invalidation --distribution-id <id> --paths "/*"`

---

## Notes

- All contracts are versioned via Git; breaking changes require coordination with dependent systems
- Terraform module interfaces follow HashiCorp best practices for input/output naming
- GitHub Actions workflows use reusable workflows where possible to reduce duplication
- Secrets rotation requires updating Secrets Manager values; ECS tasks will pick up new values on next deployment
- Cost allocation tags are required for all billable resources; AWS Config enforces this policy
