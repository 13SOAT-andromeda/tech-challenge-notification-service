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
	mockRepo.AssertExpectations(t)
}
