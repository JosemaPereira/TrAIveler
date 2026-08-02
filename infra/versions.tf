# Terraform and Provider Version Constraints
#
# Defines minimum required versions for Terraform core and AWS provider.
# These constraints ensure compatibility and prevent breaking changes from
# automatic updates.

terraform {
  # 1.5 is the floor to keep import blocks, check blocks, and improved
  # remote state data sources available, in case future work needs them.
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source = "hashicorp/aws"

      # Capped below 6.x: that major bump could break existing CloudFront
      # function, ECS task definition, or RDS cluster resource configs.
      version = "~> 5.0"
    }

    random = {
      source = "hashicorp/random"

      # Used by the RDS module to generate the database master password
      # (random_password), so a real secret value is never hardcoded.
      version = "~> 3.0"
    }
  }
}

# Region and authentication are configured via:
# - Environment variables: AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
# - AWS profiles: AWS_PROFILE
# - IAM role assumption (for CI/CD via OIDC)
provider "aws" {
  region = var.aws_region

  # Enables cost tracking, ownership identification, and compliance.
  default_tags {
    tags = {
      Project     = "TrAIveler"
      ManagedBy   = "Terraform"
      Environment = terraform.workspace
    }
  }
}

variable "aws_region" {
  description = "AWS region for resource deployment"
  type        = string
  default     = "us-east-1"
}
