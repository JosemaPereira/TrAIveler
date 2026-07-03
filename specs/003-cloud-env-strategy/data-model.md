# Data Model: Cloud & Environments Strategy

**Feature**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

**Purpose**: Define the operational data structures that describe infrastructure state, configuration, and observability for the TrAIveler cloud environments.

**Scope**: This is **not** a database schema. This document describes:
- Terraform state representations
- AWS resource metadata and configuration schemas
- CI/CD pipeline artifacts and manifests
- Observability and monitoring data structures

---

## Infrastructure Configuration Entities

### 1. Environment

**Description**: Represents a deployment environment (staging or production) with its own isolated infrastructure stack.

**Attributes**:
- `name` (string, required): Environment identifier — `"staging"` or `"production"`
- `vpc_cidr` (string, required): VPC CIDR block — `"10.0.0.0/16"` (staging), `"10.1.0.0/16"` (production)
- `availability_zones` ([]string, required): List of AZs for subnet distribution — e.g., `["us-east-1a", "us-east-1b"]`
- `is_active` (boolean, required): Whether infrastructure is provisioned — `true` (staging), `false` (production dormant)
- `cost_budget_usd` (integer, required): Monthly cost ceiling — `200` (staging), `400` (production)
- `tags` (map[string]string, required): AWS resource tags — `{"Project": "TrAIveler", "Environment": "staging|production", "ManagedBy": "terraform", "CostCenter": "mvp"}`

**Example**:
```yaml
name: staging
vpc_cidr: 10.0.0.0/16
availability_zones:
  - us-east-1a
  - us-east-1b
is_active: true
cost_budget_usd: 200
tags:
  Project: TrAIveler
  Environment: staging
  ManagedBy: terraform
  CostCenter: mvp
```

**Relationships**:
- Has one `VPCConfiguration`
- Has one `ECSClusterConfiguration`
- Has one `RDSInstanceConfiguration`
- Has one `LoadBalancerConfiguration`

---

### 2. VPCConfiguration

**Description**: Network isolation configuration for an environment.

**Attributes**:
- `environment` (string, required): Parent environment name
- `vpc_id` (string, output): AWS VPC ID — e.g., `"vpc-0abc123def456789"`
- `cidr_block` (string, required): VPC IPv4 CIDR — `"10.0.0.0/16"` or `"10.1.0.0/16"`
- `public_subnets` ([]Subnet, required): Subnets for ALB — 2 AZs × /24 each
- `private_subnets` ([]Subnet, required): Subnets for ECS tasks and RDS — 2 AZs × /20 each
- `nat_gateway_type` (enum, required): `"instance"` (staging) or `"gateway"` (production)
- `security_groups` ([]SecurityGroup, required): Firewall rules for ALB, ECS, RDS

**Subnet Sub-Schema**:
```yaml
cidr: "10.0.1.0/24"
availability_zone: "us-east-1a"
purpose: "public|private"
```

**SecurityGroup Sub-Schema**:
```yaml
name: "alb-sg|ecs-sg|rds-sg"
ingress_rules:
  - protocol: tcp
    from_port: 443
    to_port: 443
    cidr_blocks: ["0.0.0.0/0"]  # ALB only
egress_rules:
  - protocol: "-1"
    from_port: 0
    to_port: 0
    cidr_blocks: ["0.0.0.0/0"]
```

**Example**:
```yaml
environment: staging
cidr_block: 10.0.0.0/16
public_subnets:
  - cidr: 10.0.1.0/24
    availability_zone: us-east-1a
    purpose: public
  - cidr: 10.0.2.0/24
    availability_zone: us-east-1b
    purpose: public
private_subnets:
  - cidr: 10.0.16.0/20
    availability_zone: us-east-1a
    purpose: private
  - cidr: 10.0.32.0/20
    availability_zone: us-east-1b
    purpose: private
nat_gateway_type: instance
security_groups:
  - name: alb-sg
    ingress_rules:
      - protocol: tcp
        from_port: 443
        to_port: 443
        cidr_blocks: ["0.0.0.0/0"]
  - name: ecs-sg
    ingress_rules:
      - protocol: tcp
        from_port: 8080
        to_port: 8080
        source_security_group: alb-sg
  - name: rds-sg
    ingress_rules:
      - protocol: tcp
        from_port: 5432
        to_port: 5432
        source_security_group: ecs-sg
```

---

### 3. ECSClusterConfiguration

**Description**: Container orchestration configuration for backend services.

**Attributes**:
- `environment` (string, required): Parent environment name
- `cluster_name` (string, required): ECS cluster identifier — `"trAIveler-staging"` or `"trAIveler-production"`
- `cluster_arn` (string, output): AWS ARN of the ECS cluster
- `task_definition` (TaskDefinition, required): Container and resource configuration
- `service_configuration` (ServiceConfiguration, required): Deployment and scaling settings

**TaskDefinition Sub-Schema**:
```yaml
family: "trAIveler-backend-staging"
cpu: "512"  # 0.5 vCPU staging, 1024 (1 vCPU) production
memory: "1024"  # 1 GB staging, 2048 (2 GB) production
architecture: "ARM64"  # Graviton2 for cost savings
container_definitions:
  - name: backend
    image: "<account_id>.dkr.ecr.us-east-1.amazonaws.com/trAIveler-backend:latest"
    port_mappings:
      - container_port: 8080
        protocol: tcp
    environment_variables:
      - name: PORT
        value: "8080"
      - name: ENVIRONMENT
        value: "staging"
    secrets:
      - name: DATABASE_URL
        value_from: "arn:aws:secretsmanager:us-east-1:<account>:secret:staging/database-url"
      - name: ANTHROPIC_API_KEY
        value_from: "arn:aws:secretsmanager:us-east-1:<account>:secret:staging/anthropic-api-key"
      - name: JWT_SECRET
        value_from: "arn:aws:secretsmanager:us-east-1:<account>:secret:staging/jwt-secret"
    log_configuration:
      log_driver: awslogs
      options:
        awslogs-group: "/ecs/trAIveler-backend-staging"
        awslogs-region: "us-east-1"
        awslogs-stream-prefix: "ecs"
task_role_arn: "arn:aws:iam::<account>:role/ecsTaskRole-staging"
execution_role_arn: "arn:aws:iam::<account>:role/ecsTaskExecutionRole-staging"
```

**ServiceConfiguration Sub-Schema**:
```yaml
service_name: "trAIveler-backend-staging"
desired_count: 1  # staging minimum
deployment_configuration:
  maximum_percent: 200  # rolling update
  minimum_healthy_percent: 100  # zero-downtime
  deployment_circuit_breaker:
    enable: true
    rollback: true
health_check_grace_period_seconds: 60
autoscaling:
  min_capacity: 1
  max_capacity: 5  # staging limit
  target_tracking:
    target_value: 70  # CPU %
    scale_in_cooldown: 300
    scale_out_cooldown: 60
```

**Example**:
```yaml
environment: staging
cluster_name: trAIveler-staging
task_definition:
  family: trAIveler-backend-staging
  cpu: "512"
  memory: "1024"
  architecture: ARM64
  container_definitions:
    - name: backend
      image: 123456789012.dkr.ecr.us-east-1.amazonaws.com/trAIveler-backend:v1.0.0-abc123
      port_mappings:
        - container_port: 8080
      environment_variables:
        - name: PORT
          value: "8080"
      secrets:
        - name: DATABASE_URL
          value_from: "arn:aws:secretsmanager:us-east-1:123456789012:secret:staging/database-url-XYZ123"
service_configuration:
  service_name: trAIveler-backend-staging
  desired_count: 1
  autoscaling:
    min_capacity: 1
    max_capacity: 5
    target_tracking:
      target_value: 70
```

---

### 4. RDSInstanceConfiguration

**Description**: Managed PostgreSQL database configuration.

**Attributes**:
- `environment` (string, required): Parent environment name
- `instance_identifier` (string, required): RDS instance name — `"trAIveler-staging"` or `"trAIveler-production"`
- `instance_class` (string, required): Instance type — `"db.t4g.micro"` (staging) or `"db.t4g.small"` (production)
- `engine` (string, required): Database engine — `"postgres"`
- `engine_version` (string, required): PostgreSQL version — `"15.4"`
- `allocated_storage_gb` (integer, required): Initial disk size — `20` (staging), `100` (production)
- `multi_az` (boolean, required): High availability — `false` (staging), `true` (production)
- `backup_retention_days` (integer, required): Automated backup retention — `7` (staging), `30` (production)
- `maintenance_window` (string, required): Maintenance schedule — `"sun:03:00-sun:04:00"` (UTC, non-business hours)
- `database_name` (string, required): Initial database — `"trAIveler"`
- `username` (string, required): Master username — `"trAIvelerAdmin"`
- `password_secret_arn` (string, required): AWS Secrets Manager ARN for master password
- `security_group_ids` ([]string, required): Firewall rules — RDS security group only

**Example**:
```yaml
environment: staging
instance_identifier: trAIveler-staging
instance_class: db.t4g.micro
engine: postgres
engine_version: "15.4"
allocated_storage_gb: 20
multi_az: false
backup_retention_days: 7
maintenance_window: "sun:03:00-sun:04:00"
database_name: trAIveler
username: trAIvelerAdmin
password_secret_arn: "arn:aws:secretsmanager:us-east-1:123456789012:secret:staging/rds-password-XYZ123"
security_group_ids:
  - "sg-0abc123def456789"  # rds-sg
```

---

### 5. LoadBalancerConfiguration

**Description**: Application Load Balancer configuration for HTTP/HTTPS traffic.

**Attributes**:
- `environment` (string, required): Parent environment name
- `load_balancer_name` (string, required): ALB identifier — `"trAIveler-alb-staging"`
- `load_balancer_arn` (string, output): AWS ARN
- `scheme` (enum, required): `"internet-facing"`
- `security_group_ids` ([]string, required): ALB security group
- `subnet_ids` ([]string, required): Public subnets (2 AZs minimum)
- `listeners` ([]Listener, required): HTTP and HTTPS listeners
- `target_groups` ([]TargetGroup, required): ECS service backends

**Listener Sub-Schema**:
```yaml
protocol: "HTTP|HTTPS"
port: 80|443
default_action:
  type: "redirect|forward"
  target_group_arn: "<target_group_arn>"  # for forward
  redirect:  # for HTTP → HTTPS
    protocol: "HTTPS"
    port: "443"
    status_code: "HTTP_301"
certificate_arn: "<acm_certificate_arn>"  # HTTPS only
```

**TargetGroup Sub-Schema**:
```yaml
name: "trAIveler-backend-staging-tg"
protocol: "HTTP"
port: 8080
target_type: "ip"  # Fargate
vpc_id: "<vpc_id>"
health_check:
  path: "/healthz"
  interval_seconds: 30
  timeout_seconds: 5
  healthy_threshold_count: 2
  unhealthy_threshold_count: 3
  matcher: "200"
deregistration_delay_seconds: 30
```

**Example**:
```yaml
environment: staging
load_balancer_name: trAIveler-alb-staging
scheme: internet-facing
security_group_ids:
  - "sg-0abc123def456789"  # alb-sg
subnet_ids:
  - "subnet-0abc123def456789"  # public-1a
  - "subnet-0def456ghi789abc"  # public-1b
listeners:
  - protocol: HTTP
    port: 80
    default_action:
      type: redirect
      redirect:
        protocol: HTTPS
        port: "443"
        status_code: HTTP_301
  - protocol: HTTPS
    port: 443
    certificate_arn: "arn:aws:acm:us-east-1:123456789012:certificate/abc-123-def"
    default_action:
      type: forward
      target_group_arn: "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/trAIveler-backend-staging-tg/abc123"
target_groups:
  - name: trAIveler-backend-staging-tg
    protocol: HTTP
    port: 8080
    target_type: ip
    health_check:
      path: /healthz
      interval_seconds: 30
      timeout_seconds: 5
      healthy_threshold_count: 2
      unhealthy_threshold_count: 3
```

---

## CI/CD Pipeline Entities

### 6. DeploymentArtifact

**Description**: Immutable build output versioned and promoted across environments.

**Attributes**:
- `artifact_type` (enum, required): `"docker_image"` or `"static_bundle"`
- `component` (enum, required): `"backend"` or `"frontend"`
- `version` (string, required): Semantic version — `"1.0.0"`
- `git_sha` (string, required): Full commit SHA — `"abc123def456..."`
- `build_timestamp` (ISO8601, required): UTC timestamp — `"2026-07-03T10:30:00Z"`
- `image_uri` (string, required for backend): ECR image URI — `"<account>.dkr.ecr.us-east-1.amazonaws.com/trAIveler-backend:v1.0.0-abc123"`
- `s3_bucket` (string, required for frontend): S3 bucket — `"trAIveler-frontend-staging"`
- `s3_key_prefix` (string, required for frontend): S3 path prefix — `"builds/v1.0.0-abc123/"`
- `cloudfront_distribution_id` (string, required for frontend): CloudFront ID for invalidation

**Example (Backend)**:
```yaml
artifact_type: docker_image
component: backend
version: "1.0.0"
git_sha: "abc123def456789012345678901234567890abcd"
build_timestamp: "2026-07-03T10:30:00Z"
image_uri: "123456789012.dkr.ecr.us-east-1.amazonaws.com/trAIveler-backend:v1.0.0-abc123"
```

**Example (Frontend)**:
```yaml
artifact_type: static_bundle
component: frontend
version: "1.0.0"
git_sha: "abc123def456789012345678901234567890abcd"
build_timestamp: "2026-07-03T10:30:00Z"
s3_bucket: "trAIveler-frontend-staging"
s3_key_prefix: "builds/v1.0.0-abc123/"
cloudfront_distribution_id: "E1ABC123DEF456"
```

---

### 7. TerraformWorkspace

**Description**: Terraform workspace representing an environment's infrastructure state.

**Attributes**:
- `name` (string, required): Workspace identifier — `"staging"` or `"production"`
- `state_bucket` (string, required): S3 bucket for remote state — `"trAIveler-terraform-state"`
- `state_key` (string, required): S3 object key — `"env/staging/terraform.tfstate"` or `"env/production/terraform.tfstate"`
- `lock_table` (string, required): DynamoDB table for state locking — `"trAIveler-terraform-locks"`
- `tfvars_file` (string, required): Workspace-specific variables — `"staging.tfvars"` or `"production.tfvars"`
- `is_provisioned` (boolean, required): Whether `terraform apply` has been run — `true` (staging), `false` (production dormant)

**Example**:
```yaml
name: staging
state_bucket: trAIveler-terraform-state
state_key: "env/staging/terraform.tfstate"
lock_table: trAIveler-terraform-locks
tfvars_file: staging.tfvars
is_provisioned: true
```

---

## Observability Entities

### 8. InfrastructureMetrics

**Description**: CloudWatch metrics for infrastructure health monitoring.

**Attributes**:
- `environment` (string, required): Environment name
- `timestamp` (ISO8601, required): Metric collection time
- `ecs_cluster_metrics` (ECSClusterMetrics, required): Container orchestration health
- `alb_metrics` (ALBMetrics, required): Load balancer performance
- `rds_metrics` (RDSMetrics, required): Database health

**ECSClusterMetrics Sub-Schema**:
```yaml
task_count: 2
running_tasks: 2
pending_tasks: 0
cpu_utilization_percent: 45.2
memory_utilization_percent: 52.8
```

**ALBMetrics Sub-Schema**:
```yaml
request_count: 1523
target_response_time_seconds: 0.245
healthy_target_count: 2
unhealthy_target_count: 0
http_5xx_count: 0
http_4xx_count: 12
```

**RDSMetrics Sub-Schema**:
```yaml
cpu_utilization_percent: 32.1
database_connections: 8
free_storage_mb: 18432
read_iops: 45
write_iops: 12
read_latency_ms: 2.3
write_latency_ms: 3.1
```

**Example**:
```yaml
environment: staging
timestamp: "2026-07-03T15:30:00Z"
ecs_cluster_metrics:
  task_count: 2
  running_tasks: 2
  cpu_utilization_percent: 45.2
alb_metrics:
  request_count: 1523
  target_response_time_seconds: 0.245
  healthy_target_count: 2
rds_metrics:
  cpu_utilization_percent: 32.1
  database_connections: 8
  free_storage_mb: 18432
```

---

### 9. CostAllocationReport

**Description**: Monthly AWS cost breakdown by environment and service.

**Attributes**:
- `report_month` (string, required): Billing period — `"2026-07"` (YYYY-MM)
- `environment` (string, required): Environment name
- `total_cost_usd` (decimal, required): Total monthly spend
- `service_costs` ([]ServiceCost, required): Cost by AWS service
- `budget_usd` (integer, required): Monthly budget ceiling
- `budget_utilization_percent` (decimal, computed): `(total_cost / budget) * 100`

**ServiceCost Sub-Schema**:
```yaml
service_name: "ECS|RDS|S3|CloudFront|ALB|NAT|CloudWatch"
cost_usd: 45.23
usage_quantity: "512 task-hours"  # service-specific unit
```

**Example**:
```yaml
report_month: "2026-07"
environment: staging
total_cost_usd: 178.45
service_costs:
  - service_name: ECS
    cost_usd: 42.15
    usage_quantity: "512 task-hours"
  - service_name: RDS
    cost_usd: 14.46
    usage_quantity: "720 instance-hours"
  - service_name: ALB
    cost_usd: 18.72
    usage_quantity: "1.2M requests"
  - service_name: S3
    cost_usd: 2.34
    usage_quantity: "50 GB-month"
  - service_name: CloudFront
    cost_usd: 8.12
    usage_quantity: "100 GB data transfer"
  - service_name: NAT
    cost_usd: 57.00
    usage_quantity: "720 instance-hours"
  - service_name: CloudWatch
    cost_usd: 35.66
    usage_quantity: "200 GB logs ingested"
budget_usd: 200
budget_utilization_percent: 89.2
```

---

## Entity Relationships

```
Environment (1) ──has──> (1) VPCConfiguration
                ──has──> (1) ECSClusterConfiguration
                ──has──> (1) RDSInstanceConfiguration
                ──has──> (1) LoadBalancerConfiguration
                ──generates──> (0..*) InfrastructureMetrics
                ──generates──> (0..*) CostAllocationReport

ECSClusterConfiguration (1) ──deploys──> (0..*) DeploymentArtifact [backend]
LoadBalancerConfiguration (1) ──routes_to──> (1) ECSClusterConfiguration

TerraformWorkspace (1) ──manages──> (1) Environment
```

---

## Validation Rules

1. **Environment uniqueness**: `name` must be unique across all environments
2. **CIDR non-overlap**: VPC CIDR blocks must not overlap between staging and production
3. **Subnet allocation**: Each environment must have at least 2 public subnets and 2 private subnets in different AZs
4. **Cost budget**: `CostAllocationReport.total_cost_usd` must not exceed `Environment.cost_budget_usd` by more than 20% without alert
5. **Task sizing**: ECS task `cpu` and `memory` must follow valid Fargate combinations (see AWS documentation)
6. **RDS Multi-AZ**: Production environment must have `RDSInstanceConfiguration.multi_az = true`
7. **Artifact immutability**: `DeploymentArtifact.git_sha` and `version` cannot be changed after creation
8. **Terraform workspace isolation**: `TerraformWorkspace.state_key` must be unique per workspace
9. **Security group rules**: `ecs-sg` ingress must only allow traffic from `alb-sg`; `rds-sg` ingress must only allow traffic from `ecs-sg`
10. **Health check path**: ALB target group `health_check.path` must match backend `/healthz` endpoint

---

## Notes

- This data model describes **infrastructure configuration and state**, not application data (which is in spec 001)
- All timestamps are UTC in ISO 8601 format
- Cost values are in USD with 2 decimal precision
- ARNs are AWS-generated identifiers and vary by region and account
- The `is_active` flag on `Environment` determines whether Terraform provisions resources or only validates configuration
- Production infrastructure remains dormant (`is_provisioned: false`) until alpha release approval
