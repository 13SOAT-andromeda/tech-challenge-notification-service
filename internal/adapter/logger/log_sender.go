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
		notification.Type, notification.Recipient, notification.Subject)
	return nil
}
