# VPC Networking Module
#
# Creates an isolated VPC with public/private subnets across 2 AZs, an
# Internet Gateway for public egress, and a cost-aware outbound path for
# private subnets: a DIY NAT instance (staging) or a managed NAT Gateway
# (production), selected by var.enable_nat_gateway.
#
# CIDR allocation scheme (non-overlapping /24s carved out of the module's
# /16 vpc_cidr via cidrsubnet(cidr, 8, netnum)):
# - Public subnets:  netnum 0-1  (count.index)
# - Private subnets: netnum 10-11 (count.index + 10, offset to leave room
#   for additional public subnets without colliding with private ranges)

# ---------------------------------------------------------------------------
# VPC
# ---------------------------------------------------------------------------

resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.environment}-vpc"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Subnets
# ---------------------------------------------------------------------------

resource "aws_subnet" "public" {
  count = length(var.availability_zones)

  vpc_id                  = aws_vpc.main.id
  cidr_block              = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone       = var.availability_zones[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name        = "${var.environment}-public-subnet-${count.index + 1}"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_subnet" "private" {
  count = length(var.availability_zones)

  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + 10)
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name        = "${var.environment}-private-subnet-${count.index + 1}"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Internet Gateway + public routing
# ---------------------------------------------------------------------------

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "${var.environment}-igw"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name        = "${var.environment}-public-rt"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_route_table_association" "public" {
  count = length(aws_subnet.public)

  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# ---------------------------------------------------------------------------
# Outbound path for private subnets
#
# Production (enable_nat_gateway = true): managed NAT Gateway, higher
# availability, no instance management.
# Staging (enable_nat_gateway = false): DIY NAT instance on a small ARM64
# Graviton instance, consistent with this repo's Graviton cost-optimization
# convention used elsewhere for ECS (see docs/cloud-and-environments.md).
# ---------------------------------------------------------------------------

# NAT Gateway path (production)

resource "aws_eip" "nat" {
  count = var.enable_nat_gateway ? 1 : 0

  domain = "vpc"

  tags = {
    Name        = "${var.environment}-nat-eip"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_nat_gateway" "main" {
  count = var.enable_nat_gateway ? 1 : 0

  allocation_id = aws_eip.nat[0].id
  subnet_id     = aws_subnet.public[0].id

  depends_on = [aws_internet_gateway.main]

  tags = {
    Name        = "${var.environment}-nat-gateway"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# NAT Instance path (staging)

data "aws_ami" "nat_instance" {
  count = var.enable_nat_gateway ? 0 : 1

  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-arm64"]
  }

  filter {
    name   = "architecture"
    values = ["arm64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_security_group" "nat_instance" {
  count = var.enable_nat_gateway ? 0 : 1

  name        = "${var.environment}-nat-instance-sg"
  description = "Allow inbound traffic from the VPC and all outbound traffic for the NAT instance"
  vpc_id      = aws_vpc.main.id

  ingress {
    description = "All traffic from within the VPC"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "${var.environment}-nat-instance-sg"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_instance" "nat" {
  count = var.enable_nat_gateway ? 0 : 1

  ami                    = data.aws_ami.nat_instance[0].id
  instance_type          = "t4g.nano"
  subnet_id              = aws_subnet.public[0].id
  vpc_security_group_ids = [aws_security_group.nat_instance[0].id]
  source_dest_check      = false

  # DIY NAT bootstrap: enable IPv4 forwarding and masquerade outbound traffic
  # from the private subnets through this instance's primary interface.
  user_data = <<-EOF
    #!/bin/bash
    echo 1 > /proc/sys/net/ipv4/ip_forward
    sysctl -w net.ipv4.ip_forward=1
    echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf
    iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
  EOF

  tags = {
    Name        = "${var.environment}-nat-instance"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

# ---------------------------------------------------------------------------
# Private routing
# ---------------------------------------------------------------------------

locals {
  # Route target for private egress: a NAT Gateway ID (production) or the
  # NAT instance's primary network interface ID (staging), whichever path
  # is active.
  nat_gateway_id            = var.enable_nat_gateway ? aws_nat_gateway.main[0].id : null
  nat_instance_interface_id = var.enable_nat_gateway ? null : aws_instance.nat[0].primary_network_interface_id
}

resource "aws_route_table" "private" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block           = "0.0.0.0/0"
    nat_gateway_id       = local.nat_gateway_id
    network_interface_id = local.nat_instance_interface_id
  }

  tags = {
    Name        = "${var.environment}-private-rt"
    Environment = var.environment
    ManagedBy   = "terraform"
    Project     = "traveler"
  }
}

resource "aws_route_table_association" "private" {
  count = length(aws_subnet.private)

  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private.id
}
