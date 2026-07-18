# Notification Service — Design Spec

**Date:** 2026-05-17 (addendum 2026-07-18 — reused as the video-processor domain's email consumer, see below)
**Status:** Approved

---

## Context

The `tech-challenge-notification-service` was initially a Node.js Lambda stub. It was refactored to Go using hexagonal architecture. This spec defines the full feature implementation: SNS-driven email notifications with dynamic, HTTP-manageable templates stored in S3.

### Addendum 2026-07-18 — reused by `video-processor` (ADR-012)

This Lambda is intentionally project-agnostic and is now also deployed as the email consumer for the `video-processor` domain (signup email verification), documented in `docs/superpowers/specs/2026-07-18-notification-signup-integration-design.md` (workspace root) and in `video-processor-authentication-api`'s own spec. No Go code changes were needed — `ConsumeSQS` (`internal/adapter/consumer/notification_consumer.go`, added to this codebase since this spec was originally written) already parses any SNS-in-SQS envelope generically, keyed only by `MessageAttributes.templateType`. Two things do change per deployment context:

- **New template**: `email-verification` (S3 `templates/email-verification.html`, variable `$verification_link`), added via `PUT /templates/email-verification` — same mechanism as any other template, no schema change.
- **`sqs_queue_name` variable**: `iac-video-processor-infra` names its queue `notification-events-queue-${var.environment}` (with an environment suffix, per that repo's naming convention) — different from this repo's default (`notification-events-queue`, no suffix, inherited from `tech-challenge-fiap`'s convention). The `video-processor` deployment/pipeline must pass `-var="sqs_queue_name=notification-events-queue-prod"` (or equivalent per environment) explicitly; the Terraform code itself (`data "aws_sqs_queue"` + `aws_lambda_event_source_mapping`) needs no change.

The `notification-events-topic`/`-queue` resources that exist in `iac-tech-challenge-infra` (`tech-challenge-fiap`, no longer in active use) are **not** reused — `video-processor` provisions its own, separate resources with the same name pattern in its own AWS account context.

---

## Goals

- Lambda listens to SNS topic `notification-events` and sends emails via Mailtrap
- Email templates stored in S3, manageable (create/update/retrieve) via HTTP API
- Single Lambda function handles both SNS events and API Gateway HTTP requests
- JWT-authenticated HTTP API for template management
- Hexagonal architecture: domain and use cases are infrastructure-agnostic

---

## Architecture

### Approach

Single Lambda function with a raw JSON dispatcher that detects the event source and routes accordingly:

- `Records[0].EventSource == "aws:sns"` → notification consumer
- `httpMethod != ""` → HTTP template handler

### Directory Structure

```
cmd/lambda/main.go

internal/
├── domain/
│   ├── notification.go            (existing — Notification, Recipient, types)
│   └── template.go                (new — Template, Render)

├── application/
│   ├── ports/
│   │   ├── notification_sender.go (existing)
│   │   └── template_repository.go (new)
│   └── usecases/
│       ├── send_notification/
│       │   ├── usecase.go
│       │   └── usecase_test.go
│       ├── get_template/
│       │   ├── usecase.go
│       │   └── usecase_test.go
│       └── save_template/
│           ├── usecase.go
│           └── usecase_test.go

├── adapter/
│   ├── lambda/
│   │   └── dispatcher.go          (raw JSON dispatch)
│   ├── consumer/
│   │   └── notification_consumer.go
│   ├── http/
│   │   ├── handlers/
│   │   │   └── template_handler.go
│   │   └── middlewares/
│   │       └── auth_middleware.go
│   ├── mailtrap/
│   │   └── email_sender.go
│   ├── s3/
│   │   └── template_repository.go
│   └── logger/
│       └── log_sender.go          (kept for local dev)

└── mocks/
    ├── mock_notification_sender.go (existing)
    └── mock_template_repository.go (new)
```

**Dependency rule:** Domain ← Application (ports + usecases) ← Adapters. No adapter imports another adapter. Wiring lives exclusively in `main.go`.

**Naming rule:** Adapters are named after domain purpose, not infrastructure. No `sns_`, `apigw_` prefixes.

---

## Domain

### `domain/template.go`

```go
type Template struct {
    Type    string
    Subject string
    HTML    string
}

func (t *Template) Render(data map[string]string) string
// Substitutes $Key with value for each entry in data.
// Variables with no matching key in data are left as $Variable literals in the HTML.
// Never returns an error.
```

### `domain/notification.go` (additions)

```go
type Recipient struct {
    Email string
    Name  string
}
```

`Recipient` is added as a value object. The `Notification.Recipient` field changes from `string` to `domain.Recipient` so the Mailtrap adapter can read both `Email` and `Name` from a single `Send(notification)` call.

---

## Ports

### `application/ports/template_repository.go`

```go
type TemplateRepository interface {
    Get(templateType string) (*domain.Template, error)
    Save(template *domain.Template) error
}
```

`Get` returns `domain.ErrTemplateNotFound` (sentinel error) when the S3 object does not exist.

### `application/ports/notification_sender.go` (unchanged)

```go
type NotificationSender interface {
    Send(notification *domain.Notification) error
}
```

---

## Use Cases

### `send_notification`

**Input:** `templateType string`, `recipient domain.Recipient`, `data map[string]string`

**Flow:**
1. `repo.Get(templateType)` — fetch HTML from S3
2. `template.Render(data)` — substitute `$Variable` placeholders
3. Populate `domain.Notification` with rendered HTML, subject, recipient
4. `sender.Send(notification)` — POST to Mailtrap `/send`

**Errors:** propagates `ErrTemplateNotFound` if template missing; propagates sender errors on Mailtrap failure.

### `get_template`

**Input:** `templateType string`  
**Output:** `*domain.Template, error`

Calls `repo.Get(templateType)`. Returns `ErrTemplateNotFound` if not found.

### `save_template`

**Input:** `templateType, subject, html string`  
**Output:** `error`

Validates that `templateType`, `subject`, and `html` are non-empty, then calls `repo.Save`.

---

## Adapters

### `adapter/lambda/dispatcher.go`

Receives `json.RawMessage`. Unmarshals a minimal probe struct to detect event type:

```go
type probe struct {
    Records    []struct{ EventSource string } `json:"Records"`
    HTTPMethod string                          `json:"httpMethod"`
}
```

- `Records[0].EventSource == "aws:sns"` → delegates to `notification_consumer`
- `HTTPMethod != ""` → delegates to HTTP handler chain (auth middleware → template handler)
- Neither → returns explicit error with log

### `adapter/consumer/notification_consumer.go`

Parses `events.SNSEvent`:
- Reads `MessageAttributes["templateType"].Value`
- Unmarshals `Message` body as `{ "recipient": {...}, "data": {...} }`
- Calls `SendNotificationUseCase.Execute`

### `adapter/http/middlewares/auth_middleware.go`

Validates `Authorization: Bearer <token>` header in `events.APIGatewayProxyRequest.Headers`.  
Uses `github.com/golang-jwt/jwt/v5` with `JWT_SECRET` env var.  
Returns `401` JSON response if token is missing or invalid.

### `adapter/http/handlers/template_handler.go`

Routes by `httpMethod` and `pathParameters["type"]`:

| Method | Path | Handler |
|--------|------|---------|
| GET | `/templates/{type}` | `GetTemplateUseCase.Execute` → `200 text/html` |
| PUT | `/templates/{type}` | `SaveTemplateUseCase.Execute` → `200 JSON` |

Returns `404` when template not found, `400` on invalid body, `500` on unexpected error.

### `adapter/mailtrap/email_sender.go`

Implements `NotificationSender`. Mirrors the pattern from `tech-challenge-s1/internal/adapter/email/sendtrap.go`:

- `POST {MAILTRAP_URL}/send`
- Header: `Api-Token: {MAILTRAP_TOKEN}`
- Body: `{ from, to[], subject, text, html }`
- `from` fixed: `{ email: MAILTRAP_FROM_EMAIL, name: MAILTRAP_FROM_NAME }`

### `adapter/s3/template_repository.go`

Implements `TemplateRepository`:

- **Get:** `s3.GetObject(bucket, "templates/{type}.html")` → populates `Template.HTML`. Subject stored as S3 object metadata key `x-amz-meta-subject`.
- **Save:** `s3.PutObject(bucket, "templates/{type}.html", html)` with metadata `x-amz-meta-subject`.
- Returns `domain.ErrTemplateNotFound` on `NoSuchKey` S3 error.

---

## SNS Contract

**Message Attribute:**
```
templateType = "order-approval"   (type: String)
```

**Message Body (JSON):**
```json
{
  "recipient": {
    "email": "cliente@email.com",
    "name": "João"
  },
  "data": {
    "ID": "OS-001",
    "Value": "R$ 350,00",
    "DateIn": "2026-05-17",
    "DiagnosticNote": "Freios desgastados",
    "Approval_url": "https://...",
    "Repprove_url": "https://..."
  }
}
```

Message attributes enable SNS subscription filters by `templateType` in the future.

---

## HTTP API Contract

All endpoints require `Authorization: Bearer <jwt>` header.

### GET `/templates/{type}`

**Response 200:**
```
Content-Type: text/html
<raw HTML content>
```

**Response 404:**
```json
{ "success": false, "error": "template not found" }
```

### PUT `/templates/{type}`

**Request body:**
```json
{ "subject": "Aprovação de Ordem de Serviço", "html": "<!DOCTYPE html>..." }
```

**Response 200:**
```json
{ "success": true, "message": "template saved" }
```

**Response 400:** missing or empty `subject` / `html`  
**Response 401:** invalid or missing JWT

---

## Error Handling

| Situation | Behavior |
|-----------|----------|
| Template not found in S3 | `ErrTemplateNotFound` → HTTP 404 or Lambda error → SNS retry |
| Variable missing in `data` | `$Variable` left as literal in rendered HTML — send proceeds |
| Mailtrap failure | `Send` returns error → Lambda fails → SNS automatic retry (up to 3x) |
| JWT invalid or missing | `auth_middleware` returns 401 before reaching handler |
| Invalid PUT body | Handler returns 400 with descriptive message |
| S3 Get/Put failure | Error propagated → SNS retry or HTTP 500 |
| Unknown event type | `dispatcher` returns explicit error — Lambda fails visibly |

---

## Environment Variables

```
MAILTRAP_TOKEN        Mailtrap API token
MAILTRAP_URL          Mailtrap API base URL (e.g. https://send.api.mailtrap.io/api)
MAILTRAP_FROM_EMAIL   Sender email address
MAILTRAP_FROM_NAME    Sender display name
S3_BUCKET_NAME        S3 bucket name for template storage
JWT_SECRET            Secret for JWT validation
AWS_REGION            AWS region (default: us-east-1)
```

---

## Testing Strategy

**Domain — pure unit tests (no mocks):**
- `template.Render` substitutes present variables
- `template.Render` leaves missing variables as literals
- `template.Render` with empty data returns original HTML

**Use cases — unit tests with mocks:**
- `send_notification`: success, template not found, sender failure
- `get_template`: found, not found
- `save_template`: success, repository failure

**Adapters (consumer, handlers):** not tested directly — they are thin parse-and-delegate layers. Behavior is covered by use case tests.

**Mocks:** `MockNotificationSender` (existing), `MockTemplateRepository` (new).

---

## Wiring (`cmd/lambda/main.go`)

```go
templateRepo := s3adapter.NewTemplateRepository(s3Client, os.Getenv("S3_BUCKET_NAME"))
emailSender  := mailtrap.NewEmailSender(
    os.Getenv("MAILTRAP_TOKEN"),
    os.Getenv("MAILTRAP_URL"),
    os.Getenv("MAILTRAP_FROM_EMAIL"),
    os.Getenv("MAILTRAP_FROM_NAME"),
)

sendNotif    := send_notification.New(templateRepo, emailSender)
getTemplate  := get_template.New(templateRepo)
saveTemplate := save_template.New(templateRepo)

consumer := consumer.New(sendNotif)
handler  := templatehandler.New(getTemplate, saveTemplate)

lambda.Start(dispatcher.New(consumer, handler).Dispatch)
```
