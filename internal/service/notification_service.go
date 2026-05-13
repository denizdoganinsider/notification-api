package service

import (
	"errors"
	"notification-api/internal/domain"
	"notification-api/internal/repository"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	eventService     *EventService
	cacheService     *CacheService
}

func NewNotificationService(
	notificationRepo *repository.NotificationRepository,
	eventService *EventService,
	cacheService *CacheService,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		eventService:     eventService,
		cacheService:     cacheService,
	}
}

func (s *NotificationService) Create(userID int64, title string, message string) (*domain.Notification, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}

	if message == "" {
		return nil, errors.New("message is required")
	}

	notification := &domain.Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
	}

	err := s.notificationRepo.Create(notification)
	if err != nil {
		return nil, err
	}

	event := domain.NotificationCreatedEvent{
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		Title:          notification.Title,
		Message:        notification.Message,
		CreatedAt:      notification.CreatedAt,
	}

	_ = s.eventService.HandleNotificationCreated(event)

	return notification, nil
}

func (s *NotificationService) List(userID int64, page int, perPage int) ([]domain.Notification, error) {
	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	cached, err := s.cacheService.GetNotifications(userID, page, perPage)
	if err == nil {
		return cached, nil
	}

	offset := (page - 1) * perPage

	notifications, err := s.notificationRepo.ListByUserID(userID, perPage, offset)
	if err != nil {
		return nil, err
	}

	_ = s.cacheService.SetNotifications(userID, page, perPage, notifications)

	return notifications, nil
}

func (s *NotificationService) GetByID(id int64, userID int64) (*domain.Notification, error) {
	return s.notificationRepo.GetByIDAndUserID(id, userID)
}

func (s *NotificationService) ListAll(page int, perPage int) ([]domain.Notification, error) {
	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	if perPage > 100 {
		perPage = 100
	}

	offset := (page - 1) * perPage

	return s.notificationRepo.ListAll(perPage, offset)
}
