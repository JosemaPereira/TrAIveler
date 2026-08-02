# Secrets Module Outputs
#
# Exposes the ARNs of the AI API key and JWT signing key secrets. Not yet
# wired to any consumer: infra/main.tf doesn't reference module.secrets'
# outputs, so the ECS task role has no IAM read access to these ARNs yet.
# That wiring is tracked separately (003-T033 / G-INFRA-IAM-MODULE, Backlog).

output "ai_api_key_secret_arn" {
  description = "ARN of the Secrets Manager secret holding the AI API key"
  value       = aws_secretsmanager_secret.ai_api_key.arn
  sensitive   = true
}

output "jwt_signing_key_secret_arn" {
  description = "ARN of the Secrets Manager secret holding the JWT signing key"
  value       = aws_secretsmanager_secret.jwt_signing_key.arn
  sensitive   = true
}
