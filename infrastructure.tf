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
