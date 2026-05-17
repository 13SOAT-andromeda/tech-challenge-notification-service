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
