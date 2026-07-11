# ALB Module Variables
#
# Input variables for the Application Load Balancer module: environment
# name, network placement (VPC/public subnets), the ACM certificate used
# for HTTPS termination, and the ECS security group allowed to receive
# forwarded traffic from the ALB.

variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}

variable "vpc_id" {
  description = "VPC identifier where the ALB security group and target group are created"
  type        = string
}

variable "public_subnet_ids" {
  description = "List of public subnet IDs the ALB is deployed into"
  type        = list(string)
}

variable "certificate_arn" {
  description = "ACM certificate ARN used for HTTPS termination on the ALB listener (provisioned outside this module)"
  type        = string
}

variable "ecs_security_group_id" {
  description = "Security group ID of the ECS service allowed to receive forwarded traffic from the ALB on port 8080"
  type        = string
}
