package service

import "notification-api/internal/domain"

type EventService struct {
	cacheService           *CacheService
	webhookDeliveryService *WebhookDeliveryService
}

func NewEventService(
	cacheService *CacheService,
	webhookDeliveryService *WebhookDeliveryService,
) *EventService {
	return &EventService{
		cacheService:           cacheService,
		webhookDeliveryService: webhookDeliveryService,
	}
}

func (s *EventService) HandleNotificationCreated(event domain.NotificationCreatedEvent) error {
	err := s.cacheService.InvalidateUserNotifications(event.UserID)
	if err != nil {
		return err
	}

	err = s.webhookDeliveryService.SendNotificationCreated(event)
	if err != nil {
		return err
	}

	return nil
}
