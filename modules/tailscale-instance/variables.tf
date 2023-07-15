variable "aws_region" {
  description = "AWS region"
  default     = "us-west-1"
}
variable "aws_availability_zone" {
  description = "AWS availability zone"
  default     = "us-west-1a"
}
variable "aws_tags" {
  type = map(string)
  default = {
    application = "grafana"
  }
}

variable "ec2_instance" {
  description = "EC2 instance type"
  default     = "t3.small"
}
variable "ec2_ami" {
  description = "Use a specific AMI ID. Default is to use the latest Debian Stretch."
  default     = ""
}
variable "ebs_type" {
  description = "EBS volume type"
  default     = "io2"
}
variable "ebs_size" {
  description = "EBS volume size (GB)"
  default     = 32
}

variable "user_data" {
  description = "User data to be used for the EC2 instance"
  type       = string
}

variable "name" {
  description = "name of the instance in aws console"
  type       = string
}