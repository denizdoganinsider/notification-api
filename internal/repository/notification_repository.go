package repository

import (
	"database/sql"
	"notification-api/internal/domain"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notification *domain.Notification) error {
	query := `
	INSERT INTO notifications (user_id, title, message)
	VALUES (?, ?, ?)
	`

	result, err := r.db.Exec(query, notification.UserID, notification.Title, notification.Message)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	notification.ID = id

	return nil
}

func (r *NotificationRepository) GetByIDAndUserID(id int64, userID int64) (*domain.Notification, error) {
	query := `
	SELECT id, user_id, title, message, read_at, created_at
	FROM notifications
	WHERE id = ? AND user_id = ?
	`

	row := r.db.QueryRow(query, id, userID)

	var notification domain.Notification

	err := row.Scan(
		&notification.ID,
		&notification.UserID,
		&notification.Title,
		&notification.Message,
		&notification.ReadAt,
		&notification.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &notification, nil
}

func (r *NotificationRepository) ListByUserID(userID int64, limit int, offset int) ([]domain.Notification, error) {
	query := `
	SELECT id, user_id, title, message, read_at, created_at
	FROM notifications
	WHERE user_id = ?
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.Notification

	for rows.Next() {
		var notification domain.Notification

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Message,
			&notification.ReadAt,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) ListAll(limit int, offset int) ([]domain.Notification, error) {
	query := `
	SELECT id, user_id, title, message, read_at, created_at
	FROM notifications
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []domain.Notification

	for rows.Next() {
		var notification domain.Notification

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.Title,
			&notification.Message,
			&notification.ReadAt,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepository) MarkAsRead(id int64, userID int64) error {
	query := `
	UPDATE notifications
	SET read_at = NOW()
	WHERE id = ? AND user_id = ? AND read_at IS NULL
	`

	_, err := r.db.Exec(query, id, userID)
	return err
}

func (r *NotificationRepository) DeleteByIDAndUserID(id int64, userID int64) error {
	query := `
	DELETE FROM notifications
	WHERE id = ? AND user_id = ?
	`

	_, err := r.db.Exec(query, id, userID)
	return err
}

func (r *NotificationRepository) CountByUserID(userID int64) (int64, error) {
	query := `
	SELECT COUNT(*) FROM notifications WHERE user_id = ?
	`

	var count int64
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

func (r *NotificationRepository) CountAll() (int64, error) {
	query := `
	SELECT COUNT(*) FROM notifications
	`

	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
