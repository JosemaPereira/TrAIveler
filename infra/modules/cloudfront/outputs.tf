# CloudFront Module Outputs
#
# Exposes the S3 bucket name (for frontend build sync) and the CloudFront
# distribution's ID and domain name (for cache invalidation and DNS/root
# module wiring), consumed by CI/CD and downstream root module wiring.

output "s3_bucket_name" {
  description = "Name of the S3 bucket holding the frontend build artifacts"
  value       = aws_s3_bucket.frontend.bucket
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.id
}

output "cloudfront_domain_name" {
  description = "Domain name of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.domain_name
}
