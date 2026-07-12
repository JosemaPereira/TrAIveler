# Secrets Manager Module
#
# Creates empty AWS Secrets Manager secrets for the two application secrets
# that are populated manually, post-apply, via `aws secretsmanager
# put-secret-value` (see infra/README.md's "Secrets Population" section):
# the Anthropic AI API key and the JWT RS256 signing key.
#
# DB credentials are NOT created here — the RDS module
# (infra/modules/rds/main.tf) already creates and owns a
# `traveler-${var.environment}-db-credentials` secret, auto-populated via
# the random provider. Duplicating that here would create a second,
# conflicting source of truth for database credentials.
#
# These two secrets are deliberately created with no
# `aws_secretsmanager_secret_version` — they hold no value until a human
# populates them after `terraform apply`.

# ---------------------------------------------------------------------------
# AI API key
# ---------------------------------------------------------------------------

resource "aws_secretsmanager_secret" "ai_api_key" {
  name        = "traveler-${var.environment}-ai-api-key"
  description = "Anthropic AI API key for the ${var.environment} environment (populated manually post-apply)"

  tags = {
    Name        = "${var.environment}-ai-api-key"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# JWT signing key
# ---------------------------------------------------------------------------

resource "aws_secretsmanager_secret" "jwt_signing_key" {
  name        = "traveler-${var.environment}-jwt-signing-key"
  description = "JWT RS256 signing key (PEM) for the ${var.environment} environment (populated manually post-apply)"

  tags = {
    Name        = "${var.environment}-jwt-signing-key"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}
