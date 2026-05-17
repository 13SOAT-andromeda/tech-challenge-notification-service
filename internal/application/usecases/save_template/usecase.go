package save_template

import (
	"errors"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/ports"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type UseCase struct {
	repo ports.TemplateRepository
}

func New(repo ports.TemplateRepository) *UseCase {
	return &UseCase{repo: repo}
}

func (uc *UseCase) Execute(templateType, subject, html string) error {
	if templateType == "" {
		return errors.New("template type is required")
	}
	if subject == "" {
		return errors.New("subject is required")
	}
	if html == "" {
		return errors.New("html is required")
	}
	return uc.repo.Save(&domain.Template{
		Type:    templateType,
		Subject: subject,
		HTML:    html,
	})
}
