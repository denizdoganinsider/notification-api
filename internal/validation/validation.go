package validation

import (
	"errors"
	"net/mail"
	"unicode"
)

func ValidateEmail(email string) error {
	if len(email) > 255 {
		return errors.New("email must be at most 255 characters")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email format")
	}

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return errors.New("password must be at most 128 characters")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}

	return nil
}

func ValidateNotificationTitle(title string) error {
	if title == "" {
		return errors.New("title is required")
	}

	if len(title) > 255 {
		return errors.New("title must be at most 255 characters")
	}

	return nil
}

func ValidateNotificationMessage(message string) error {
	if message == "" {
		return errors.New("message is required")
	}

	if len(message) > 65535 {
		return errors.New("message must be at most 65535 characters")
	}

	return nil
}
