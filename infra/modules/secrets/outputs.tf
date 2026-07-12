# Secrets Module Outputs
#
# Exposes the ARNs of the AI API key and JWT signing key secrets, consumed
# by the ECS module and downstream root module wiring to grant task-level
# read access via IAM.

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
