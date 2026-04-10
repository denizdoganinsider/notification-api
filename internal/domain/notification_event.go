package domain

import "time"

type NotificationCreatedEvent struct {
	NotificationID int64     `json:"notification_id"`
	UserID         int64     `json:"user_id"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}
