package ports

import "github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"

type NotificationSender interface {
	Send(notification *domain.Notification) error
}
