# Application Load Balancer Module
#
# Creates the public-facing Application Load Balancer, its target group
# (forwarding to the ECS service on port 8080), and two listeners:
# - HTTPS (443): terminates TLS using an externally-provisioned ACM
#   certificate and forwards to the target group.
# - HTTP (80): redirects to HTTPS only — it never forwards traffic to the
#   target group.

# ---------------------------------------------------------------------------
# Security group
#
# This is the ONE deliberately public-facing security group in the whole
# stack: ingress is open to 0.0.0.0/0, but only on ports 80 and 443 (the
# ALB never exposes any other port to the internet). Egress is "all
# outbound traffic," matching the RDS module's convention, so the ALB can
# always reach the ECS target group on port 8080 regardless of how the
# ECS security group's own ingress rules are defined.
# ---------------------------------------------------------------------------

resource "aws_security_group" "alb" {
  name        = "${var.environment}-alb-sg"
  description = "Allow inbound HTTP/HTTPS traffic from the internet"
  vpc_id      = var.vpc_id

  ingress {
    description = "HTTP from the internet"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS from the internet"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.environment}-alb-sg"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Application Load Balancer
# ---------------------------------------------------------------------------

resource "aws_lb" "main" {
  name               = "${var.environment}-alb"
  load_balancer_type = "application"
  internal           = false
  subnets            = var.public_subnet_ids
  security_groups    = [aws_security_group.alb.id]

  tags = {
    Name        = "${var.environment}-alb"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Target group
#
# Targets the ECS service by IP on port 8080. Health checks hit the
# backend's real /healthz route (see backend/cmd/api/routes.go) — do not
# change this path independently of that route.
# ---------------------------------------------------------------------------

resource "aws_lb_target_group" "main" {
  name        = "${var.environment}-tg"
  target_type = "ip"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = var.vpc_id

  health_check {
    path                = "/healthz"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
  }

  tags = {
    Name        = "${var.environment}-tg"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Listeners
# ---------------------------------------------------------------------------

# HTTPS listener: terminates TLS and forwards to the target group.
resource "aws_lb_listener" "https" {
  load_balancer_arn = aws_lb.main.arn
  port              = 443
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS13-1-2-2021-06"
  certificate_arn   = var.certificate_arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  tags = {
    Name        = "${var.environment}-alb-listener-https"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# HTTP listener: redirects to HTTPS only. Must never forward to the target
# group directly.
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }

  tags = {
    Name        = "${var.environment}-alb-listener-http"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}
