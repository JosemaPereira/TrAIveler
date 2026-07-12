# ECS Module Variables
#
# Input variables for the ECS Fargate cluster/service module: environment
# name, network placement (VPC/private subnets), Fargate task sizing,
# autoscaling bounds, log retention, and the ECR image to run.
#
# Two variables beyond the flat list in the originating issue/tasks.md are
# also declared here (alb_security_group_id, target_group_arn) — see the
# top-of-file comment in main.tf for why they are required, not scope creep.

variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}

variable "vpc_id" {
  description = "VPC identifier where the ECS security group is created"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs the ECS service's tasks run in"
  type        = list(string)
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
  description = "Minimum number of running tasks (autoscaling lower bound)"
  type        = number
  default     = 1 # staging: 1, production: 2
}

variable "max_tasks" {
  description = "Maximum number of running tasks (autoscaling upper bound)"
  type        = number
  default     = 2 # staging: 2, production: 20
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs for the ECS service"
  type        = number
  default     = 7 # staging: 7, production: 30
}

variable "ecr_repository_url" {
  description = "ECR repository URL (including tag) for the backend container image"
  type        = string
}

# ---------------------------------------------------------------------------
# Additive variables (not in the original flat list) — required to close
# the dependency graph with the ALB module. See main.tf's top-of-file
# comment for the full rationale.
# ---------------------------------------------------------------------------

variable "alb_security_group_id" {
  description = "Security group ID of the ALB, allowed to reach the ECS service on port 8080 (consumed from the ALB module's alb_security_group_id output)"
  type        = string
}

variable "target_group_arn" {
  description = "ARN of the ALB target group the ECS service registers its tasks with (consumed from the ALB module's target_group_arn output)"
  type        = string
}
