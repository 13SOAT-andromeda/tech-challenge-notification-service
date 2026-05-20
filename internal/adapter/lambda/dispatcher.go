package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/consumer"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/http/handlers"
)

type Dispatcher struct {
	consumer *consumer.NotificationConsumer
	handler  *handlers.TemplateHandler
}

func NewDispatcher(c *consumer.NotificationConsumer, h *handlers.TemplateHandler) *Dispatcher {
	return &Dispatcher{consumer: c, handler: h}
}

type eventProbe struct {
	Records []struct {
		EventSource string `json:"eventSource"`
	} `json:"Records"`
	HTTPMethod string `json:"httpMethod"`
}

func (d *Dispatcher) Dispatch(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var probe eventProbe
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil, fmt.Errorf("failed to probe event type: %w", err)
	}

	if len(probe.Records) > 0 {
		switch probe.Records[0].EventSource {
		case "aws:sns":
			var snsEvent events.SNSEvent
			if err := json.Unmarshal(raw, &snsEvent); err != nil {
				return nil, fmt.Errorf("failed to unmarshal SNS event: %w", err)
			}
			return nil, d.consumer.Consume(ctx, snsEvent)
		case "aws:sqs":
			var sqsEvent events.SQSEvent
			if err := json.Unmarshal(raw, &sqsEvent); err != nil {
				return nil, fmt.Errorf("failed to unmarshal SQS event: %w", err)
			}
			return nil, d.consumer.ConsumeSQS(ctx, sqsEvent)
		}
	}

	if probe.HTTPMethod != "" {
		var apiEvent events.APIGatewayProxyRequest
		if err := json.Unmarshal(raw, &apiEvent); err != nil {
			return nil, fmt.Errorf("failed to unmarshal API Gateway event: %w", err)
		}
		response := d.handler.Handle(apiEvent)
		return response, nil
	}

	log.Printf("unknown event type received")
	return nil, fmt.Errorf("unknown event type")
}
