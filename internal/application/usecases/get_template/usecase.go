package get_template

import (
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/ports"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type UseCase struct {
	repo ports.TemplateRepository
}

func New(repo ports.TemplateRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(templateType string) (*domain.Template, error) {
	return uc.repo.Get(templateType)
}
