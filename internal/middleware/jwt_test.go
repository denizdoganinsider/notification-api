package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func init() {
	InitJWT("test-secret-key")
}

func TestGenerateToken_Valid(t *testing.T) {
	tokenString, err := GenerateToken(1, "user")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if tokenString == "" {
		t.Fatal("GenerateToken() returned empty token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return JwtSecret, nil
	})
	if err != nil {
		t.Fatalf("jwt.Parse() error = %v", err)
	}

	if !token.Valid {
		t.Fatal("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to parse claims")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok || int64(userID) != 1 {
		t.Errorf("user_id = %v, want 1", claims["user_id"])
	}

	role, ok := claims["role"].(string)
	if !ok || role != "user" {
		t.Errorf("role = %v, want 'user'", claims["role"])
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	e := echo.New()

	tokenString, _ := GenerateToken(42, "user")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := JWTMiddleware(func(c echo.Context) error {
		userID := c.Get("user_id").(int64)
		role := c.Get("role").(string)

		if userID != 42 {
			t.Errorf("user_id = %d, want 42", userID)
		}
		if role != "user" {
			t.Errorf("role = %q, want %q", role, "user")
		}

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

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := JWTMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	_ = handler(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestJWTMiddleware_InvalidFormat(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic some-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := JWTMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	_ = handler(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	e := echo.New()

	claims := jwt.MapClaims{
		"user_id": float64(1),
		"role":    "user",
		"exp":     time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(JwtSecret)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := JWTMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	_ = handler(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestJWTMiddleware_MalformedToken(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := JWTMiddleware(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	_ = handler(c)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
