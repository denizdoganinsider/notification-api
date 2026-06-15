//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"notification-api/config"
	"notification-api/internal/controller"
	"notification-api/internal/middleware"
	"notification-api/internal/repository"
	"notification-api/internal/service"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func setupTestServer(t *testing.T) *echo.Echo {
	t.Helper()

	cfg := config.LoadConfig()
	middleware.InitJWT(cfg.JWTSecret)

	db := config.NewDatabase(cfg)
	t.Cleanup(func() { db.Close() })

	redisClient := config.NewRedisClient(cfg)
	t.Cleanup(func() { redisClient.Close() })

	// Clean up tables for a fresh test run
	db.Exec("DELETE FROM webhooks")
	db.Exec("DELETE FROM notifications")
	db.Exec("DELETE FROM users")

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	notificationRepo := repository.NewNotificationRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	cacheService := service.NewCacheService(redisClient)
	webhookDeliveryService := service.NewWebhookDeliveryService(webhookRepo)
	eventService := service.NewEventService(cacheService, webhookDeliveryService)

	notificationService := service.NewNotificationService(notificationRepo, eventService, cacheService)
	notificationController := controller.NewNotificationController(notificationService)

	webhookService := service.NewWebhookService(webhookRepo)
	webhookController := controller.NewWebhookController(webhookService)

	adminController := controller.NewAdminController(userService, notificationService)

	e := echo.New()

	e.Use(middleware.RequestIDMiddleware)
	e.Use(middleware.LoggerMiddleware)
	e.Use(middleware.RateLimiterMiddleware(redisClient, 100, 1*time.Minute))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/register", userController.Register)
	e.POST("/login", userController.Login)

	auth := e.Group("")
	auth.Use(middleware.JWTMiddleware)

	auth.GET("/me", userController.Me)
	auth.POST("/notifications", notificationController.Create)
	auth.GET("/notifications", notificationController.List)
	auth.GET("/notifications/:id", notificationController.GetByID)
	auth.PUT("/notifications/:id/read", notificationController.MarkAsRead)
	auth.DELETE("/notifications/:id", notificationController.Delete)

	auth.POST("/webhooks", webhookController.Create)
	auth.GET("/webhooks", webhookController.List)
	auth.DELETE("/webhooks/:id", webhookController.Delete)

	admin := auth.Group("/admin")
	admin.Use(middleware.AdminMiddleware)
	admin.GET("/users", adminController.ListUsers)
	admin.GET("/notifications", adminController.ListNotifications)

	return e
}

func doRequest(e *echo.Echo, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	e := setupTestServer(t)

	rec := doRequest(e, http.MethodGet, "/health", nil, "")

	if rec.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want %q", resp["status"], "ok")
	}
}

func TestRegisterAndLogin(t *testing.T) {
	e := setupTestServer(t)

	// Register
	rec := doRequest(e, http.MethodPost, "/register", map[string]string{
		"email":    "integration@example.com",
		"password": "Password1",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /register status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	// Login
	rec = doRequest(e, http.MethodPost, "/login", map[string]string{
		"email":    "integration@example.com",
		"password": "Password1",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /login status = %d, want %d", rec.Code, http.StatusOK)
	}

	var loginResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	if loginResp["token"] == "" {
		t.Fatal("login response should contain a token")
	}
}

func TestRegisterValidation(t *testing.T) {
	e := setupTestServer(t)

	// Weak password
	rec := doRequest(e, http.MethodPost, "/register", map[string]string{
		"email":    "test@example.com",
		"password": "weak",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /register with weak password status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	// Invalid email
	rec = doRequest(e, http.MethodPost, "/register", map[string]string{
		"email":    "not-an-email",
		"password": "Password1",
	}, "")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /register with invalid email status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestNotificationsCRUD(t *testing.T) {
	e := setupTestServer(t)

	// Register and login
	doRequest(e, http.MethodPost, "/register", map[string]string{
		"email":    "notify@example.com",
		"password": "Password1",
	}, "")

	rec := doRequest(e, http.MethodPost, "/login", map[string]string{
		"email":    "notify@example.com",
		"password": "Password1",
	}, "")

	var loginResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	token := loginResp["token"]

	// Create notification
	rec = doRequest(e, http.MethodPost, "/notifications", map[string]string{
		"title":   "Test Notification",
		"message": "Hello World",
	}, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /notifications status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var notification map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &notification)
	notificationID := fmt.Sprintf("%.0f", notification["id"].(float64))

	// List notifications
	rec = doRequest(e, http.MethodGet, "/notifications", nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /notifications status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Get by ID
	rec = doRequest(e, http.MethodGet, "/notifications/"+notificationID, nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /notifications/%s status = %d, want %d", notificationID, rec.Code, http.StatusOK)
	}

	// Mark as read
	rec = doRequest(e, http.MethodPut, "/notifications/"+notificationID+"/read", nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("PUT /notifications/%s/read status = %d, want %d", notificationID, rec.Code, http.StatusOK)
	}

	// Delete
	rec = doRequest(e, http.MethodDelete, "/notifications/"+notificationID, nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("DELETE /notifications/%s status = %d, want %d", notificationID, rec.Code, http.StatusOK)
	}
}

func TestNotificationValidation(t *testing.T) {
	e := setupTestServer(t)

	// Register and login
	doRequest(e, http.MethodPost, "/register", map[string]string{
		"email":    "validate@example.com",
		"password": "Password1",
	}, "")

	rec := doRequest(e, http.MethodPost, "/login", map[string]string{
		"email":    "validate@example.com",
		"password": "Password1",
	}, "")

	var loginResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	token := loginResp["token"]

	// Empty title
	rec = doRequest(e, http.MethodPost, "/notifications", map[string]string{
		"title":   "",
		"message": "Hello",
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /notifications empty title status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	// Empty message
	rec = doRequest(e, http.MethodPost, "/notifications", map[string]string{
		"title":   "Title",
		"message": "",
	}, token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /notifications empty message status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestWebhooksCRUD(t *testing.T) {
	e := setupTestServer(t)

	// Register and login
	doRequest(e, http.MethodPost, "/register", map[string]string{
		"email":    "webhooks@example.com",
		"password": "Password1",
	}, "")

	rec := doRequest(e, http.MethodPost, "/login", map[string]string{
		"email":    "webhooks@example.com",
		"password": "Password1",
	}, "")

	var loginResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	token := loginResp["token"]

	// Create webhook
	rec = doRequest(e, http.MethodPost, "/webhooks", map[string]string{
		"url": "https://example.com/webhook",
	}, token)

	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /webhooks status = %d, want %d, body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var webhook map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &webhook)
	webhookID := fmt.Sprintf("%.0f", webhook["id"].(float64))

	// List webhooks
	rec = doRequest(e, http.MethodGet, "/webhooks", nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /webhooks status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Delete webhook
	rec = doRequest(e, http.MethodDelete, "/webhooks/"+webhookID, nil, token)
	if rec.Code != http.StatusOK {
		t.Errorf("DELETE /webhooks/%s status = %d, want %d", webhookID, rec.Code, http.StatusOK)
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	e := setupTestServer(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/me"},
		{http.MethodPost, "/notifications"},
		{http.MethodGet, "/notifications"},
		{http.MethodPost, "/webhooks"},
		{http.MethodGet, "/webhooks"},
		{http.MethodGet, "/admin/users"},
		{http.MethodGet, "/admin/notifications"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			rec := doRequest(e, ep.method, ep.path, nil, "")
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("%s %s status = %d, want %d", ep.method, ep.path, rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAdminEndpoints_NonAdmin(t *testing.T) {
	e := setupTestServer(t)

	// Register regular user and login
	doRequest(e, http.MethodPost, "/register", map[string]string{
		"email":    "regular@example.com",
		"password": "Password1",
	}, "")

	rec := doRequest(e, http.MethodPost, "/login", map[string]string{
		"email":    "regular@example.com",
		"password": "Password1",
	}, "")

	var loginResp map[string]string
	json.Unmarshal(rec.Body.Bytes(), &loginResp)
	token := loginResp["token"]

	// Try admin endpoints
	rec = doRequest(e, http.MethodGet, "/admin/users", nil, token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("GET /admin/users status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	rec = doRequest(e, http.MethodGet, "/admin/notifications", nil, token)
	if rec.Code != http.StatusForbidden {
		t.Errorf("GET /admin/notifications status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRequestIDHeader(t *testing.T) {
	e := setupTestServer(t)

	rec := doRequest(e, http.MethodGet, "/health", nil, "")

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header should be present in response")
	}
}
