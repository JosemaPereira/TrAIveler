# ALB Module Outputs
#
# Exposes the load balancer's DNS name and hosted zone ID (used for DNS/
# CloudFront wiring) and the target group ARN, consumed by the ECS module
# and downstream root module wiring.

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = aws_lb.main.dns_name
}

output "alb_zone_id" {
  description = "Hosted zone ID of the Application Load Balancer"
  value       = aws_lb.main.zone_id
}

output "target_group_arn" {
  description = "ARN of the ALB target group"
  value       = aws_lb_target_group.main.arn
}
