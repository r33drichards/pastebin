terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.20.0"
    }
  }

  required_version = "~> 1.0"
  backend "http" {

  }
}

variable "aws_region" {
  description = "region to create aws app runner service & ecr repo"
  default = "us-east-1"
}
variable "image_tag"  {
  description = "image tag to be deployed to aws app runner"
  default = "latest"
}

variable "pbin_aws_access_key_id" {}

variable "pbin_aws_secret_access_key" {}

variable "pbin_table_name" {
  default = "pbin_prod"
}

variable "pbin_url" {
  default = "https://pbin.jjk.is"
}

variable "tailscale_auth_key" {}

variable "openapikey" {
  type = string
}

provider "aws" {
  region = var.aws_region
}


resource "aws_ecr_repository" "pbin" {
  name                 = "pbin"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_apprunner_service" "pbin" {
  service_name = "pbin"

  source_configuration {
    authentication_configuration {
	  access_role_arn = "arn:aws:iam::150301572911:role/service-role/AppRunnerECRAccessRole"
	}
    image_repository {
      image_configuration {
        port = "8000"
        runtime_environment_variables = {
          AWS_ACCESS_KEY_ID = var.pbin_aws_access_key_id
          AWS_SECRET_ACCESS_KEY = var.pbin_aws_secret_access_key
          AWS_REGION = var.aws_region
          PBIN_TABLE_NAME = var.pbin_table_name
          PBIN_URL = var.pbin_url
        }
      }
	  image_identifier      = "${aws_ecr_repository.pbin.repository_url}:${var.image_tag}"
      image_repository_type = "ECR"
    }
  }
}


# module "tailscale_instance" {
#   source = "./modules/tailscale-instance"
#   aws_region = var.aws_region
#   aws_availability_zone = "us-east-1a"
#   user_data = <<EOF
# # install docker 
# sudo apt-get update
# sudo apt install docker.io -y
# sudo systemctl start docker
# sudo systemctl enable docker

# sudo apt  install awscli -y

# # login to ecr
# eval $(aws ecr get-login --no-include-email --region us-east-1)

# # docker run --env-file .env -p 8000:8000 pbin:latest 
# sudo docker run -d --restart=always \
#   -e AWS_ACCESS_KEY_ID=${var.pbin_aws_access_key_id} \
#   -e AWS_SECRET_ACCESS_KEY=${var.pbin_aws_secret_access_key} \
#   -e AWS_REGION=${var.aws_region} \
#   -e PBIN_TABLE_NAME=${var.pbin_table_name} \
#   -e PBIN_URL=tsbin:8000 \
#   -e OPENAPIKEY=${var.openapikey} \
#   -p 8000:8000 ${aws_ecr_repository.pbin.repository_url}:${var.image_tag}

# EOF
# # admin access 
#   iam_policy = <<EOF
# {
#     "Version": "2012-10-17",
#     "Statement": [
#         {
#             "Sid": "VisualEditor1",
#             "Effect": "Allow",
#             "Action": "*",
#             "Resource": "*"
#         }
#     ]

# }
# EOF
#   name = "tsbin"
#   tailscale_auth_key = var.tailscale_auth_key

# }