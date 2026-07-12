# ALB Module Variables
#
# Input variables for the Application Load Balancer module: environment
# name, network placement (VPC/public subnets), and the ACM certificate used
# for HTTPS termination.
#
# No ecs_security_group_id variable: main.tf's aws_security_group.alb already
# uses open ("all outbound") egress rather than scoping it to the ECS
# security group (see the rationale comment preceding that resource in
# main.tf) — an ecs_security_group_id input was declared here previously
# but never referenced by any resource. It was removed during root-module
# wiring (005-T102 / issue #90) because passing it would have required
# module.ecs's output as this module's input while ecs's own
# alb_security_group_id/target_group_arn variables require this module's
# output in return — a real circular module dependency Terraform cannot
# resolve, for a value nothing in this module actually used.

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
