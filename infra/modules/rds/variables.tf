# RDS Module Variables
#
# Input variables for the RDS PostgreSQL module: environment name, network
# placement (VPC/private subnets), the ECS security group allowed to reach
# the database, instance sizing/availability, backup retention, and the
# database name/username.

variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}

variable "vpc_id" {
  description = "VPC identifier where the RDS security group is created"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for the RDS subnet group"
  type        = list(string)
}

variable "ecs_security_group_id" {
  description = "Security group ID of the ECS service allowed to reach RDS on port 5432"
  type        = string
}

variable "instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro" # staging: db.t4g.micro, production: db.t4g.small
}

variable "multi_az" {
  description = "Enable Multi-AZ deployment for high availability"
  type        = bool
  default     = false # staging: false, production: true
}

variable "backup_retention_days" {
  description = "Number of days to retain automated backups"
  type        = number
  default     = 1 # staging: 1, production: 30
}

variable "db_name" {
  description = "Name of the default database created on the RDS instance"
  type        = string
  default     = "traveler"
}

variable "db_username" {
  description = "Master username for the RDS instance"
  type        = string
  default     = "traiveler_admin"
}
