package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/consumer"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/http/handlers"
	lambdaadapter "github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/lambda"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/mailtrap"
	s3adapter "github.com/JulioCVaz/tech-challenge-notification-service/internal/adapter/s3"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/get_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/save_template"
	"github.com/JulioCVaz/tech-challenge-notification-service/internal/application/usecases/send_notification"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	templateRepo := s3adapter.NewTemplateRepository(s3Client, os.Getenv("S3_BUCKET_NAME"))
	emailSender := mailtrap.NewEmailSender(
		os.Getenv("MAILTRAP_TOKEN"),
		os.Getenv("MAILTRAP_URL"),
		os.Getenv("MAILTRAP_FROM_EMAIL"),
		os.Getenv("MAILTRAP_FROM_NAME"),
	)

	sendNotifUC := send_notification.New(templateRepo, emailSender)
	getTemplUC  := get_template.New(templateRepo)
	saveTemplUC := save_template.New(templateRepo)

	notifConsumer   := consumer.New(sendNotifUC)
	templateHandler := handlers.NewTemplateHandler(getTemplUC, saveTemplUC)
	dispatcher      := lambdaadapter.NewDispatcher(notifConsumer, templateHandler)

	lambda.Start(dispatcher.Dispatch)
}
