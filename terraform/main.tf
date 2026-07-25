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

resource "aws_ecr_repository" "this" {
  name                 = "tech-challenge-notification-service-repo"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "this" {
  repository = aws_ecr_repository.this.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# Templates are seeded post-apply via `aws s3 cp` (see templates/) — the
# service reads them at runtime from templates/{templateType}.html; there's
# no HTTP route wired to PUT /templates in this Academy Lab deploy.
resource "aws_s3_bucket" "templates" {
  bucket        = var.s3_bucket_name
  force_destroy = true
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
  image_uri     = "${aws_ecr_repository.this.repository_url}:${var.image_tag}"

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
