package service

import (
	"errors"
	"notification-api/internal/domain"
	"testing"
	"time"
)

type mockWebhookRepo struct {
	webhooks map[int64]*domain.Webhook
	nextID   int64
}

func newMockWebhookRepo() *mockWebhookRepo {
	return &mockWebhookRepo{
		webhooks: make(map[int64]*domain.Webhook),
		nextID:   1,
	}
}

func (m *mockWebhookRepo) Create(w *domain.Webhook) error {
	w.ID = m.nextID
	w.CreatedAt = time.Now()
	m.nextID++
	m.webhooks[w.ID] = w
	return nil
}

func (m *mockWebhookRepo) ListByUserID(userID int64) ([]domain.Webhook, error) {
	var result []domain.Webhook
	for _, w := range m.webhooks {
		if w.UserID == userID {
			result = append(result, *w)
		}
	}
	return result, nil
}

func (m *mockWebhookRepo) DeleteByIDAndUserID(id int64, userID int64) error {
	w, exists := m.webhooks[id]
	if !exists || w.UserID != userID {
		return errors.New("not found")
	}
	delete(m.webhooks, id)
	return nil
}

func TestWebhookService_Create_Valid(t *testing.T) {
	repo := newMockWebhookRepo()
	svc := NewWebhookService(repo)

	w, err := svc.Create(1, "https://example.com/webhook")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if w.URL != "https://example.com/webhook" {
		t.Errorf("w.URL = %q, want %q", w.URL, "https://example.com/webhook")
	}

	if w.UserID != 1 {
		t.Errorf("w.UserID = %d, want %d", w.UserID, 1)
	}

	if w.ID == 0 {
		t.Error("w.ID should not be 0")
	}
}

func TestWebhookService_Create_EmptyURL(t *testing.T) {
	repo := newMockWebhookRepo()
	svc := NewWebhookService(repo)

	_, err := svc.Create(1, "")
	if err == nil {
		t.Fatal("Create() should return error for empty URL")
	}

	if err.Error() != "webhook url is required" {
		t.Errorf("error = %q, want %q", err.Error(), "webhook url is required")
	}
}

func TestWebhookService_Create_InvalidURL(t *testing.T) {
	repo := newMockWebhookRepo()
	svc := NewWebhookService(repo)

	tests := []struct {
		name string
		url  string
	}{
		{"no scheme", "example.com/webhook"},
		{"invalid format", "not-a-url"},
		{"missing host", "https://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Create(1, tt.url)
			if err == nil {
				t.Fatal("Create() should return error for invalid URL")
			}
			if err.Error() != "invalid webhook url" {
				t.Errorf("error = %q, want %q", err.Error(), "invalid webhook url")
			}
		})
	}
}

func TestWebhookService_List(t *testing.T) {
	repo := newMockWebhookRepo()
	svc := NewWebhookService(repo)

	_, _ = svc.Create(1, "https://example.com/hook1")
	_, _ = svc.Create(1, "https://example.com/hook2")
	_, _ = svc.Create(2, "https://example.com/hook3")

	webhooks, err := svc.List(1)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(webhooks) != 2 {
		t.Errorf("len(webhooks) = %d, want 2", len(webhooks))
	}
}

func TestWebhookService_Delete(t *testing.T) {
	repo := newMockWebhookRepo()
	svc := NewWebhookService(repo)

	w, _ := svc.Create(1, "https://example.com/hook")

	err := svc.Delete(w.ID, 1)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	webhooks, _ := svc.List(1)
	if len(webhooks) != 0 {
		t.Errorf("len(webhooks) after delete = %d, want 0", len(webhooks))
	}
}

func TestWebhookService_Delete_WrongUser(t *testing.T) {
	repo := newMockWebhookRepo()
	svc := NewWebhookService(repo)

	w, _ := svc.Create(1, "https://example.com/hook")

	err := svc.Delete(w.ID, 2)
	if err == nil {
		t.Fatal("Delete() should return error when user doesn't own webhook")
	}
}
