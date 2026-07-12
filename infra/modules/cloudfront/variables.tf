# CloudFront Module Variables
#
# Input variables for the CloudFront & S3 frontend delivery module:
# environment name, the CDN's public domain alias, the ACM certificate used
# for HTTPS, and the CloudFront price class.

variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}

variable "domain_name" {
  description = "Public domain name (CNAME alias) served by the CloudFront distribution"
  type        = string
}

variable "certificate_arn" {
  description = "ACM certificate ARN used for HTTPS on the CloudFront distribution (must be issued in us-east-1, provisioned outside this module)"
  type        = string
}

variable "price_class" {
  description = "CloudFront price class controlling which edge locations serve traffic"
  type        = string
  default     = "PriceClass_100" # staging: PriceClass_100, production: PriceClass_200
}
