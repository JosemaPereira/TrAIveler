# Root Module Variables
#
# All environment-tunable inputs for the root module: environment identity,
# VPC/NAT sizing, ECS Fargate sizing, RDS sizing, and the ACM certificates
# used for HTTPS on the ALB (backend) and CloudFront (frontend) modules.
# frontend_domain is the public domain name the frontend certificate covers
# and is wired into CloudFront's alias. There is no backend_domain: the alb
# module has no domain_name-equivalent input to wire it to (only
# certificate_arn) — add it back once a DNS/Route 53 resource for the
# backend actually consumes it.
# Defaults (where present) match the staging profile; production overrides
# every one of them via environments/production.tfvars.

variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
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

variable "task_cpu" {
  description = "Fargate task-level vCPU units (e.g. \"256\", \"1024\")"
  type        = string
  default     = "256" # staging: 256, production: 1024
}

variable "task_memory" {
  description = "Fargate task-level memory in MiB (must be a valid pairing with task_cpu)"
  type        = string
  default     = "512" # staging: 512, production: 2048
}

variable "min_tasks" {
  description = "Minimum number of running ECS tasks (autoscaling lower bound)"
  type        = number
  default     = 1 # staging: 1, production: 2
}

variable "max_tasks" {
  description = "Maximum number of running ECS tasks (autoscaling upper bound)"
  type        = number
  default     = 2 # staging: 2, production: 20
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs for the ECS service"
  type        = number
  default     = 7 # staging: 7, production: 30
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro" # staging: db.t4g.micro, production: db.t4g.small
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment for the RDS instance"
  type        = bool
  default     = false # staging: false, production: true
}

variable "backup_retention_days" {
  description = "Number of days to retain automated RDS backups"
  type        = number
  default     = 1 # staging: 1, production: 30
}

variable "frontend_domain" {
  description = "Public domain name for the frontend SPA, served as the CloudFront distribution alias"
  type        = string
}

# ---------------------------------------------------------------------------
# Additive variables (not in the flat list in issue #90 / 005-T103) —
# required, no-default inputs to close the dependency graph, following the
# same pattern already used inside the ecs/rds/alb modules (see e.g.
# infra/modules/ecs/variables.tf's own "Additive variables" section): they
# are the minimum necessary to actually wire modules whose variables.tf
# require values this root module is the only place that can supply.
# ---------------------------------------------------------------------------

variable "ecr_repository_url" {
  description = "ECR repository URL, including image tag, for the backend container image (consumed by the ecs module; the ECR repository itself is provisioned outside this root module)"
  type        = string
}

variable "backend_certificate_arn" {
  description = "ACM certificate ARN for the backend API's domain, used for HTTPS termination on the ALB listener (provisioned outside this root module)"
  type        = string
}

variable "frontend_certificate_arn" {
  description = "ACM certificate ARN for frontend_domain, used for HTTPS on the CloudFront distribution — must be issued in us-east-1 (provisioned outside this root module)"
  type        = string
}
