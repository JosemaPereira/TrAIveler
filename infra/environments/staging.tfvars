# Staging Environment Configuration
#
# Cost-optimized profile (~$200/month target): NAT instance instead of NAT
# Gateway, minimal Fargate task sizing, single-AZ RDS, short log/backup
# retention. See docs/cloud-and-environments.md for the full cost strategy
# and infra/README.md's "Environment Configurations" section.

environment        = "staging"
vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-east-1a", "us-east-1b"]
enable_nat_gateway = false

# ECS configuration
task_cpu           = "256" # 0.25 vCPU
task_memory        = "512" # 0.5 GB RAM
min_tasks          = 1
max_tasks          = 2
log_retention_days = 7

# RDS configuration
instance_class        = "db.t4g.micro"
multi_az              = false
backup_retention_days = 1

# Domain configuration
frontend_domain = "staging.traveler.example.com"

# ECR / ACM configuration
#
# Placeholders: the ECR repository and ACM certificates referenced below are
# not provisioned yet (staging infra is still dormant, per this repo's
# AWS-cost-avoidance policy — see G-SPRINT3-INFRA-CICD / issue #91 and
# infra/README.md's "AWS Account Setup" section). Replace with real values
# before the first real `terraform apply` against this environment.
ecr_repository_url       = "123456789012.dkr.ecr.us-east-1.amazonaws.com/traveler-backend-staging:latest"
backend_certificate_arn  = "arn:aws:acm:us-east-1:123456789012:certificate/REPLACE_WITH_STAGING_BACKEND_CERT"
frontend_certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/REPLACE_WITH_STAGING_FRONTEND_CERT"
