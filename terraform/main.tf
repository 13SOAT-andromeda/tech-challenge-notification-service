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

data "aws_sqs_queue" "notification_events" {
  name = var.sqs_queue_name
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
    command = ["bootstrap"]
  }

  environment {
    variables = {
      S3_BUCKET_NAME      = var.s3_bucket_name
      MAILTRAP_TOKEN      = var.mailtrap_token
      MAILTRAP_URL        = var.mailtrap_url
      MAILTRAP_FROM_EMAIL = var.mailtrap_from_email
      MAILTRAP_FROM_NAME  = var.mailtrap_from_name
      JWT_SECRET          = var.jwt_secret
      INTERNAL_AUTH_TOKEN = var.internal_auth_token
      ADMIN_EMAIL         = var.admin_email
      ADMIN_DOCUMENT      = var.admin_document
    }
  }
}

resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = data.aws_sqs_queue.notification_events.arn
  function_name    = aws_lambda_function.this.arn
  batch_size       = 10
}
