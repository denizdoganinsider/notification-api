package repository

import "notification-api/internal/domain"

type UserRepositoryInterface interface {
	Create(user *domain.User) error
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int64) (*domain.User, error)
	GetAll() ([]domain.User, error)
}

type NotificationRepositoryInterface interface {
	Create(notification *domain.Notification) error
	GetByIDAndUserID(id int64, userID int64) (*domain.Notification, error)
	ListByUserID(userID int64, limit int, offset int) ([]domain.Notification, error)
	ListAll(limit int, offset int) ([]domain.Notification, error)
	MarkAsRead(id int64, userID int64) error
	DeleteByIDAndUserID(id int64, userID int64) error
	CountByUserID(userID int64) (int64, error)
	CountAll() (int64, error)
}

type WebhookRepositoryInterface interface {
	Create(webhook *domain.Webhook) error
	ListByUserID(userID int64) ([]domain.Webhook, error)
	DeleteByIDAndUserID(id int64, userID int64) error
}
