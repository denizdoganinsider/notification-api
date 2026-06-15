package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"notification-api/internal/domain"
	"notification-api/internal/repository"
	"time"
)

const maxRetries = 3

type WebhookDeliveryService struct {
	webhookRepo repository.WebhookRepositoryInterface
	httpClient  *http.Client
}

func NewWebhookDeliveryService(webhookRepo repository.WebhookRepositoryInterface) *WebhookDeliveryService {
	return &WebhookDeliveryService{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *WebhookDeliveryService) SendNotificationCreated(event domain.NotificationCreatedEvent) error {
	webhooks, err := s.webhookRepo.ListByUserID(event.UserID)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"event": "notification.created",
		"data":  event,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	for _, webhook := range webhooks {
		s.deliverWithRetry(webhook, body)
	}

	return nil
}

func (s *WebhookDeliveryService) deliverWithRetry(webhook domain.Webhook, body []byte) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}

		req, err := http.NewRequest(http.MethodPost, webhook.URL, bytes.NewBuffer(body))
		if err != nil {
			slog.Error("webhook delivery failed",
				"webhook_id", webhook.ID,
				"url", webhook.URL,
				"error", err,
			)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		resp.Body.Close()

		if resp.StatusCode < 500 {
			return
		}

		lastErr = fmt.Errorf("server error: status=%d", resp.StatusCode)
	}

	slog.Error("webhook delivery failed after retries",
		"retries", maxRetries,
		"webhook_id", webhook.ID,
		"url", webhook.URL,
		"error", lastErr,
	)
}
