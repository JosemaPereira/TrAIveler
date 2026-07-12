# Secrets Module Variables
#
# Input variables for the Secrets Manager module: the environment name used
# to build the `traveler-${var.environment}-<name>` secret naming pattern.

variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}
