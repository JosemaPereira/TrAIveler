# RDS Module Outputs
#
# Exposes the database endpoint, database name, and the ARN of the
# Secrets Manager secret holding the generated credentials. Not yet wired to
# any consumer: infra/main.tf doesn't pass these to the ecs module or expose
# them as root outputs, so the backend has no way to receive DATABASE_URL or
# read db_secret_arn via IAM yet. That wiring is future work (see 003-T033 /
# G-INFRA-IAM-MODULE for the IAM half).

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
