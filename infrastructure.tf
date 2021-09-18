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
      }
	  image_identifier      = "${aws_ecr_repository.pbin.repository_url}:${var.image_tag}"
      image_repository_type = "ECR"
    }
  }
}
