#!/bin/bash
# Terraform remote-state bootstrap script (issue #96, task 003-T009)
#
# Purpose:
#   Idempotently provisions the two AWS resources that infra/backend.tf's S3 backend
#   requires before `terraform init` can succeed:
#     1. S3 bucket for state storage — versioned, AES-256 server-side encrypted, and
#        fully blocked from public access.
#     2. DynamoDB table for state locking — partition key LockID (String), on-demand
#        (PAY_PER_REQUEST) billing.
#
#   Each resource is guarded by an existence check, so re-running this script after
#   the resources already exist is a safe no-op for that resource.
#
# Scope:
#   Exactly these two resources. No IAM/OIDC setup, no other Phase-2 bootstrap tasks —
#   see infra/README.md's CI/CD Integration section for what's still pending.
#
# Usage:
#   ./infra/scripts/bootstrap-state.sh
#
# Requirements:
#   - AWS CLI v2 configured with credentials that can create S3 buckets and
#     DynamoDB tables in us-east-1.
#
# Note: per this repo's AWS-cost-avoidance policy, do not run this script against a
# real AWS account until infra is confirmed ready for it.

set -euo pipefail

# Must match infra/backend.tf exactly.
STATE_BUCKET_NAME="traveler-terraform-state"
LOCK_TABLE_NAME="traveler-terraform-locks"
AWS_REGION_NAME="us-east-1"

bucket_exists() {
  aws s3api head-bucket --bucket "$STATE_BUCKET_NAME" --region "$AWS_REGION_NAME" >/dev/null 2>&1
}

table_exists() {
  aws dynamodb describe-table --table-name "$LOCK_TABLE_NAME" --region "$AWS_REGION_NAME" >/dev/null 2>&1
}

create_state_bucket() {
  echo "Creating S3 bucket: $STATE_BUCKET_NAME"

  if [[ "$AWS_REGION_NAME" == "us-east-1" ]]; then
    # us-east-1 is the one region that rejects an explicit LocationConstraint.
    aws s3api create-bucket \
      --bucket "$STATE_BUCKET_NAME" \
      --region "$AWS_REGION_NAME"
  else
    aws s3api create-bucket \
      --bucket "$STATE_BUCKET_NAME" \
      --region "$AWS_REGION_NAME" \
      --create-bucket-configuration LocationConstraint="$AWS_REGION_NAME"
  fi

  echo "Enabling versioning on: $STATE_BUCKET_NAME"
  aws s3api put-bucket-versioning \
    --bucket "$STATE_BUCKET_NAME" \
    --versioning-configuration Status=Enabled

  echo "Enabling AES-256 server-side encryption on: $STATE_BUCKET_NAME"
  aws s3api put-bucket-encryption \
    --bucket "$STATE_BUCKET_NAME" \
    --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'

  echo "Blocking all public access on: $STATE_BUCKET_NAME"
  aws s3api put-public-access-block \
    --bucket "$STATE_BUCKET_NAME" \
    --public-access-block-configuration \
    BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
}

create_lock_table() {
  echo "Creating DynamoDB table: $LOCK_TABLE_NAME"
  aws dynamodb create-table \
    --table-name "$LOCK_TABLE_NAME" \
    --attribute-definitions AttributeName=LockID,AttributeType=S \
    --key-schema AttributeName=LockID,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    --region "$AWS_REGION_NAME"
}

main() {
  if bucket_exists; then
    echo "S3 bucket $STATE_BUCKET_NAME already exists, skipping creation."
  else
    create_state_bucket
  fi

  if table_exists; then
    echo "DynamoDB table $LOCK_TABLE_NAME already exists, skipping creation."
  else
    create_lock_table
  fi

  echo "Bootstrap complete."
}

# Only run main when this script is executed directly, not when it is sourced
# (e.g. by bootstrap-state.test.sh, which sources it to call main() under a mocked aws).
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
