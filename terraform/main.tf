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

resource "aws_lambda_function" "this" {
  function_name = "tech-challenge-notification-service"
  role          = data.aws_iam_role.lab_role.arn
  package_type  = "Image"
  image_uri     = "${data.aws_ecr_repository.this.repository_url}:${var.image_tag}"

  reserved_concurrent_executions = 3

  timeout     = 30
  memory_size = 128

  image_config {
    command = ["index.handler"]
  }
}
