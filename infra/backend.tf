# Terraform Backend Configuration
#
# Configures remote state storage in AWS S3 with DynamoDB locking.
# This enables team collaboration and prevents concurrent state modifications.
#
# Prerequisites:
# 1. AWS S3 bucket: traveler-terraform-state (us-east-1)
#    - Versioning enabled
#    - Encryption enabled (AES-256)
#    - Block public access
# 2. AWS DynamoDB table: traveler-terraform-locks (us-east-1)
#    - Primary key: LockID (String)
#    - On-demand billing mode
#
# Setup: Create these resources manually once per AWS account before running terraform init.

terraform {
  backend "s3" {
    # S3 bucket for state storage
    bucket = "traveler-terraform-state"

    # State file path (workspace-specific via terraform workspace)
    # Staging: terraform.tfstate.d/staging/terraform.tfstate
    # Production: terraform.tfstate.d/production/terraform.tfstate
    key = "terraform.tfstate"

    # AWS region
    region = "us-east-1"

    # DynamoDB table for state locking
    # Prevents concurrent modifications by multiple users/CI jobs
    dynamodb_table = "traveler-terraform-locks"

    # Enable encryption at rest (S3 server-side encryption)
    encrypt = true

    # Workspace prefix for state isolation
    # Creates separate state files per workspace in S3
    workspace_key_prefix = "workspaces"
  }
}
