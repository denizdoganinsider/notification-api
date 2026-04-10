package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"notification-api/internal/domain"
	"notification-api/internal/repository"
	"time"
)

type WebhookDeliveryService struct {
	webhookRepo *repository.WebhookRepository
	httpClient  *http.Client
}

func NewWebhookDeliveryService(webhookRepo *repository.WebhookRepository) *WebhookDeliveryService {
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
		req, err := http.NewRequest(http.MethodPost, webhook.URL, bytes.NewBuffer(body))
		if err != nil {
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}

		resp.Body.Close()
	}

	return nil
}
