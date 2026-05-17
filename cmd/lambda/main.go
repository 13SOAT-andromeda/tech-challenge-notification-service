package main

import (
	"github.com/aws/aws-lambda-go/lambda"

	adapterlambda "github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/lambda"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/logger"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/services"
)

func main() {
	sender := logger.NewLogSender()
	svc := services.NewNotificationService(sender)
	handler := adapterlambda.NewHandler(svc)
	lambda.Start(handler.Handle)
}
