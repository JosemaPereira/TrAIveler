# VPC Module Outputs
#
# Exposes the VPC identifier and subnet ID lists consumed by downstream
# modules (ALB in public subnets; ECS and RDS in private subnets).

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
