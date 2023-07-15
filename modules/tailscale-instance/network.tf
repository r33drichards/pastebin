resource "aws_vpc" "grafana" {
  cidr_block = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true
  tags = var.aws_tags
}

resource "aws_subnet" "grafana" {
  availability_zone = var.aws_availability_zone
  vpc_id = aws_vpc.grafana.id
  cidr_block = cidrsubnet(aws_vpc.grafana.cidr_block, 8, 0)
  ipv6_cidr_block = cidrsubnet(aws_vpc.grafana.ipv6_cidr_block, 8, 0)
  assign_ipv6_address_on_creation = true
  tags = var.aws_tags
}

resource "aws_internet_gateway" "grafana" {
  vpc_id = aws_vpc.grafana.id
  tags = var.aws_tags
}

resource "aws_route_table" "grafana" {
  vpc_id = aws_vpc.grafana.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.grafana.id
  }
  route {
    ipv6_cidr_block = "::/0"
    gateway_id = aws_internet_gateway.grafana.id
  }
  tags = var.aws_tags
}

resource "aws_route_table_association" "grafana" {
  subnet_id = aws_subnet.grafana.id
  route_table_id = aws_route_table.grafana.id
}

resource "aws_security_group" "grafana" {
  name = var.name
  vpc_id = aws_vpc.grafana.id 


  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  
  tags = var.aws_tags
}
