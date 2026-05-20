package send_notification

import (
	"fmt"
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
	if recipient.Email == "" {
		return fmt.Errorf("recipient email is required")
	}

	tmpl, err := uc.repo.Get(templateType)
	if err != nil {
		return err
	}

	merged := make(map[string]string, len(data)+2)
	for k, v := range data {
		merged[k] = v
	}
	merged["name"] = recipient.Name
	merged["email"] = recipient.Email

	renderedHTML := tmpl.Render(merged)

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
