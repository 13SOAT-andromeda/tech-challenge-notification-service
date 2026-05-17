package mocks

import (
	"github.com/stretchr/testify/mock"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) Send(notification *domain.Notification) error {
	args := m.Called(notification)
	return args.Error(0)
}
