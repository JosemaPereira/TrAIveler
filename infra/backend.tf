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
    bucket = "traveler-terraform-state"

    # Non-default workspaces are stored under workspace_key_prefix below,
    # e.g. "workspaces/staging/terraform.tfstate" for the staging workspace.
    key = "terraform.tfstate"

    region = "us-east-1"

    # Prevents concurrent state modifications by multiple users/CI jobs.
    dynamodb_table = "traveler-terraform-locks"

    encrypt = true

    workspace_key_prefix = "workspaces"
  }
}
