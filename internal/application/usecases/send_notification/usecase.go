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
