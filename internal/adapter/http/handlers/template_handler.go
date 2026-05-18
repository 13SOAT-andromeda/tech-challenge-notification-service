package handlers

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/http/middlewares"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/get_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/save_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type TemplateHandler struct {
	getTemplate  *get_template.UseCase
	saveTemplate *save_template.UseCase
}

func NewTemplateHandler(getTemplate *get_template.UseCase, saveTemplate *save_template.UseCase) *TemplateHandler {
	return &TemplateHandler{
		getTemplate:  getTemplate,
		saveTemplate: saveTemplate,
	}
}

type saveTemplateRequest struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

func (h *TemplateHandler) Handle(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	if resp := middlewares.ValidateJWT(request); resp != nil {
		return *resp
	}

	templateType := request.PathParameters["type"]

	switch request.HTTPMethod {
	case "GET":
		return h.handleGet(templateType)
	case "PUT":
		return h.handlePut(templateType, request.Body)
	default:
		return jsonResponse(405, map[string]interface{}{"success": false, "error": "method not allowed"})
	}
}

func (h *TemplateHandler) handleGet(templateType string) events.APIGatewayProxyResponse {
	tmpl, err := h.getTemplate.Execute(templateType)
	if err != nil {
		if errors.Is(err, domain.ErrTemplateNotFound) {
			return jsonResponse(404, map[string]interface{}{"success": false, "error": "template not found"})
		}
		return jsonResponse(500, map[string]interface{}{"success": false, "error": "internal server error"})
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "text/html"},
		Body:       tmpl.HTML,
	}
}

func (h *TemplateHandler) handlePut(templateType, body string) events.APIGatewayProxyResponse {
	var req saveTemplateRequest
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		return jsonResponse(400, map[string]interface{}{"success": false, "error": "invalid request body"})
	}

	if err := h.saveTemplate.Execute(templateType, req.Subject, req.HTML); err != nil {
		return jsonResponse(400, map[string]interface{}{"success": false, "error": err.Error()})
	}

	return jsonResponse(200, map[string]interface{}{"success": true, "message": "template saved"})
}

func jsonResponse(statusCode int, body interface{}) events.APIGatewayProxyResponse {
	b, _ := json.Marshal(body)
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(b),
	}
}
