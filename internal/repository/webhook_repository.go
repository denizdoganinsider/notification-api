package repository

import (
	"database/sql"
	"notification-api/internal/domain"
)

type WebhookRepository struct {
	db *sql.DB
}

func NewWebhookRepository(db *sql.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

func (r *WebhookRepository) Create(webhook *domain.Webhook) error {
	query := `
	INSERT INTO webhooks (user_id, url)
	VALUES (?, ?)
	`

	result, err := r.db.Exec(query, webhook.UserID, webhook.URL)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	webhook.ID = id

	return nil
}

func (r *WebhookRepository) ListByUserID(userID int64) ([]domain.Webhook, error) {
	query := `
	SELECT id, user_id, url, created_at
	FROM webhooks
	WHERE user_id = ?
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []domain.Webhook

	for rows.Next() {
		var webhook domain.Webhook

		err := rows.Scan(
			&webhook.ID,
			&webhook.UserID,
			&webhook.URL,
			&webhook.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		webhooks = append(webhooks, webhook)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (r *WebhookRepository) DeleteByIDAndUserID(id int64, userID int64) error {
	query := `
	DELETE FROM webhooks
	WHERE id = ? AND user_id = ?
	`

	_, err := r.db.Exec(query, id, userID)

	return err
}
