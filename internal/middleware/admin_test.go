package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestAdminMiddleware_AllowAdmin(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "admin")

	handler := AdminMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAdminMiddleware_RejectNonAdmin(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("role", "user")

	handler := AdminMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	_ = handler(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAdminMiddleware_RejectNoRole(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := AdminMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	_ = handler(c)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
