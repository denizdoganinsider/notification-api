package service

import (
	"errors"
	"notification-api/internal/domain"
	"strings"
	"testing"
	"time"
)

type mockNotificationRepo struct {
	notifications map[int64]*domain.Notification
	nextID        int64
}

func newMockNotificationRepo() *mockNotificationRepo {
	return &mockNotificationRepo{
		notifications: make(map[int64]*domain.Notification),
		nextID:        1,
	}
}

func (m *mockNotificationRepo) Create(n *domain.Notification) error {
	n.ID = m.nextID
	n.CreatedAt = time.Now()
	m.nextID++
	m.notifications[n.ID] = n
	return nil
}

func (m *mockNotificationRepo) GetByIDAndUserID(id int64, userID int64) (*domain.Notification, error) {
	n, exists := m.notifications[id]
	if !exists || n.UserID != userID {
		return nil, errors.New("not found")
	}
	return n, nil
}

func (m *mockNotificationRepo) ListByUserID(userID int64, limit int, offset int) ([]domain.Notification, error) {
	var result []domain.Notification
	for _, n := range m.notifications {
		if n.UserID == userID {
			result = append(result, *n)
		}
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockNotificationRepo) ListAll(limit int, offset int) ([]domain.Notification, error) {
	var result []domain.Notification
	for _, n := range m.notifications {
		result = append(result, *n)
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockNotificationRepo) MarkAsRead(id int64, userID int64) error {
	n, exists := m.notifications[id]
	if !exists || n.UserID != userID {
		return errors.New("not found")
	}
	now := time.Now()
	n.ReadAt = &now
	return nil
}

func (m *mockNotificationRepo) DeleteByIDAndUserID(id int64, userID int64) error {
	n, exists := m.notifications[id]
	if !exists || n.UserID != userID {
		return errors.New("not found")
	}
	delete(m.notifications, id)
	return nil
}

func (m *mockNotificationRepo) CountByUserID(userID int64) (int64, error) {
	var count int64
	for _, n := range m.notifications {
		if n.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockNotificationRepo) CountAll() (int64, error) {
	return int64(len(m.notifications)), nil
}

type mockCacheService struct {
	invalidated bool
}

func (m *mockCacheService) InvalidateUserNotifications(userID int64) error {
	m.invalidated = true
	return nil
}

type noopEventService struct{}

func newNoopEventService() *EventService {
	return &EventService{
		cacheService:           NewCacheService(nil),
		webhookDeliveryService: nil,
	}
}

func TestNotificationService_Create_Valid(t *testing.T) {
	repo := newMockNotificationRepo()
	cache := &stubCacheService{}
	event := &EventService{cacheService: &CacheService{}, webhookDeliveryService: nil}
	svc := &NotificationService{
		notificationRepo: repo,
		eventService:     event,
		cacheService:     &CacheService{},
	}
	_ = cache

	n, err := svc.Create(1, "Test Title", "Test Message")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if n.Title != "Test Title" {
		t.Errorf("n.Title = %q, want %q", n.Title, "Test Title")
	}

	if n.UserID != 1 {
		t.Errorf("n.UserID = %d, want %d", n.UserID, 1)
	}
}

func TestNotificationService_Create_EmptyTitle(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := &NotificationService{
		notificationRepo: repo,
		eventService:     &EventService{cacheService: &CacheService{}},
		cacheService:     &CacheService{},
	}

	_, err := svc.Create(1, "", "Test Message")
	if err == nil {
		t.Fatal("Create() should return error for empty title")
	}

	if err.Error() != "title is required" {
		t.Errorf("error = %q, want %q", err.Error(), "title is required")
	}
}

func TestNotificationService_Create_EmptyMessage(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := &NotificationService{
		notificationRepo: repo,
		eventService:     &EventService{cacheService: &CacheService{}},
		cacheService:     &CacheService{},
	}

	_, err := svc.Create(1, "Title", "")
	if err == nil {
		t.Fatal("Create() should return error for empty message")
	}

	if err.Error() != "message is required" {
		t.Errorf("error = %q, want %q", err.Error(), "message is required")
	}
}

func TestNotificationService_Create_TitleTooLong(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := &NotificationService{
		notificationRepo: repo,
		eventService:     &EventService{cacheService: &CacheService{}},
		cacheService:     &CacheService{},
	}

	longTitle := strings.Repeat("a", 256)
	_, err := svc.Create(1, longTitle, "Message")
	if err == nil {
		t.Fatal("Create() should return error for title too long")
	}

	if err.Error() != "title must be at most 255 characters" {
		t.Errorf("error = %q, want %q", err.Error(), "title must be at most 255 characters")
	}
}

func TestNotificationService_Create_MessageTooLong(t *testing.T) {
	repo := newMockNotificationRepo()
	svc := &NotificationService{
		notificationRepo: repo,
		eventService:     &EventService{cacheService: &CacheService{}},
		cacheService:     &CacheService{},
	}

	longMessage := strings.Repeat("a", 65536)
	_, err := svc.Create(1, "Title", longMessage)
	if err == nil {
		t.Fatal("Create() should return error for message too long")
	}

	if err.Error() != "message must be at most 65535 characters" {
		t.Errorf("error = %q, want %q", err.Error(), "message must be at most 65535 characters")
	}
}

func TestNotificationService_List_Pagination(t *testing.T) {
	repo := newMockNotificationRepo()

	for i := 0; i < 15; i++ {
		repo.Create(&domain.Notification{
			UserID:  1,
			Title:   "Title",
			Message: "Message",
		})
	}

	svc := &NotificationService{
		notificationRepo: repo,
		eventService:     &EventService{cacheService: &CacheService{}},
		cacheService:     &CacheService{},
	}

	result, err := svc.List(1, 1, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if result.Total != 15 {
		t.Errorf("Total = %d, want 15", result.Total)
	}

	if result.Page != 1 {
		t.Errorf("Page = %d, want 1", result.Page)
	}

	if result.PerPage != 10 {
		t.Errorf("PerPage = %d, want 10", result.PerPage)
	}

	if len(result.Data) != 10 {
		t.Errorf("len(Data) = %d, want 10", len(result.Data))
	}
}

func TestNotificationService_Delete(t *testing.T) {
	repo := newMockNotificationRepo()
	repo.Create(&domain.Notification{
		UserID:  1,
		Title:   "Title",
		Message: "Message",
	})

	svc := &NotificationService{
		notificationRepo: repo,
		eventService:     &EventService{cacheService: &CacheService{}},
		cacheService:     &CacheService{},
	}

	err := svc.Delete(1, 1)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	count, _ := repo.CountByUserID(1)
	if count != 0 {
		t.Errorf("count after delete = %d, want 0", count)
	}
}

// stubCacheService is used to avoid nil pointer in tests that don't need cache
type stubCacheService struct{}
