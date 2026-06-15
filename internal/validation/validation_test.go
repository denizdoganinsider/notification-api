package validation

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr string
	}{
		{"valid email", "user@example.com", ""},
		{"valid email with name", "User Name <user@example.com>", ""},
		{"empty email", "", "invalid email format"},
		{"missing @", "userexample.com", "invalid email format"},
		{"missing domain", "user@", "invalid email format"},
		{"too long", "a" + strings.Repeat("b", 255) + "@example.com", "email must be at most 255 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateEmail(%q) = %v, want nil", tt.email, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateEmail(%q) = nil, want error containing %q", tt.email, tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("ValidateEmail(%q) = %q, want %q", tt.email, err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{"valid password", "Password1", ""},
		{"valid complex", "MyP@ssw0rd!", ""},
		{"too short", "Pass1", "password must be at least 8 characters"},
		{"too long", strings.Repeat("A", 129), "password must be at most 128 characters"},
		{"no uppercase", "password1", "password must contain at least one uppercase letter"},
		{"no lowercase", "PASSWORD1", "password must contain at least one lowercase letter"},
		{"no digit", "Passwordd", "password must contain at least one digit"},
		{"empty", "", "password must be at least 8 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidatePassword(%q) = %v, want nil", tt.password, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidatePassword(%q) = nil, want error containing %q", tt.password, tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("ValidatePassword(%q) = %q, want %q", tt.password, err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestValidateNotificationTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr string
	}{
		{"valid title", "Hello World", ""},
		{"empty title", "", "title is required"},
		{"too long title", strings.Repeat("a", 256), "title must be at most 255 characters"},
		{"max length title", strings.Repeat("a", 255), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotificationTitle(tt.title)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateNotificationTitle(%q) = %v, want nil", tt.title, err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateNotificationTitle(%q) = nil, want error %q", tt.title, tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("ValidateNotificationTitle(%q) = %q, want %q", tt.title, err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestValidateNotificationMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr string
	}{
		{"valid message", "This is a notification", ""},
		{"empty message", "", "message is required"},
		{"too long message", strings.Repeat("a", 65536), "message must be at most 65535 characters"},
		{"max length message", strings.Repeat("a", 65535), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotificationMessage(tt.message)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("ValidateNotificationMessage() = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateNotificationMessage() = nil, want error %q", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("ValidateNotificationMessage() = %q, want %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}
