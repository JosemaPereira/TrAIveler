# Terraform and Provider Version Constraints
#
# Defines minimum required versions for Terraform core and AWS provider.
# These constraints ensure compatibility and prevent breaking changes from
# automatic updates.

terraform {
  # Require Terraform >= 1.5
  # Version 1.5 introduced:
  # - import blocks for managing existing resources
  # - check blocks for continuous validation
  # - improvements to remote state data sources
  required_version = ">= 1.5"

  # AWS Provider configuration
  required_providers {
    aws = {
      source = "hashicorp/aws"
      
      # Use AWS provider ~> 5.0 (any 5.x version, but not 6.x)
      # Provider 5.x brings:
      # - CloudFront function improvements
      # - ECS task definition enhancements
      # - RDS cluster improvements
      # Constraint prevents major version upgrades that could break existing configurations
      version = "~> 5.0"
    }
  }
}

# AWS Provider Configuration
# Region and authentication are configured via:
# - Environment variables: AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
# - AWS profiles: AWS_PROFILE
# - IAM role assumption (for CI/CD via OIDC)
provider "aws" {
  region = var.aws_region

  # Default tags applied to all resources
  # Enables cost tracking, ownership identification, and compliance
  default_tags {
    tags = {
      Project     = "TrAIveler"
      ManagedBy   = "Terraform"
      Environment = terraform.workspace
    }
  }
}

# Variable for AWS region
# Override via terraform.tfvars or -var flag
variable "aws_region" {
  description = "AWS region for resource deployment"
  type        = string
  default     = "us-east-1"
}
