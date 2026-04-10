package service

import (
	"errors"
	"net/url"
	"notification-api/internal/domain"
	"notification-api/internal/repository"
)

type WebhookService struct {
	webhookRepo *repository.WebhookRepository
}

func NewWebhookService(webhookRepo *repository.WebhookRepository) *WebhookService {
	return &WebhookService{
		webhookRepo: webhookRepo,
	}
}

func (s *WebhookService) Create(userID int64, webhookURL string) (*domain.Webhook, error) {
	if webhookURL == "" {
		return nil, errors.New("webhook url is required")
	}

	parsedURL, err := url.ParseRequestURI(webhookURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return nil, errors.New("invalid webhook url")
	}

	webhook := &domain.Webhook{
		UserID: userID,
		URL:    webhookURL,
	}

	err = s.webhookRepo.Create(webhook)
	if err != nil {
		return nil, err
	}

	return webhook, nil
}

func (s *WebhookService) List(userID int64) ([]domain.Webhook, error) {
	return s.webhookRepo.ListByUserID(userID)
}

func (s *WebhookService) Delete(id int64, userID int64) error {
	return s.webhookRepo.DeleteByIDAndUserID(id, userID)
}
