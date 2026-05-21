variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "image_tag" {
  description = "ECR image tag"
  type        = string
}

variable "sqs_queue_name" {
  description = "SQS queue name that triggers the notification Lambda"
  type        = string
  default     = "notification-events-queue"
}

variable "s3_bucket_name" {
  description = "S3 bucket name for notification templates"
  type        = string
  default     = "tech-challenge-notification-templates"
}

variable "mailtrap_token" {
  description = "Mailtrap API token"
  type        = string
  sensitive   = true
}

variable "mailtrap_url" {
  description = "Mailtrap API URL"
  type        = string
  default     = "https://send.api.mailtrap.io/api"
}

variable "mailtrap_from_email" {
  description = "Sender email address"
  type        = string
  default     = "contato@nohats.net.br"
}

variable "mailtrap_from_name" {
  description = "Sender display name"
  type        = string
  default     = "Nohats"
}

variable "jwt_secret" {
  description = "JWT secret for API authentication"
  type        = string
  sensitive   = true
}

variable "internal_auth_token" {
  description = "Shared secret token for Service-to-Service (S2S) authentication"
  type        = string
  sensitive   = true
}
