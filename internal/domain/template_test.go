package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

func TestTemplate_Render_SubstitutesAllVariables(t *testing.T) {
	tmpl := &domain.Template{
		Type:    "order-approval",
		Subject: "Aprovação de OS",
		HTML:    "<p>Olá $Name, seu pedido $ID vale $Value</p>",
	}
	result := tmpl.Render(map[string]string{
		"Name":  "João",
		"ID":    "OS-001",
		"Value": "R$ 350,00",
	})
	assert.Equal(t, "<p>Olá João, seu pedido OS-001 vale R$ 350,00</p>", result)
}

func TestTemplate_Render_LeavesMissingVariablesAsLiteral(t *testing.T) {
	tmpl := &domain.Template{
		HTML: "<p>Olá $Name, código: $Code</p>",
	}
	result := tmpl.Render(map[string]string{"Name": "João"})
	assert.Equal(t, "<p>Olá João, código: $Code</p>", result)
}

func TestTemplate_Render_EmptyDataReturnsOriginalHTML(t *testing.T) {
	tmpl := &domain.Template{
		HTML: "<p>Olá $Name</p>",
	}
	result := tmpl.Render(map[string]string{})
	assert.Equal(t, "<p>Olá $Name</p>", result)
}
