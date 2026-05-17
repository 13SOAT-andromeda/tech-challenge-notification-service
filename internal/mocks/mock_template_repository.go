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
