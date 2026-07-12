# ECS Fargate Cluster & Service Module
#
# Creates the ECS cluster, an ARM64 (Graviton2) Fargate task definition and
# service for the Go backend, CPU-based target-tracking autoscaling, a
# dedicated CloudWatch log group, and a least-privilege security group that
# only accepts traffic from the ALB on port 8080.
#
# runtime_platform.cpu_architecture is pinned to ARM64 per the constitution's
# Infrastructure Layer section and the "ECS Fargate for AI Workloads" pattern
# (see docs/cloud-and-environments.md, .github/memory/patterns-discovered.md)
# — do not switch this to X86_64 without revisiting that documented decision.
#
# Deviations from the originating issue's flat variable list (both required
# for these resources to actually work, not scope creep — see
# variables.tf/outputs.tf for the same note):
#
# 1. alb_security_group_id / target_group_arn variables: T074 requires ECS
#    ingress "from the ALB security group" and T071 requires the service to
#    attach to "the ALB target group," but neither was in the flat variable
#    list, and infra/modules/alb/outputs.tf did not yet export the ALB's
#    security group ID. Both are added here as the minimum necessary input
#    to implement the actual requirement; infra/modules/alb/outputs.tf gained
#    a matching alb_security_group_id output (additive only, ALB module
#    otherwise untouched). Root-module wiring that passes these through is a
#    separate, already-tracked ticket (005-T102 / issue #90).
# 2. aws_iam_role.ecs_task_execution: not itemized in the issue/tasks.md
#    checklist at all, but T070 itself requires "container definition
#    referencing the ECR image URL ... CloudWatch log configuration" — ECS
#    cannot pull a private ECR image or write awslogs without a task
#    execution role. Declared entirely inside this module (no new variables,
#    no other module touched); this is the standard AWS-managed
#    AmazonECSTaskExecutionRolePolicy, nothing custom.

# ---------------------------------------------------------------------------
# Cluster
# ---------------------------------------------------------------------------

resource "aws_ecs_cluster" "main" {
  name = "traveler-${var.environment}-cluster"

  tags = {
    Name        = "traveler-${var.environment}-cluster"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Task execution IAM role
#
# Grants ECS itself (not the application) permission to pull the container
# image from ECR and write logs to CloudWatch on the task's behalf. This is
# the standard AWS-managed execution role — it grants no application-level
# permissions.
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "ecs_task_execution_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_task_execution" {
  name               = "${var.environment}-ecs-task-execution-role"
  assume_role_policy = data.aws_iam_policy_document.ecs_task_execution_assume_role.json

  tags = {
    Name        = "${var.environment}-ecs-task-execution-role"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# ---------------------------------------------------------------------------
# CloudWatch log group
#
# Named to match the log-tail example already documented in infra/README.md's
# troubleshooting section (`aws logs tail /ecs/traveler-staging --follow`).
# ---------------------------------------------------------------------------

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/traveler-${var.environment}"
  retention_in_days = var.log_retention_days

  tags = {
    Name        = "/ecs/traveler-${var.environment}"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Security group
#
# Ingress is restricted to the ALB's security group on port 8080 only — no
# 0.0.0.0/0 ingress is permitted on this SG.
# ---------------------------------------------------------------------------

resource "aws_security_group" "ecs" {
  name        = "${var.environment}-ecs-sg"
  description = "Allow inbound traffic from the ALB only, on the backend's listen port"
  vpc_id      = var.vpc_id

  ingress {
    description     = "Backend traffic from the ALB"
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [var.alb_security_group_id]
  }

  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.environment}-ecs-sg"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Task definition
#
# ARM64/Graviton2, awsvpc networking, FARGATE only — matches this module's
# non-negotiable compute platform choice. Container port 8080 matches the
# ALB target group and the backend's real listen port.
# ---------------------------------------------------------------------------

resource "aws_ecs_task_definition" "main" {
  family                   = "${var.environment}-traveler-backend"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn

  runtime_platform {
    cpu_architecture        = "ARM64"
    operating_system_family = "LINUX"
  }

  container_definitions = jsonencode([
    {
      name      = "backend"
      image     = var.ecr_repository_url
      essential = true

      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]

      environment = [
        {
          name  = "ENVIRONMENT"
          value = var.environment
        },
        {
          name  = "PORT"
          value = "8080"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = data.aws_region.current.name
          "awslogs-stream-prefix" = "backend"
        }
      }
    }
  ])

  tags = {
    Name        = "${var.environment}-traveler-backend"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

data "aws_region" "current" {}

# ---------------------------------------------------------------------------
# Service
# ---------------------------------------------------------------------------

resource "aws_ecs_service" "main" {
  name            = "traveler-${var.environment}-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count   = var.min_tasks
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.private_subnet_ids
    security_groups = [aws_security_group.ecs.id]
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = "backend"
    container_port   = 8080
  }

  health_check_grace_period_seconds = 60

  tags = {
    Name        = "traveler-${var.environment}-service"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Autoscaling: CPU-based target tracking at 70%
# ---------------------------------------------------------------------------

resource "aws_appautoscaling_target" "ecs" {
  min_capacity       = var.min_tasks
  max_capacity       = var.max_tasks
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.main.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "ecs_cpu" {
  name               = "${var.environment}-ecs-cpu-target-tracking"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value = 70
  }
}
