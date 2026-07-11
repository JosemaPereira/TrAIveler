# VPC Module Variables
#
# Input variables for the VPC networking module: environment name, VPC CIDR,
# availability zones, and the NAT strategy toggle (instance vs. gateway).

variable "environment" {
  description = "Environment name (staging or production)"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16" # staging: 10.0.0.0/16, production: 10.1.0.0/16
}

variable "availability_zones" {
  description = "List of availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

variable "enable_nat_gateway" {
  description = "Use NAT Gateway (production) vs NAT Instance (staging)"
  type        = bool
  default     = false # staging: false, production: true
}
