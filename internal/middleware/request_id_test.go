package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRequestIDMiddleware_GeneratesWhenAbsent(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestIDMiddleware(func(c echo.Context) error {
		requestID := c.Get(RequestIDKey).(string)
		if requestID == "" {
			t.Error("request_id should not be empty")
		}
		if len(requestID) != 32 { // 16 bytes = 32 hex chars
			t.Errorf("request_id length = %d, want 32", len(requestID))
		}
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}

	responseID := rec.Header().Get(RequestIDHeader)
	if responseID == "" {
		t.Error("X-Request-ID response header should be set")
	}
}

func TestRequestIDMiddleware_PreservesExisting(t *testing.T) {
	e := echo.New()

	existingID := "my-custom-request-id"
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := RequestIDMiddleware(func(c echo.Context) error {
		requestID := c.Get(RequestIDKey).(string)
		if requestID != existingID {
			t.Errorf("request_id = %q, want %q", requestID, existingID)
		}
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("handler error = %v", err)
	}

	responseID := rec.Header().Get(RequestIDHeader)
	if responseID != existingID {
		t.Errorf("X-Request-ID header = %q, want %q", responseID, existingID)
	}
}

func TestRequestIDMiddleware_InResponseHeader(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var capturedID string
	handler := RequestIDMiddleware(func(c echo.Context) error {
		capturedID = c.Get(RequestIDKey).(string)
		return c.NoContent(http.StatusOK)
	})

	_ = handler(c)

	responseID := rec.Header().Get(RequestIDHeader)
	if responseID != capturedID {
		t.Errorf("response header X-Request-ID = %q, want %q", responseID, capturedID)
	}
}
