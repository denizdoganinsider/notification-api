package service

import (
	"errors"
	"notification-api/internal/domain"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users  map[string]*domain.User
	nextID int64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[string]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(user *domain.User) error {
	if _, exists := m.users[user.Email]; exists {
		return errors.New("duplicate email")
	}
	user.ID = m.nextID
	user.CreatedAt = time.Now()
	m.nextID++
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepo) GetByID(id int64) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepo) GetAll() ([]domain.User, error) {
	var users []domain.User
	for _, user := range m.users {
		users = append(users, *user)
	}
	return users, nil
}

func TestUserService_Register_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	user, err := svc.Register("test@example.com", "Password1")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("user.Email = %q, want %q", user.Email, "test@example.com")
	}

	if user.Role != "user" {
		t.Errorf("user.Role = %q, want %q", user.Role, "user")
	}

	if user.ID == 0 {
		t.Error("user.ID should not be 0")
	}
}

func TestUserService_Register_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.Register("test@example.com", "Password1")
	if err != nil {
		t.Fatalf("first Register() error = %v", err)
	}

	_, err = svc.Register("test@example.com", "Password2")
	if err == nil {
		t.Fatal("second Register() should return error for duplicate email")
	}

	if err.Error() != "email already exists" {
		t.Errorf("error = %q, want %q", err.Error(), "email already exists")
	}
}

func TestUserService_Register_InvalidEmail(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.Register("not-an-email", "Password1")
	if err == nil {
		t.Fatal("Register() should return error for invalid email")
	}

	if err.Error() != "invalid email format" {
		t.Errorf("error = %q, want %q", err.Error(), "invalid email format")
	}
}

func TestUserService_Register_WeakPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{"too short", "Pass1", "password must be at least 8 characters"},
		{"no uppercase", "password1", "password must contain at least one uppercase letter"},
		{"no lowercase", "PASSWORD1", "password must contain at least one lowercase letter"},
		{"no digit", "Passwordd", "password must contain at least one digit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register("valid@example.com", tt.password)
			if err == nil {
				t.Fatal("Register() should return error for weak password")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestUserService_Login_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.Register("test@example.com", "Password1")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	user, err := svc.Login("test@example.com", "Password1")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("user.Email = %q, want %q", user.Email, "test@example.com")
	}
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	hash, _ := bcrypt.GenerateFromPassword([]byte("Password1"), bcrypt.DefaultCost)
	repo.users["test@example.com"] = &domain.User{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hash),
		Role:         "user",
	}

	_, err := svc.Login("test@example.com", "WrongPassword1")
	if err == nil {
		t.Fatal("Login() should return error for wrong password")
	}

	if err.Error() != "invalid email or password" {
		t.Errorf("error = %q, want %q", err.Error(), "invalid email or password")
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.Login("nonexistent@example.com", "Password1")
	if err == nil {
		t.Fatal("Login() should return error for nonexistent user")
	}

	if err.Error() != "invalid email or password" {
		t.Errorf("error = %q, want %q", err.Error(), "invalid email or password")
	}
}
