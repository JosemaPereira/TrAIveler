# Root Module Outputs
#
# CI/CD-relevant outputs consumed by backend-ci.yml / frontend-ci.yml
# (Sprint 10) for image push, ECS deployment, and frontend artifact
# sync/cache invalidation. See infra/README.md's "Terraform Outputs for
# CI/CD" section for how these get exported once infra-apply.yml exists.

output "ecr_repository_url" {
  description = "ECR repository URL for the backend container image (echoes the ecr_repository_url input variable, for CI/CD convenience)"
  value       = var.ecr_repository_url
}

output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = module.ecs.service_name
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket holding the frontend build artifacts"
  value       = module.cloudfront.s3_bucket_name
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = module.cloudfront.cloudfront_distribution_id
}
