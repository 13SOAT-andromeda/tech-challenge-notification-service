package ports

import "github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"

type TemplateRepository interface {
	Get(templateType string) (*domain.Template, error)
	Save(template *domain.Template) error
}
