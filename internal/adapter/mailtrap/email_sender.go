package mailtrap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/JulioCVaz/tech-challenge-notification-service/internal/domain"
)

type EmailSender struct {
	apiKey    string
	apiUrl    string
	fromEmail string
	fromName  string
}

func NewEmailSender(apiKey, apiUrl, fromEmail, fromName string) *EmailSender {
	return &EmailSender{
		apiKey:    apiKey,
		apiUrl:    apiUrl,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

type contact struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type sendPayload struct {
	From    contact   `json:"from"`
	To      []contact `json:"to"`
	Subject string    `json:"subject"`
	Text    string    `json:"text"`
	Html    string    `json:"html"`
}

func (s *EmailSender) Send(notification *domain.Notification) error {
	p := sendPayload{
		From:    contact{Email: s.fromEmail, Name: s.fromName},
		To:      []contact{{Email: notification.Recipient.Email, Name: notification.Recipient.Name}},
		Subject: notification.Subject,
		Text:    notification.Subject,
		Html:    notification.Body,
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.apiUrl+"/send", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Api-Token", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("mailtrap request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("mailtrap error %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
