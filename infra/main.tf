# Root Module Wiring
#
# Instantiates the six infrastructure modules and wires them together via
# module outputs, closing the dependency graph declared by each module's own
# variables.tf. Call order follows the actual data-flow chain:
#
#   vpc -> alb -> ecs -> rds
#
# cloudfront and secrets have no dependency on the VPC-based chain and are
# wired independently. See infra/modules/*/variables.tf and outputs.tf for
# the full input/output contract of each module.

module "vpc" {
  source = "./modules/vpc"

  environment        = var.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
  enable_nat_gateway = var.enable_nat_gateway
}

module "alb" {
  source = "./modules/alb"

  environment       = var.environment
  vpc_id            = module.vpc.vpc_id
  public_subnet_ids = module.vpc.public_subnet_ids
  certificate_arn   = var.backend_certificate_arn
}

module "ecs" {
  source = "./modules/ecs"

  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  task_cpu           = var.task_cpu
  task_memory        = var.task_memory
  min_tasks          = var.min_tasks
  max_tasks          = var.max_tasks
  log_retention_days = var.log_retention_days
  ecr_repository_url = var.ecr_repository_url

  # From the alb module — closes the ALB -> ECS half of the dependency
  # chain (ECS's security group only accepts traffic from the ALB's; the
  # ECS service registers into the ALB's target group).
  alb_security_group_id = module.alb.alb_security_group_id
  target_group_arn      = module.alb.target_group_arn
}

module "rds" {
  source = "./modules/rds"

  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids

  # From the ecs module — RDS only accepts traffic from the ECS service's
  # security group on port 5432.
  ecs_security_group_id = module.ecs.ecs_security_group_id

  instance_class        = var.instance_class
  multi_az              = var.multi_az
  backup_retention_days = var.backup_retention_days
}

module "cloudfront" {
  source = "./modules/cloudfront"

  environment     = var.environment
  domain_name     = var.frontend_domain
  certificate_arn = var.frontend_certificate_arn
}

module "secrets" {
  source = "./modules/secrets"

  environment = var.environment
}
