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
	mockRepo.AssertExpectations(t)
}
