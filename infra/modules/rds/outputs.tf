# RDS Module Outputs
#
# Exposes the database endpoint, database name, and the ARN of the
# Secrets Manager secret holding the generated credentials, consumed by
# the ECS module and downstream root module wiring.

output "db_endpoint" {
  description = "Connection endpoint of the RDS instance"
  value       = aws_db_instance.main.endpoint
}

output "db_name" {
  description = "Name of the default database on the RDS instance"
  value       = aws_db_instance.main.db_name
}

output "db_secret_arn" {
  description = "ARN of the Secrets Manager secret holding the DB credentials"
  value       = aws_secretsmanager_secret.db_credentials.arn
  sensitive   = true
}
