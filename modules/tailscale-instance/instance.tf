// most recent ubuntu ami 
data "aws_ami" "ubuntu" {
  most_recent = true
  filter {
    name = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
  owners = ["099720109477"] # Canonical
}

variable "tailscale_auth_key" {
  type = string
  
}


resource "aws_instance" "grafana" {
  availability_zone = var.aws_availability_zone
  iam_instance_profile = aws_iam_instance_profile.grafana.name
  ami = var.ec2_ami != "" ? var.ec2_ami : data.aws_ami.ubuntu.id
  instance_type = var.ec2_instance
  subnet_id = aws_subnet.grafana.id
  ipv6_address_count = 1
  vpc_security_group_ids = [aws_security_group.grafana.id]
  user_data = <<-EOF
#!/bin/bash
# Install tailscale
curl -fsSL https://tailscale.com/install.sh | sh
sudo tailscale up --authkey ${var.tailscale_auth_key} --hostname ${var.name} --ssh
${var.user_data}
EOF

  lifecycle {
    ignore_changes = [
      ami,
    ]
  }

  tags = {
    application = var.name
    instance = var.name
    Name = var.name
  }
  depends_on = [ 
    aws_iam_instance_profile.grafana,
   ]

}

# public ip address
resource "aws_eip" "grafana" {
  instance = aws_instance.grafana.id
  vpc = true
  tags = var.aws_tags
}
