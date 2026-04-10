package domain

import "time"

type Webhook struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
