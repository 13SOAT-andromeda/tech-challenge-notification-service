# Notification Service Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement SNS-driven email notification delivery with S3-backed, JWT-protected HTTP template management in a single Go Lambda function.

**Architecture:** Single Lambda using a raw-JSON dispatcher that routes SNS events to a notification consumer and API Gateway events to an HTTP template handler. Domain and use cases are infrastructure-agnostic; all AWS details live in adapters. Three use cases (send_notification, get_template, save_template) replace the previous service layer.

**Tech Stack:** Go 1.22, aws-lambda-go v1.47, aws-sdk-go-v2 (S3), golang-jwt/jwt v5, Mailtrap HTTP API, testify mocks.

**Spec:** `docs/superpowers/specs/2026-05-17-notification-service-design.md`

---

## File Map

| Action | Path |
|--------|------|
| Delete | `internal/adapter/lambda/handler.go` |
| Delete | `internal/application/services/notification_service.go` |
| Delete | `internal/application/services/notification_service_test.go` |
| Modify | `internal/domain/notification.go` — add `Recipient` struct, change field type |
| Modify | `internal/adapter/logger/log_sender.go` — use `Recipient.Email` |
| Modify | `cmd/lambda/main.go` — full rewrite with new wiring |
| Modify | `.env.example` — add new env vars |
| Create | `internal/domain/template.go` |
| Create | `internal/domain/template_test.go` |
| Create | `internal/application/ports/template_repository.go` |
| Create | `internal/mocks/mock_template_repository.go` |
| Create | `internal/application/usecases/send_notification/usecase.go` |
| Create | `internal/application/usecases/send_notification/usecase_test.go` |
| Create | `internal/application/usecases/get_template/usecase.go` |
| Create | `internal/application/usecases/get_template/usecase_test.go` |
| Create | `internal/application/usecases/save_template/usecase.go` |
| Create | `internal/application/usecases/save_template/usecase_test.go` |
| Create | `internal/adapter/mailtrap/email_sender.go` |
| Create | `internal/adapter/s3/template_repository.go` |
| Create | `internal/adapter/http/middlewares/auth_middleware.go` |
| Create | `internal/adapter/http/handlers/template_handler.go` |
| Create | `internal/adapter/consumer/notification_consumer.go` |
| Create | `internal/adapter/lambda/dispatcher.go` |

---

### Task 1: Remove old files

**Files:**
- Delete: `internal/adapter/lambda/handler.go`
- Delete: `internal/application/services/notification_service.go`
- Delete: `internal/application/services/notification_service_test.go`

- [ ] **Step 1: Remove the three stale files**

```bash
rm internal/adapter/lambda/handler.go
rm internal/application/services/notification_service.go
rm internal/application/services/notification_service_test.go
```

- [ ] **Step 2: Verify the project no longer references the deleted service**

```bash
grep -r "services.NotificationService\|services.NewNotificationService" . --include="*.go"
```

Expected: no output.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: remove old service layer and SQS handler"
```

---

### Task 2: Install new dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Add AWS SDK v2 (S3 + config) and JWT library**

```bash
go get github.com/aws/aws-sdk-go-v2/config@latest
go get github.com/aws/aws-sdk-go-v2/service/s3@latest
go get github.com/golang-jwt/jwt/v5@latest
go mod tidy
```

- [ ] **Step 2: Verify build still compiles (only main.go is currently broken from the deleted service)**

```bash
go build ./internal/... 2>&1
```

Expected: errors only from `cmd/lambda/main.go` referencing the deleted service. The `internal/` packages should be clean.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add aws-sdk-go-v2 s3, config and golang-jwt dependencies"
```

---

### Task 3: Domain — Template entity (TDD)

**Files:**
- Create: `internal/domain/template.go`
- Create: `internal/domain/template_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/domain/template_test.go`:

```go
package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

func TestTemplate_Render_SubstitutesAllVariables(t *testing.T) {
	tmpl := &domain.Template{
		Type:    "order-approval",
		Subject: "Aprovação de OS",
		HTML:    "<p>Olá $Name, seu pedido $ID vale $Value</p>",
	}
	result := tmpl.Render(map[string]string{
		"Name":  "João",
		"ID":    "OS-001",
		"Value": "R$ 350,00",
	})
	assert.Equal(t, "<p>Olá João, seu pedido OS-001 vale R$ 350,00</p>", result)
}

func TestTemplate_Render_LeavesMissingVariablesAsLiteral(t *testing.T) {
	tmpl := &domain.Template{
		HTML: "<p>Olá $Name, código: $Code</p>",
	}
	result := tmpl.Render(map[string]string{"Name": "João"})
	assert.Equal(t, "<p>Olá João, código: $Code</p>", result)
}

func TestTemplate_Render_EmptyDataReturnsOriginalHTML(t *testing.T) {
	tmpl := &domain.Template{
		HTML: "<p>Olá $Name</p>",
	}
	result := tmpl.Render(map[string]string{})
	assert.Equal(t, "<p>Olá $Name</p>", result)
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/domain/... -v -run TestTemplate
```

Expected: compilation error — `domain.Template` undefined.

- [ ] **Step 3: Implement `internal/domain/template.go`**

```go
package domain

import (
	"errors"
	"strings"
)

var ErrTemplateNotFound = errors.New("template not found")

type Template struct {
	Type    string
	Subject string
	HTML    string
}

func (t *Template) Render(data map[string]string) string {
	result := t.HTML
	for key, value := range data {
		result = strings.ReplaceAll(result, "$"+key, value)
	}
	return result
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/domain/... -v -run TestTemplate
```

Expected: 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/domain/template.go internal/domain/template_test.go
git commit -m "feat: add Template domain entity with Render method"
```

---

### Task 4: Domain — update Notification + fix logger

**Files:**
- Modify: `internal/domain/notification.go`
- Modify: `internal/adapter/logger/log_sender.go`

- [ ] **Step 1: Update `internal/domain/notification.go`**

Replace the entire file content:

```go
package domain

import (
	"errors"
	"time"
)

type NotificationType string

const (
	NotificationTypeEmail NotificationType = "email"
	NotificationTypeSMS   NotificationType = "sms"
	NotificationTypePush  NotificationType = "push"
)

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

type Recipient struct {
	Email string
	Name  string
}

type Notification struct {
	ID        string
	Type      NotificationType
	Recipient Recipient
	Subject   string
	Body      string
	Status    NotificationStatus
	CreatedAt time.Time
}

func NewNotification(notifType NotificationType, recipient Recipient, subject, body string) (*Notification, error) {
	if recipient.Email == "" {
		return nil, errors.New("recipient email is required")
	}
	if body == "" {
		return nil, errors.New("body is required")
	}
	if notifType != NotificationTypeEmail && notifType != NotificationTypeSMS && notifType != NotificationTypePush {
		return nil, errors.New("invalid notification type")
	}
	return &Notification{
		Type:      notifType,
		Recipient: recipient,
		Subject:   subject,
		Body:      body,
		Status:    NotificationStatusPending,
		CreatedAt: time.Now(),
	}, nil
}
```

- [ ] **Step 2: Update `internal/adapter/logger/log_sender.go`** to use `Recipient.Email`

```go
package logger

import (
	"log"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) Send(notification *domain.Notification) error {
	log.Printf("[NOTIFICATION] type=%s recipient=%s subject=%q",
		notification.Type, notification.Recipient.Email, notification.Subject)
	return nil
}
```

- [ ] **Step 3: Confirm internal packages compile**

```bash
go build ./internal/...
```

Expected: no errors (only `cmd/lambda/main.go` still broken from old wiring — ignore for now).

- [ ] **Step 4: Commit**

```bash
git add internal/domain/notification.go internal/adapter/logger/log_sender.go
git commit -m "feat: add Recipient value object and update Notification domain"
```

---

### Task 5: Port — TemplateRepository interface

**Files:**
- Create: `internal/application/ports/template_repository.go`

- [ ] **Step 1: Create `internal/application/ports/template_repository.go`**

```go
package ports

import "github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"

type TemplateRepository interface {
	Get(templateType string) (*domain.Template, error)
	Save(template *domain.Template) error
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/application/ports/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/application/ports/template_repository.go
git commit -m "feat: add TemplateRepository port interface"
```

---

### Task 6: Mock — MockTemplateRepository

**Files:**
- Create: `internal/mocks/mock_template_repository.go`

- [ ] **Step 1: Create `internal/mocks/mock_template_repository.go`**

```go
package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type MockTemplateRepository struct {
	mock.Mock
}

func (m *MockTemplateRepository) Get(templateType string) (*domain.Template, error) {
	args := m.Called(templateType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Template), args.Error(1)
}

func (m *MockTemplateRepository) Save(template *domain.Template) error {
	args := m.Called(template)
	return args.Error(0)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/mocks/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/mocks/mock_template_repository.go
git commit -m "feat: add MockTemplateRepository testify mock"
```

---

### Task 7: Use Case — send_notification (TDD)

**Files:**
- Create: `internal/application/usecases/send_notification/usecase.go`
- Create: `internal/application/usecases/send_notification/usecase_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/application/usecases/send_notification/usecase_test.go`:

```go
package send_notification_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/send_notification"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/mocks"
)

func TestSendNotification_Execute_Success(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	mockSender := new(mocks.MockNotificationSender)
	uc := send_notification.New(mockRepo, mockSender)

	tmpl := &domain.Template{
		Type:    "order-approval",
		Subject: "Aprovação de OS",
		HTML:    "<p>Olá $Name</p>",
	}
	recipient := domain.Recipient{Email: "cliente@email.com", Name: "João"}

	mockRepo.On("Get", "order-approval").Return(tmpl, nil)
	mockSender.On("Send", mock.AnythingOfType("*domain.Notification")).Return(nil)

	err := uc.Execute("order-approval", recipient, map[string]string{"Name": "João"})

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestSendNotification_Execute_TemplateNotFound(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	mockSender := new(mocks.MockNotificationSender)
	uc := send_notification.New(mockRepo, mockSender)

	mockRepo.On("Get", "unknown").Return(nil, domain.ErrTemplateNotFound)

	err := uc.Execute("unknown", domain.Recipient{Email: "a@b.com", Name: "A"}, nil)

	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
	mockSender.AssertNotCalled(t, "Send", mock.Anything)
}

func TestSendNotification_Execute_SenderFailure(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	mockSender := new(mocks.MockNotificationSender)
	uc := send_notification.New(mockRepo, mockSender)

	tmpl := &domain.Template{Type: "order-approval", Subject: "Subj", HTML: "<p>Hi</p>"}
	mockRepo.On("Get", "order-approval").Return(tmpl, nil)
	mockSender.On("Send", mock.AnythingOfType("*domain.Notification")).Return(errors.New("mailtrap unavailable"))

	err := uc.Execute("order-approval", domain.Recipient{Email: "a@b.com", Name: "A"}, nil)

	assert.EqualError(t, err, "mailtrap unavailable")
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/application/usecases/send_notification/... -v
```

Expected: compilation error — package `send_notification` not found.

- [ ] **Step 3: Implement `internal/application/usecases/send_notification/usecase.go`**

```go
package send_notification

import (
	"time"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/ports"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type UseCase struct {
	repo   ports.TemplateRepository
	sender ports.NotificationSender
}

func New(repo ports.TemplateRepository, sender ports.NotificationSender) *UseCase {
	return &UseCase{repo: repo, sender: sender}
}

func (uc *UseCase) Execute(templateType string, recipient domain.Recipient, data map[string]string) error {
	tmpl, err := uc.repo.Get(templateType)
	if err != nil {
		return err
	}

	renderedHTML := tmpl.Render(data)

	notification := &domain.Notification{
		Type:      domain.NotificationTypeEmail,
		Recipient: recipient,
		Subject:   tmpl.Subject,
		Body:      renderedHTML,
		Status:    domain.NotificationStatusPending,
		CreatedAt: time.Now(),
	}

	return uc.sender.Send(notification)
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/application/usecases/send_notification/... -v
```

Expected: 3 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/application/usecases/send_notification/
git commit -m "feat: add send_notification use case"
```

---

### Task 8: Use Case — get_template (TDD)

**Files:**
- Create: `internal/application/usecases/get_template/usecase.go`
- Create: `internal/application/usecases/get_template/usecase_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/application/usecases/get_template/usecase_test.go`:

```go
package get_template_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/get_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/mocks"
)

func TestGetTemplate_Execute_Found(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	uc := get_template.New(mockRepo)

	expected := &domain.Template{
		Type:    "order-approval",
		Subject: "Aprovação de OS",
		HTML:    "<p>HTML content</p>",
	}
	mockRepo.On("Get", "order-approval").Return(expected, nil)

	result, err := uc.Execute("order-approval")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	mockRepo.AssertExpectations(t)
}

func TestGetTemplate_Execute_NotFound(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	uc := get_template.New(mockRepo)

	mockRepo.On("Get", "unknown").Return(nil, domain.ErrTemplateNotFound)

	result, err := uc.Execute("unknown")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/application/usecases/get_template/... -v
```

Expected: compilation error — package `get_template` not found.

- [ ] **Step 3: Implement `internal/application/usecases/get_template/usecase.go`**

```go
package get_template

import (
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/ports"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type UseCase struct {
	repo ports.TemplateRepository
}

func New(repo ports.TemplateRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(templateType string) (*domain.Template, error) {
	return uc.repo.Get(templateType)
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/application/usecases/get_template/... -v
```

Expected: 2 tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/application/usecases/get_template/
git commit -m "feat: add get_template use case"
```

---

### Task 9: Use Case — save_template (TDD)

**Files:**
- Create: `internal/application/usecases/save_template/usecase.go`
- Create: `internal/application/usecases/save_template/usecase_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/application/usecases/save_template/usecase_test.go`:

```go
package save_template_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/save_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/mocks"
)

func TestSaveTemplate_Execute_Success(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	uc := save_template.New(mockRepo)

	mockRepo.On("Save", mock.AnythingOfType("*domain.Template")).Return(nil)

	err := uc.Execute("order-approval", "Aprovação de OS", "<p>HTML</p>")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSaveTemplate_Execute_MissingType(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	uc := save_template.New(mockRepo)

	err := uc.Execute("", "Subject", "<p>HTML</p>")

	assert.EqualError(t, err, "template type is required")
	mockRepo.AssertNotCalled(t, "Save", mock.Anything)
}

func TestSaveTemplate_Execute_MissingSubject(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	uc := save_template.New(mockRepo)

	err := uc.Execute("order-approval", "", "<p>HTML</p>")

	assert.EqualError(t, err, "subject is required")
	mockRepo.AssertNotCalled(t, "Save", mock.Anything)
}

func TestSaveTemplate_Execute_MissingHTML(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	uc := save_template.New(mockRepo)

	err := uc.Execute("order-approval", "Subject", "")

	assert.EqualError(t, err, "html is required")
	mockRepo.AssertNotCalled(t, "Save", mock.Anything)
}

func TestSaveTemplate_Execute_RepositoryFailure(t *testing.T) {
	mockRepo := new(mocks.MockTemplateRepository)
	uc := save_template.New(mockRepo)

	mockRepo.On("Save", mock.AnythingOfType("*domain.Template")).Return(errors.New("s3 error"))

	err := uc.Execute("order-approval", "Subject", "<p>HTML</p>")

	assert.EqualError(t, err, "s3 error")
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./internal/application/usecases/save_template/... -v
```

Expected: compilation error — package `save_template` not found.

- [ ] **Step 3: Implement `internal/application/usecases/save_template/usecase.go`**

```go
package save_template

import (
	"errors"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/ports"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type UseCase struct {
	repo ports.TemplateRepository
}

func New(repo ports.TemplateRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(templateType, subject, html string) error {
	if templateType == "" {
		return errors.New("template type is required")
	}
	if subject == "" {
		return errors.New("subject is required")
	}
	if html == "" {
		return errors.New("html is required")
	}
	return uc.repo.Save(&domain.Template{
		Type:    templateType,
		Subject: subject,
		HTML:    html,
	})
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./internal/application/usecases/save_template/... -v
```

Expected: 5 tests PASS.

- [ ] **Step 5: Run the full test suite**

```bash
go test ./...
```

Expected: all tests in `domain`, `send_notification`, `get_template`, `save_template` pass. `cmd/lambda` has no test files.

- [ ] **Step 6: Commit**

```bash
git add internal/application/usecases/save_template/
git commit -m "feat: add save_template use case"
```

---

### Task 10: Adapter — Mailtrap email_sender

**Files:**
- Create: `internal/adapter/mailtrap/email_sender.go`

- [ ] **Step 1: Create `internal/adapter/mailtrap/email_sender.go`**

```go
package mailtrap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type EmailSender struct {
	apiKey    string
	apiUrl    string
	fromEmail string
	fromName  string
}

func NewEmailSender(apiKey, apiUrl, fromEmail, fromName string) *EmailSender {
	return &EmailSender{
		apiKey:    apiKey,
		apiUrl:    apiUrl,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

type contact struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type sendPayload struct {
	From    contact   `json:"from"`
	To      []contact `json:"to"`
	Subject string    `json:"subject"`
	Text    string    `json:"text"`
	Html    string    `json:"html"`
}

func (s *EmailSender) Send(notification *domain.Notification) error {
	p := sendPayload{
		From:    contact{Email: s.fromEmail, Name: s.fromName},
		To:      []contact{{Email: notification.Recipient.Email, Name: notification.Recipient.Name}},
		Subject: notification.Subject,
		Text:    notification.Subject,
		Html:    notification.Body,
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/send", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Api-Token", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailtrap request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailtrap error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
```

- [ ] **Step 2: Verify it compiles and implements the port**

```bash
go build ./internal/adapter/mailtrap/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/mailtrap/
git commit -m "feat: add Mailtrap email sender adapter"
```

---

### Task 11: Adapter — S3 template_repository

**Files:**
- Create: `internal/adapter/s3/template_repository.go`

- [ ] **Step 1: Create `internal/adapter/s3/template_repository.go`**

```go
package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type TemplateRepository struct {
	client *s3.Client
	bucket string
}

func NewTemplateRepository(client *s3.Client, bucket string) *TemplateRepository {
	return &TemplateRepository{client: client, bucket: bucket}
}

func (r *TemplateRepository) objectKey(templateType string) string {
	return "templates/" + templateType + ".html"
}

func (r *TemplateRepository) Get(templateType string) (*domain.Template, error) {
	result, err := r.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(r.objectKey(templateType)),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get template from S3: %w", err)
	}
	defer result.Body.Close()

	htmlBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read template body: %w", err)
	}

	subject := ""
	if result.Metadata != nil {
		subject = result.Metadata["subject"]
	}

	return &domain.Template{
		Type:    templateType,
		Subject: subject,
		HTML:    string(htmlBytes),
	}, nil
}

func (r *TemplateRepository) Save(template *domain.Template) error {
	_, err := r.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(r.objectKey(template.Type)),
		Body:        bytes.NewReader([]byte(template.HTML)),
		ContentType: aws.String("text/html"),
		Metadata:    map[string]string{"subject": template.Subject},
	})
	if err != nil {
		return fmt.Errorf("failed to save template to S3: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/adapter/s3/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/s3/
git commit -m "feat: add S3 template repository adapter"
```

---

### Task 12: Adapter — HTTP auth middleware

**Files:**
- Create: `internal/adapter/http/middlewares/auth_middleware.go`

- [ ] **Step 1: Create `internal/adapter/http/middlewares/auth_middleware.go`**

```go
package middlewares

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	header := request.Headers["Authorization"]
	if header == "" {
		header = request.Headers["authorization"]
	}

	if !strings.HasPrefix(header, "Bearer ") {
		return unauthorizedResponse("missing token")
	}

	tokenStr := strings.TrimPrefix(header, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return unauthorizedResponse("invalid token")
	}

	return nil
}

func unauthorizedResponse(message string) *events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{"success": false, "error": message})
	return &events.APIGatewayProxyResponse{
		StatusCode: 401,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/adapter/http/middlewares/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/http/middlewares/auth_middleware.go
git commit -m "feat: add JWT auth middleware for API Gateway requests"
```

---

### Task 13: Adapter — HTTP template handler

**Files:**
- Create: `internal/adapter/http/handlers/template_handler.go`

- [ ] **Step 1: Create `internal/adapter/http/handlers/template_handler.go`**

```go
package handlers

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/http/middlewares"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/get_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/save_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type TemplateHandler struct {
	getTemplate  *get_template.UseCase
	saveTemplate *save_template.UseCase
}

func NewTemplateHandler(getTemplate *get_template.UseCase, saveTemplate *save_template.UseCase) *TemplateHandler {
	return &TemplateHandler{
		getTemplate:  getTemplate,
		saveTemplate: saveTemplate,
	}
}

type saveTemplateRequest struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

func (h *TemplateHandler) Handle(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	if resp := middlewares.ValidateJWT(request); resp != nil {
		return *resp
	}

	templateType := request.PathParameters["type"]

	switch request.HTTPMethod {
	case "GET":
		return h.handleGet(templateType)
	case "PUT":
		return h.handlePut(templateType, request.Body)
	default:
		return jsonResponse(405, map[string]interface{}{"success": false, "error": "method not allowed"})
	}
}

func (h *TemplateHandler) handleGet(templateType string) events.APIGatewayProxyResponse {
	tmpl, err := h.getTemplate.Execute(templateType)
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return jsonResponse(404, map[string]interface{}{"success": false, "error": "template not found"})
		}
		return jsonResponse(500, map[string]interface{}{"success": false, "error": "internal server error"})
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "text/html"},
		Body:       tmpl.HTML,
	}
}

func (h *TemplateHandler) handlePut(templateType, body string) events.APIGatewayProxyResponse {
	var req saveTemplateRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, map[string]interface{}{"success": false, "error": "invalid request body"})
	}

	if err := h.saveTemplate.Execute(templateType, req.Subject, req.HTML); err != nil {
		return jsonResponse(400, map[string]interface{}{"success": false, "error": err.Error()})
	}

	return jsonResponse(200, map[string]interface{}{"success": true, "message": "template saved"})
}

func jsonResponse(statusCode int, body interface{}) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/adapter/http/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/http/handlers/template_handler.go
git commit -m "feat: add HTTP template handler for GET and PUT routes"
```

---

### Task 14: Adapter — notification consumer

**Files:**
- Create: `internal/adapter/consumer/notification_consumer.go`

- [ ] **Step 1: Create `internal/adapter/consumer/notification_consumer.go`**

```go
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/send_notification"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type NotificationConsumer struct {
	useCase *send_notification.UseCase
}

func New(useCase *send_notification.UseCase) *NotificationConsumer {
	return &NotificationConsumer{useCase: useCase}
}

type notificationBody struct {
	Recipient struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"recipient"`
	Data map[string]string `json:"data"`
}

func (c *NotificationConsumer) Consume(ctx context.Context, event events.SNSEvent) error {
	for _, record := range event.Records {
		log.Printf("processing SNS message: %s", record.SNS.MessageID)

		templateTypeAttr, ok := record.SNS.MessageAttributes["templateType"]
		if !ok {
			return fmt.Errorf("missing templateType message attribute in record %s", record.SNS.MessageID)
		}

		templateType := templateTypeAttr.Value

		var body notificationBody
		if err := json.Unmarshal([]byte(record.SNS.Message), &body); err != nil {
			return fmt.Errorf("failed to parse SNS message body in record %s: %w", record.SNS.MessageID, err)
		}

		recipient := domain.Recipient{
			Email: body.Recipient.Email,
			Name:  body.Recipient.Name,
		}

		if err := c.useCase.Execute(templateType, recipient, body.Data); err != nil {
			return fmt.Errorf("failed to process notification in record %s: %w", record.SNS.MessageID, err)
		}

		log.Printf("notification processed successfully: %s", record.SNS.MessageID)
	}
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/adapter/consumer/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/consumer/
git commit -m "feat: add notification consumer adapter for SNS events"
```

---

### Task 15: Adapter — Lambda dispatcher

**Files:**
- Create: `internal/adapter/lambda/dispatcher.go`

- [ ] **Step 1: Create `internal/adapter/lambda/dispatcher.go`**

```go
package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/consumer"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/http/handlers"
)

type Dispatcher struct {
	consumer *consumer.NotificationConsumer
	handler  *handlers.TemplateHandler
}

func NewDispatcher(c *consumer.NotificationConsumer, h *handlers.TemplateHandler) *Dispatcher {
	return &Dispatcher{consumer: c, handler: h}
}

type eventProbe struct {
	Records []struct {
		EventSource string `json:"EventSource"`
	} `json:"Records"`
	HTTPMethod string `json:"httpMethod"`
}

func (d *Dispatcher) Dispatch(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var probe eventProbe
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("failed to probe event type: %w", err)
	}

	if len(probe.Records) > 0 && probe.Records[0].EventSource == "aws:sns" {
		var snsEvent events.SNSEvent
		if err := json.Unmarshal(raw, &snsEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal SNS event: %w", err)
		}
		return nil, d.consumer.Consume(ctx, snsEvent)
	}

	if probe.HTTPMethod != "" {
		var apiEvent events.APIGatewayProxyRequest
		if err := json.Unmarshal(raw, &apiEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal API Gateway event: %w", err)
		}
		response := d.handler.Handle(apiEvent)
		return response, nil
	}

	log.Printf("unknown event type received")
	return nil, fmt.Errorf("unknown event type")
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/adapter/lambda/...
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/lambda/dispatcher.go
git commit -m "feat: add raw JSON Lambda dispatcher"
```

---

### Task 16: Wiring — rewrite main.go

**Files:**
- Modify: `cmd/lambda/main.go`

- [ ] **Step 1: Rewrite `cmd/lambda/main.go`**

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/consumer"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/http/handlers"
	lambdaadapter "github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/lambda"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/mailtrap"
	s3adapter "github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/s3"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/get_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/save_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/send_notification"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	templateRepo := s3adapter.NewTemplateRepository(s3Client, os.Getenv("S3_BUCKET_NAME"))
	emailSender := mailtrap.NewEmailSender(
		os.Getenv("MAILTRAP_TOKEN"),
		os.Getenv("MAILTRAP_URL"),
		os.Getenv("MAILTRAP_FROM_EMAIL"),
		os.Getenv("MAILTRAP_FROM_NAME"),
	)

	sendNotifUC  := send_notification.New(templateRepo, emailSender)
	getTemplUC   := get_template.New(templateRepo)
	saveTemplUC  := save_template.New(templateRepo)

	notifConsumer   := consumer.New(sendNotifUC)
	templateHandler := handlers.NewTemplateHandler(getTemplUC, saveTemplUC)
	dispatcher      := lambdaadapter.NewDispatcher(notifConsumer, templateHandler)

	lambda.Start(dispatcher.Dispatch)
}
```

- [ ] **Step 2: Verify full build**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 3: Run full test suite**

```bash
go test ./...
```

Expected: all tests pass, output similar to:
```
ok  github.com/JulioCVaz/tech-challenge-notification-service/internal/domain
ok  github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/send_notification
ok  github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/get_template
ok  github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/save_template
```

- [ ] **Step 4: Commit**

```bash
git add cmd/lambda/main.go
git commit -m "feat: wire all adapters and use cases in Lambda entry point"
```

---

### Task 17: Finalize — env vars, go mod tidy, final verification

**Files:**
- Modify: `.env.example`

- [ ] **Step 1: Update `.env.example`**

```
AWS_REGION=us-east-1
S3_BUCKET_NAME=tech-challenge-notification-templates

MAILTRAP_TOKEN=your-mailtrap-api-token
MAILTRAP_URL=https://send.api.mailtrap.io/api
MAILTRAP_FROM_EMAIL=contato@nohats.net.br
MAILTRAP_FROM_NAME=Nohats

JWT_SECRET=change-me-in-production
```

- [ ] **Step 2: Run go mod tidy to clean up any indirect deps**

```bash
go mod tidy
```

- [ ] **Step 3: Final build and test**

```bash
go build ./... && go test ./...
```

Expected: clean build, all tests pass.

- [ ] **Step 4: Commit**

```bash
git add .env.example go.mod go.sum
git commit -m "chore: update env vars and tidy go modules"
```

---

## Self-Review

### Spec Coverage

| Requirement | Task |
|-------------|------|
| Lambda listens to SNS `notification-events` | Task 14 (dispatcher) + Task 15 (consumer) |
| Sends emails via Mailtrap | Task 10 (mailtrap adapter) |
| Templates stored in S3 | Task 11 (S3 adapter) |
| Templates manageable via HTTP GET | Task 13 (template handler `handleGet`) |
| Templates manageable via HTTP PUT | Task 13 (template handler `handlePut`) |
| JWT-protected HTTP API | Task 12 (auth middleware) |
| Single Lambda, raw JSON dispatch | Task 15 (dispatcher) |
| Template variables `$Variable` substituted | Task 3 (Template.Render) |
| Missing variables left as literal | Task 3 (test + implementation) |
| `domain.Recipient` value object | Task 4 |
| `ErrTemplateNotFound` sentinel | Task 3 |
| SNS MessageAttribute `templateType` | Task 14 (consumer) |
| S3 object key `templates/{type}.html` | Task 11 |
| S3 subject stored in metadata | Task 11 |
| Use cases instead of services | Tasks 7–9 |
| No infrastructure names in adapter names | All tasks follow naming rule |

### Type Consistency

- `domain.Recipient{Email, Name string}` — defined Task 4, used in Tasks 7, 14
- `domain.Template{Type, Subject, HTML string}` + `Render(map[string]string) string` — defined Task 3, used in Tasks 7, 8, 9, 11
- `domain.ErrTemplateNotFound` — defined Task 3, used in Tasks 8, 11, 13
- `send_notification.UseCase.Execute(templateType string, recipient domain.Recipient, data map[string]string) error` — consistent across Tasks 7, 14
- `get_template.UseCase.Execute(templateType string) (*domain.Template, error)` — consistent across Tasks 8, 13
- `save_template.UseCase.Execute(templateType, subject, html string) error` — consistent across Tasks 9, 13
- `handlers.TemplateHandler.Handle(events.APIGatewayProxyRequest) events.APIGatewayProxyResponse` — consistent Tasks 13, 15
- `consumer.NotificationConsumer.Consume(ctx, events.SNSEvent) error` — consistent Tasks 14, 15
- `dispatcher.Dispatcher.Dispatch(ctx, json.RawMessage) (interface{}, error)` — consistent Tasks 15, 16
