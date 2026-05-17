package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type TemplateRepository struct {
	client *s3.Client
	bucket string
}

func NewTemplateRepository(client *s3.Client, bucket string) *TemplateRepository {
	return &TemplateRepository{client: client, bucket: bucket}
}

func (r *TemplateRepository) objectKey(templateType string) string {
	return "templates/" + templateType + ".html"
}

func (r *TemplateRepository) Get(templateType string) (*domain.Template, error) {
	result, err := r.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(r.objectKey(templateType)),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("failed to get template from S3: %w", err)
	}
	defer result.Body.Close()

	htmlBytes, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read template body: %w", err)
	}

	subject := ""
	if result.Metadata != nil {
		subject = result.Metadata["subject"]
	}

	return &domain.Template{
		Type:    templateType,
		Subject: subject,
		HTML:    string(htmlBytes),
	}, nil
}

func (r *TemplateRepository) Save(template *domain.Template) error {
	_, err := r.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(r.objectKey(template.Type)),
		Body:        bytes.NewReader([]byte(template.HTML)),
		ContentType: aws.String("text/html"),
		Metadata:    map[string]string{"subject": template.Subject},
	})
	if err != nil {
		return fmt.Errorf("failed to save template to S3: %w", err)
	}
	return nil
}
