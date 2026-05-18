package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/send_notification"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type NotificationConsumer struct {
	useCase *send_notification.UseCase
}

func New(useCase *send_notification.UseCase) *NotificationConsumer {
	return &NotificationConsumer{useCase: useCase}
}

type notificationBody struct {
	Recipient struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"recipient"`
	Data map[string]string `json:"data"`
}

func (c *NotificationConsumer) Consume(ctx context.Context, event events.SNSEvent) error {
	for _, record := range event.Records {
		log.Printf("processing SNS message: %s", record.SNS.MessageID)

		templateTypeAttr, ok := record.SNS.MessageAttributes["templateType"]
		if !ok {
			return fmt.Errorf("missing templateType message attribute in record %s", record.SNS.MessageID)
		}

		// MessageAttributes is map[string]interface{}, extract the Value field
		attrMap, ok := templateTypeAttr.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid templateType message attribute format in record %s", record.SNS.MessageID)
		}

		templateTypeVal, ok := attrMap["Value"]
		if !ok {
			return fmt.Errorf("missing Value in templateType message attribute in record %s", record.SNS.MessageID)
		}

		templateType, ok := templateTypeVal.(string)
		if !ok {
			return fmt.Errorf("templateType value is not a string in record %s", record.SNS.MessageID)
		}

		var body notificationBody
		if err := json.Unmarshal([]byte(record.SNS.Message), &body); err != nil {
			return fmt.Errorf("failed to parse SNS message body in record %s: %w", record.SNS.MessageID, err)
		}

		recipient := domain.Recipient{
			Email: body.Recipient.Email,
			Name:  body.Recipient.Name,
		}

		if err := c.useCase.Execute(templateType, recipient, body.Data); err != nil {
			return fmt.Errorf("failed to process notification in record %s: %w", record.SNS.MessageID, err)
		}

		log.Printf("notification processed successfully: %s", record.SNS.MessageID)
	}
	return nil
}
