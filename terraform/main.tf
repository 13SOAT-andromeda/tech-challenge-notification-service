terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    key     = "lambda-notification-service.tfstate"
    region  = "us-east-1"
    encrypt = true
  }
}

provider "aws" {
  region = var.aws_region
}

data "aws_ecr_repository" "this" {
  name = "tech-challenge-notification-service-repo"
}

data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

module "lambda" {
  source = "github.com/13SOAT-andromeda/iac-tech-challenge-infra//modules/lambda?ref=main"

  function_name                  = "tech-challenge-notification-service"
  image_uri                      = "${data.aws_ecr_repository.this.repository_url}:latest"
  role_arn                       = data.aws_iam_role.lab_role.arn
  reserved_concurrent_executions = 3
}
