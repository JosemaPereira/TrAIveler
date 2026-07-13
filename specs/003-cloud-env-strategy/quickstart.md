# Quickstart: Cloud & Environments Strategy Validation

**Feature**: [spec.md](spec.md) | **Plan**: [plan.md](plan.md)

**Purpose**: Runnable validation scenarios that prove the cloud infrastructure and deployment pipelines work end-to-end. These scenarios test infrastructure provisioning, deployment workflows, and operational observability without implementing application features.

**Audience**: DevOps engineers, infrastructure reviewers, and anyone validating that the cloud foundation is production-ready.

---

## Prerequisites

Before running any scenario, ensure you have:

1. **AWS Account Access**:
   - AWS CLI configured (`aws configure`)
   - IAM permissions to create VPCs, ECS clusters, RDS instances, ALB, S3 buckets, CloudFront distributions
   - AWS account ID noted for use in commands

2. **Terraform Installed**:
   ```bash
   terraform version  # Should be >= 1.5.0
   ```

3. **Docker Installed** (for backend deployments):
   ```bash
   docker --version
   ```

4. **GitHub Repository Access**:
   - Repository cloned locally
   - GitHub Actions enabled
   - OIDC federation configured (see Scenario 2)

5. **Environment Variables**:
   ```bash
   export AWS_ACCOUNT_ID="123456789012"  # Replace with your account ID
   export AWS_REGION="us-east-1"
   ```

---

## Scenario 1: Bootstrap Terraform Remote State

**Goal**: Create S3 bucket and DynamoDB table for Terraform state management before provisioning any infrastructure.

**Why First**: This is the foundation for all other scenarios. Terraform workspaces require remote state.

**Commands**:

```bash
# 1. Navigate to infrastructure directory
cd infra/terraform

# 2. Run bootstrap script (creates S3 + DynamoDB with proper configuration)
./scripts/bootstrap-state.sh

# Expected Output:
# ✓ S3 bucket created: trAIveler-terraform-state
# ✓ Versioning enabled on bucket
# ✓ Encryption enabled (AES-256)
# ✓ DynamoDB table created: trAIveler-terraform-locks
# ✓ Point-in-time recovery enabled

# 3. Verify S3 bucket exists
aws s3 ls | grep trAIveler-terraform-state

# Expected Output:
# 2026-07-03 10:30:00 trAIveler-terraform-state

# 4. Verify DynamoDB table exists
aws dynamodb describe-table --table-name trAIveler-terraform-locks --query 'Table.TableName'

# Expected Output:
# "trAIveler-terraform-locks"
```

**Validation**:
- S3 bucket `trAIveler-terraform-state` exists
- S3 bucket has versioning enabled
- S3 bucket has default encryption enabled
- DynamoDB table `trAIveler-terraform-locks` exists with `LockID` as hash key

**Rollback** (if needed):
```bash
# Delete DynamoDB table
aws dynamodb delete-table --table-name trAIveler-terraform-locks

# Empty and delete S3 bucket
aws s3 rm s3://trAIveler-terraform-state --recursive
aws s3api delete-bucket --bucket trAIveler-terraform-state
```

---

## Scenario 2: Configure GitHub OIDC Federation

**Goal**: Set up GitHub Actions to authenticate to AWS without long-lived credentials.

**Why Second**: Required before any CI/CD workflows can run terraform apply or deploy containers.

**Commands**:

```bash
# 1. Create OIDC identity provider in AWS
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1

# Expected Output:
# {
#     "OpenIDConnectProviderArn": "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"
# }

# 2. Create IAM role for GitHub Actions (staging environment)
aws iam create-role \
  --role-name GitHubActionsRole-Staging \
  --assume-role-policy-document file://infra/iam/github-actions-trust-policy-staging.json

# 3. Attach policies to role
aws iam attach-role-policy \
  --role-name GitHubActionsRole-Staging \
  --policy-arn arn:aws:iam::aws:policy/AmazonECS_FullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsRole-Staging \
  --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryPowerUser

aws iam attach-role-policy \
  --role-name GitHubActionsRole-Staging \
  --policy-arn arn:aws:iam::aws:policy/AWSCloudFormationFullAccess

# 4. Verify role exists
aws iam get-role --role-name GitHubActionsRole-Staging --query 'Role.Arn'

# Expected Output:
# "arn:aws:iam::123456789012:role/GitHubActionsRole-Staging"
```

**Trust Policy** (`infra/iam/github-actions-trust-policy-staging.json`):
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::${AWS_ACCOUNT_ID}:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:${GITHUB_ORG}/${GITHUB_REPO}:ref:refs/heads/main"
        }
      }
    }
  ]
}
```

**Validation**:
- OIDC provider exists in IAM
- IAM role `GitHubActionsRole-Staging` exists with correct trust policy
- GitHub Actions can assume the role (test in Scenario 5)

---

## Scenario 3: Provision Staging Environment

**Goal**: Use Terraform to provision the complete staging infrastructure from scratch.

**Duration**: ~15-20 minutes (includes RDS initialization)

**Commands**:

```bash
# 1. Navigate to Terraform directory
cd infra/terraform

# 2. Initialize Terraform with remote state backend
terraform init

# Expected Output:
# Initializing the backend...
# Successfully configured the backend "s3"!
# ...
# Terraform has been successfully initialized!

# 3. Create and select staging workspace
terraform workspace new staging
terraform workspace select staging

# Expected Output:
# Created and switched to workspace "staging"!

# 4. Validate Terraform configuration
terraform validate

# Expected Output:
# Success! The configuration is valid.

# 5. Review planned infrastructure changes
terraform plan -var-file=staging.tfvars -out=staging.tfplan

# Expected Output (summary):
# Plan: 45 to add, 0 to change, 0 to destroy.

# 6. Apply infrastructure (provision all resources)
terraform apply staging.tfplan

# Expected Output (after 15-20 minutes):
# Apply complete! Resources: 45 added, 0 changed, 0 destroyed.
# 
# Outputs:
# alb_dns_name = "trAIveler-alb-staging-1234567890.us-east-1.elb.amazonaws.com"
# ecs_cluster_name = "trAIveler-staging"
# rds_endpoint = "trAIveler-staging.abc123def456.us-east-1.rds.amazonaws.com:5432"
# ...

# 7. Export outputs for later use
terraform output -json > staging-outputs.json
```

**Validation**:
- VPC exists with CIDR `10.0.0.0/16`
- 2 public subnets and 2 private subnets created in different AZs
- NAT instance running in public subnet (cheaper than NAT Gateway)
- ALB created and listening on ports 80 (redirect) and 443
- ECS cluster created with zero running tasks (no deployment yet)
- RDS instance `trAIveler-staging` available (db.t4g.micro, single-AZ)
- S3 bucket `trAIveler-frontend-staging` created
- CloudFront distribution created (may take 10-15 minutes to fully deploy)

**Cost Estimate**: ~$6.50/day (~$200/month) based on [research.md](research.md) analysis.

---

## Scenario 4: Test Infrastructure Health Checks

**Goal**: Verify all infrastructure components are operational and accessible.

**Commands**:

```bash
# 1. Get ALB DNS name from Terraform outputs
export ALB_DNS=$(terraform output -raw alb_dns_name)
echo "ALB DNS: $ALB_DNS"

# 2. Test ALB HTTP → HTTPS redirect
curl -I http://$ALB_DNS

# Expected Output:
# HTTP/1.1 301 Moved Permanently
# Location: https://...

# 3. Test ALB HTTPS listener (will fail with 503 until backend is deployed)
curl -k https://$ALB_DNS/healthz

# Expected Output (before backend deployment):
# <html><body><h1>503 Service Temporarily Unavailable</h1>
# No server is available to handle this request.
# </body></html>

# 4. Test RDS connectivity from local machine (requires security group modification or bastion host)
# For now, verify RDS is accepting connections via AWS Console or CLI
export RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
aws rds describe-db-instances --db-instance-identifier trAIveler-staging --query 'DBInstances[0].DBInstanceStatus'

# Expected Output:
# "available"

# 5. Test S3 bucket accessibility
aws s3 ls s3://trAIveler-frontend-staging/

# Expected Output (empty bucket):
# (no output - bucket is empty)

# 6. Test CloudFront distribution status
export CF_DIST_ID=$(terraform output -raw cloudfront_distribution_id)
aws cloudfront get-distribution --id $CF_DIST_ID --query 'Distribution.Status'

# Expected Output:
# "Deployed"
```

**Validation**:
- ALB responds on port 80 with HTTP 301 redirect to HTTPS
- ALB port 443 listener is active (503 is expected without backend)
- RDS instance status is "available"
- S3 bucket exists and is accessible
- CloudFront distribution status is "Deployed"

---

## Scenario 5: Deploy Backend to ECS via GitHub Actions

**Goal**: Trigger CI/CD pipeline to build backend Docker image and deploy to ECS Fargate.

**Prerequisites**:
- Scenario 1-3 completed (infrastructure provisioned)
- Scenario 2 completed (OIDC federation configured)
- Backend code exists in `backend/` directory with working Dockerfile

**Commands**:

```bash
# 1. Create GitHub repository secrets/variables (do this via GitHub UI or CLI)
gh secret set AWS_ROLE_ARN_STAGING --body "arn:aws:iam::${AWS_ACCOUNT_ID}:role/GitHubActionsRole-Staging"
gh variable set ECR_REPOSITORY_NAME --body "trAIveler-backend"
gh variable set ECS_CLUSTER_NAME_STAGING --body "trAIveler-staging"
gh variable set ECS_SERVICE_NAME_STAGING --body "trAIveler-backend-staging"

# 2. Store secrets in AWS Secrets Manager (required by ECS tasks)
# Naming convention: traveler-${environment}-${secret_name} — matches the shipped
# infra/modules/secrets/main.tf and infra/modules/rds/main.tf (Sprint 3, issue #89).
# traveler-staging-db-credentials is created and auto-populated by the RDS module itself
# (JSON: {"username": ..., "password": ...}) — do not recreate it here.
aws secretsmanager create-secret \
  --name traveler-staging-ai-api-key \
  --secret-string "sk-ant-api03-YOUR_KEY_HERE"

aws secretsmanager create-secret \
  --name traveler-staging-jwt-signing-key \
  --secret-string "$(cat jwt-signing-key.pem)"

# 3. Push code to main branch to trigger deployment workflow
git add backend/
git commit -m "feat: initial backend implementation"
git push origin main

# 4. Monitor GitHub Actions workflow
gh run watch

# Expected Output (after 5-10 minutes):
# ✓ Build backend Docker image (linux/arm64)
# ✓ Push image to ECR
# ✓ Update ECS task definition
# ✓ Deploy to ECS service
# ✓ Wait for service stability
# ✓ All tasks healthy

# 5. Verify ECS service is running
aws ecs describe-services \
  --cluster trAIveler-staging \
  --services trAIveler-backend-staging \
  --query 'services[0].runningCount'

# Expected Output:
# 1

# 6. Test backend health check via ALB
export ALB_DNS=$(cd infra/terraform && terraform output -raw alb_dns_name)
curl -k https://$ALB_DNS/healthz

# Expected Output:
# {
#   "status": "ok",
#   "version": "1.0.0",
#   "uptime_seconds": 45
# }
```

**Validation**:
- ECR repository contains Docker image with tag `v1.0.0-<sha>`
- ECS task definition registered with correct image URI and secrets
- ECS service has 1 running task
- ALB target group shows 1 healthy target
- `GET /healthz` returns HTTP 200 with valid JSON

**Duration**: ~5-10 minutes (Docker build + ECS deployment)

---

## Scenario 6: Deploy Frontend to S3 + CloudFront

**Goal**: Build frontend static assets and deploy to S3, then invalidate CloudFront cache.

**Prerequisites**:
- Scenario 1-3 completed (infrastructure provisioned)
- Frontend code exists in `frontend/` directory

**Commands**:

```bash
# 1. Set GitHub variables for frontend deployment
export ALB_DNS=$(cd infra/terraform && terraform output -raw alb_dns_name)
gh variable set API_URL_STAGING --body "https://$ALB_DNS"
gh variable set S3_BUCKET_STAGING --body "trAIveler-frontend-staging"
gh variable set CLOUDFRONT_DISTRIBUTION_ID_STAGING --body "$(cd infra/terraform && terraform output -raw cloudfront_distribution_id)"

# 2. Trigger frontend deployment (automatic on push to main)
git add frontend/
git commit -m "feat: initial frontend implementation"
git push origin main

# 3. Monitor GitHub Actions workflow
gh run watch --repo <org>/<repo> --workflow=frontend-deploy.yml

# Expected Output (after 3-5 minutes):
# ✓ Install dependencies
# ✓ Build production bundle
# ✓ Sync to S3
# ✓ Invalidate CloudFront cache
# ✓ Deployment complete

# 4. Verify S3 bucket contains build artifacts
aws s3 ls s3://trAIveler-frontend-staging/

# Expected Output:
# 2026-07-03 10:30:00   1234 index.html
# 2026-07-03 10:30:00    456 assets/...

# 5. Test frontend via CloudFront
export CF_DOMAIN=$(cd infra/terraform && terraform output -raw cloudfront_domain_name)
curl -I https://$CF_DOMAIN

# Expected Output:
# HTTP/2 200
# content-type: text/html
# ...
```

**Validation**:
- S3 bucket contains `index.html` and `assets/` directory
- CloudFront invalidation completed
- Frontend accessible via CloudFront domain
- Frontend makes successful API calls to backend ALB

**Duration**: ~3-5 minutes (build + S3 sync + CloudFront invalidation)

---

## Scenario 7: Test Auto-Scaling

**Goal**: Verify ECS Fargate auto-scaling triggers when CPU exceeds 70%.

**Prerequisites**:
- Scenario 5 completed (backend deployed to ECS)

**Commands**:

```bash
# 1. Generate load on backend to trigger CPU scaling
# Install k6 load testing tool first (from spec 002 research)
k6 run tests/load/k6/baseline.js --env BASE_URL=https://$(cd infra/terraform && terraform output -raw alb_dns_name)

# OR manually generate load:
export ALB_DNS=$(cd infra/terraform && terraform output -raw alb_dns_name)
for i in {1..1000}; do
  curl -k https://$ALB_DNS/healthz &
done

# 2. Monitor ECS service task count
watch -n 5 'aws ecs describe-services --cluster trAIveler-staging --services trAIveler-backend-staging --query "services[0].runningCount"'

# Expected Output (over 5-10 minutes):
# 1  # Initial count
# 2  # After ~2 minutes of high CPU
# 3  # Scales up to handle load
# ...
# Max 5 tasks (staging limit)

# 3. Stop load generation (Ctrl+C)
# Wait 5-10 minutes for scale-in cooldown

# 4. Verify tasks scale back down to minimum (1)
aws ecs describe-services --cluster trAIveler-staging --services trAIveler-backend-staging --query 'services[0].runningCount'

# Expected Output (after cooldown):
# 1
```

**Validation**:
- ECS service scales from 1 to 2+ tasks when CPU > 70%
- Maximum task count respects `ecs_max_capacity` limit (5 for staging)
- Tasks scale back down to `ecs_min_capacity` (1) after load subsides
- All tasks remain healthy during scaling events

**Duration**: ~15-20 minutes (scale-up + scale-down)

---

## Scenario 8: Test Deployment Rollback

**Goal**: Verify ECS deployment circuit breaker automatically rolls back failed deployments.

**Prerequisites**:
- Scenario 5 completed (backend deployed to ECS)

**Commands**:

```bash
# 1. Intentionally break the backend (e.g., bad health check response)
# Modify backend code to return HTTP 500 from /healthz endpoint
# Commit and push to trigger deployment

# 2. Monitor ECS deployment
aws ecs describe-services --cluster trAIveler-staging --services trAIveler-backend-staging --query 'services[0].deployments'

# Expected Output (during failed deployment):
# [
#   {
#     "status": "PRIMARY",
#     "taskDefinition": "...:10",  # Old healthy version
#     "desiredCount": 1,
#     "runningCount": 1
#   },
#   {
#     "status": "ACTIVE",
#     "taskDefinition": "...:11",  # New failing version
#     "desiredCount": 1,
#     "runningCount": 0,
#     "failedTasks": 1,
#     "rolloutState": "FAILED"
#   }
# ]

# 3. Wait for circuit breaker to trigger (60-90 seconds)
# ECS will automatically stop the failed deployment and keep old tasks running

# 4. Verify old tasks still running and healthy
aws ecs describe-services --cluster trAIveler-staging --services trAIveler-backend-staging --query 'services[0].runningCount'

# Expected Output:
# 1  # Old healthy tasks still running

# 5. Test ALB still routes to healthy tasks
curl -k https://$(cd infra/terraform && terraform output -raw alb_dns_name)/healthz

# Expected Output (old healthy version still responding):
# {
#   "status": "ok",
#   "version": "1.0.0",
#   "uptime_seconds": 1234
# }
```

**Validation**:
- Failed deployment is detected after health check grace period (60 seconds)
- Circuit breaker stops deployment and marks it as FAILED
- Old healthy tasks remain running
- No service downtime occurs
- ALB continues routing traffic to healthy tasks

**Duration**: ~2-3 minutes

---

## Scenario 9: Validate Production Configuration (Dry-Run)

**Goal**: Verify production Terraform configuration is valid without provisioning resources.

**Why Important**: Ensures production can be deployed when needed for alpha release.

**Commands**:

```bash
# 1. Navigate to Terraform directory
cd infra/terraform

# 2. Create production workspace (if not exists)
terraform workspace new production || terraform workspace select production

# 3. Validate production configuration
terraform validate

# Expected Output:
# Success! The configuration is valid.

# 4. Run terraform plan (no apply)
terraform plan -var-file=production.tfvars -out=production.tfplan

# Expected Output (summary):
# Plan: 47 to add, 0 to change, 0 to destroy.
#
# Key differences from staging:
# - RDS: Multi-AZ enabled, db.t4g.small instance
# - ECS: 1024 CPU / 2048 memory, max 20 tasks
# - NAT: NAT Gateway instead of NAT instance
# - Backup retention: 30 days instead of 7

# 5. Review outputs (do NOT apply)
terraform show -json production.tfplan | jq '.planned_values.outputs'

# Expected Output (production configuration values):
# {
#   "alb_dns_name": {...},
#   "ecs_cluster_name": {...},  # "trAIveler-production"
#   "rds_endpoint": {...},
#   ...
# }

# 6. Clean up plan file
rm production.tfplan
```

**Validation**:
- Production plan succeeds without errors
- Plan shows ~47 resources to create (slightly more than staging due to Multi-AZ RDS)
- No resources are actually created (`terraform apply` NOT run)
- Production configuration differences validated:
  - VPC CIDR: `10.1.0.0/16` (different from staging)
  - RDS Multi-AZ: enabled
  - ECS task size: 1 vCPU / 2 GB
  - NAT: Gateway (not instance)

**Cost Estimate**: ~$13/day (~$400/month) when provisioned.

---

## Scenario 10: Monitor Costs and Budget Alerts

**Goal**: Verify AWS Cost Explorer shows environment-tagged spending and budget alerts trigger.

**Prerequisites**:
- Scenario 3 completed (staging provisioned for at least 24 hours)

**Commands**:

```bash
# 1. Query AWS Cost Explorer for staging environment costs
aws ce get-cost-and-usage \
  --time-period Start=2026-07-01,End=2026-07-31 \
  --granularity DAILY \
  --metrics "UnblendedCost" \
  --group-by Type=TAG,Key=Environment \
  --filter file://cost-filter-staging.json

# cost-filter-staging.json:
# {
#   "Tags": {
#     "Key": "Environment",
#     "Values": ["staging"]
#   }
# }

# Expected Output (after 24 hours):
# {
#   "ResultsByTime": [
#     {
#       "TimePeriod": {...},
#       "Total": {},
#       "Groups": [
#         {
#           "Keys": ["Environment$staging"],
#           "Metrics": {
#             "UnblendedCost": {
#               "Amount": "6.50",
#               "Unit": "USD"
#             }
#           }
#         }
#       ]
#     }
#   ]
# }

# 2. Verify budget alert is configured
aws budgets describe-budgets --account-id $AWS_ACCOUNT_ID --query 'Budgets[?BudgetName==`TrAIveler-Staging-Monthly`]'

# Expected Output:
# [
#   {
#     "BudgetName": "TrAIveler-Staging-Monthly",
#     "BudgetLimit": {
#       "Amount": "200",
#       "Unit": "USD"
#     },
#     "TimeUnit": "MONTHLY"
#   }
# ]

# 3. Test budget alert threshold
# If costs exceed 80% ($160), SNS notification should be sent
# This can only be validated manually after costs accumulate
```

**Validation**:
- Cost Explorer shows daily costs broken down by `Environment` tag
- Staging costs are approximately $6-7/day
- Budget alert exists for $200/month ceiling
- SNS topic configured for 80% and 100% threshold notifications

**Duration**: Requires 24+ hours of resource usage for meaningful cost data.

---

## Troubleshooting

### Issue: Terraform state locked

**Symptom**:
```
Error: Error locking state: Error acquiring the state lock
```

**Resolution**:
```bash
# Force unlock (use with caution)
terraform force-unlock <LOCK_ID>

# Or delete DynamoDB lock item manually
aws dynamodb delete-item \
  --table-name trAIveler-terraform-locks \
  --key '{"LockID":{"S":"trAIveler-terraform-state/env/staging/terraform.tfstate-md5"}}'
```

---

### Issue: ECS tasks failing health checks

**Symptom**: ECS service shows unhealthy tasks, ALB returns 503.

**Resolution**:
```bash
# 1. Check ECS task logs
aws logs tail /ecs/trAIveler-backend-staging --follow

# 2. Verify secrets are accessible
aws secretsmanager get-secret-value --secret-id traveler-staging-db-credentials

# 3. Check security group rules
aws ec2 describe-security-groups --group-ids <ecs-sg-id>

# 4. Manually test health check from within VPC (requires bastion host)
curl http://<task-private-ip>:8080/healthz
```

---

### Issue: CloudFront invalidation takes too long

**Symptom**: Frontend updates not visible after deployment.

**Resolution**:
```bash
# 1. Check invalidation status
aws cloudfront get-invalidation \
  --distribution-id <dist-id> \
  --id <invalidation-id>

# 2. Force cache clear (wait 10-15 minutes)
# CloudFront invalidations typically complete in 10-15 minutes

# 3. Bypass cache for testing
curl https://<cloudfront-domain>/ -H "Cache-Control: no-cache"
```

---

### Issue: RDS connection timeout

**Symptom**: Backend cannot connect to database.

**Resolution**:
```bash
# 1. Verify RDS security group allows ECS traffic
aws ec2 describe-security-groups --group-ids <rds-sg-id> \
  --query 'SecurityGroups[0].IpPermissions'

# 2. Verify ECS tasks are in correct subnets
aws ecs describe-tasks --cluster trAIveler-staging --tasks <task-arn> \
  --query 'tasks[0].attachments[0].details'

# 3. Test connectivity from ECS task (requires exec into task)
aws ecs execute-command --cluster trAIveler-staging --task <task-id> --command "nc -zv <rds-endpoint> 5432"
```

---

## Next Steps

After completing all scenarios:

1. **Update Roadmap**: Mark infrastructure tasks as complete in `docs/roadmap.md`
2. **Document Production Cutover Plan**: Create runbook for activating production environment
3. **Set Up Monitoring Dashboards**: Create CloudWatch dashboards for operational metrics
4. **Configure Alerting**: Set up PagerDuty or SNS notifications for critical infrastructure alerts
5. **Proceed to Application Development**: Begin implementing user stories from spec 001

---

## Reference

- **Spec**: [spec.md](spec.md)
- **Research**: [research.md](research.md)
- **Data Model**: [data-model.md](data-model.md)
- **Contracts**: [contracts/infrastructure.md](contracts/infrastructure.md)
- **Terraform Modules**: `infra/terraform/modules/`
- **GitHub Workflows**: `.github/workflows/`
