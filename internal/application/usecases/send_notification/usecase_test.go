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
	mockRepo.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}
