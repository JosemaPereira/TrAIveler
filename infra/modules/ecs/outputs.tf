# ECS Module Outputs
#
# Exposes cluster/service/task-definition/log-group identifiers, plus this
# module's own security group ID. ecs_security_group_id is additive (not in
# the original flat output list) — infra/modules/rds/variables.tf already
# declares an ecs_security_group_id input expecting to consume exactly this
# value once root-module wiring closes the dependency graph.

output "cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.main.name
}

output "task_definition_family" {
  description = "Family name of the ECS task definition"
  value       = aws_ecs_task_definition.main.family
}

output "log_group_name" {
  description = "Name of the CloudWatch log group used by the ECS service"
  value       = aws_cloudwatch_log_group.ecs.name
}

output "ecs_security_group_id" {
  description = "Security group ID of the ECS service, allowed to reach RDS on port 5432"
  value       = aws_security_group.ecs.id
}
