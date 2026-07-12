# Production Environment Configuration
#
# Production-sized profile (~$300-400/month target when active): NAT Gateway
# for reliability, full Fargate task sizing, Multi-AZ RDS, longer log/backup
# retention. See docs/cloud-and-environments.md for the full cost strategy
# and infra/README.md's "Environment Configurations" section.
#
# Dormant: production stays IaC-defined but unprovisioned until alpha
# release (see docs/roadmap.md's note on 005-T108) — do not run
# `terraform apply -var-file=environments/production.tfvars` for real until
# then.

environment        = "production"
vpc_cidr           = "10.1.0.0/16" # non-overlapping with staging's 10.0.0.0/16
availability_zones = ["us-east-1a", "us-east-1b"]
enable_nat_gateway = true

# ECS configuration
task_cpu           = "1024" # 1 vCPU
task_memory        = "2048" # 2 GB RAM
min_tasks          = 2
max_tasks          = 20
log_retention_days = 30

# RDS configuration
instance_class        = "db.t4g.small"
multi_az              = true
backup_retention_days = 30

# Domain configuration
frontend_domain = "traveler.example.com"

# ECR / ACM configuration
#
# Placeholders: the ECR repository and ACM certificates referenced below are
# not provisioned yet (production infra is dormant, per this repo's
# AWS-cost-avoidance policy — see G-SPRINT3-INFRA-CICD / issue #91 and
# infra/README.md's "AWS Account Setup" section). Replace with real values
# before the first real `terraform apply` against this environment.
ecr_repository_url       = "123456789012.dkr.ecr.us-east-1.amazonaws.com/traveler-backend-production:latest"
backend_certificate_arn  = "arn:aws:acm:us-east-1:123456789012:certificate/REPLACE_WITH_PRODUCTION_BACKEND_CERT"
frontend_certificate_arn = "arn:aws:acm:us-east-1:123456789012:certificate/REPLACE_WITH_PRODUCTION_FRONTEND_CERT"
