package common

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{2,31}$`)

func NormalizeUsername(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ValidateUsername(raw string) *APIError {
	value := strings.TrimSpace(raw)
	if value == "" || !usernamePattern.MatchString(value) {
		return NewAPIError("INVALID_USERNAME",
			"Username must be 3-32 characters and may contain letters, digits, dots, hyphens and underscores")
	}
	return nil
}

// ValidatePassword mirrors the reference implementation: 8-20 characters with
// at least one uppercase letter, one lowercase letter and one digit.
func ValidatePassword(password string) *APIError {
	if len(password) < 8 || len(password) > 20 {
		return NewAPIError("INVALID_PASSWORD", "Password must be between 8 and 20 characters")
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
	if !hasUpper || !hasLower || !hasDigit {
		return NewAPIError("INVALID_PASSWORD",
			"Password must contain at least one uppercase letter, one lowercase letter and one digit")
	}
	return nil
}

func NormalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func ValidateEmail(raw string) *APIError {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > 254 {
		return NewAPIError("INVALID_EMAIL", "Please enter a valid email address")
	}
	parsed, err := mail.ParseAddress(value)
	if err != nil || !strings.EqualFold(parsed.Address, value) || !strings.Contains(value, "@") {
		return NewAPIError("INVALID_EMAIL", "Please enter a valid email address")
	}
	parts := strings.Split(value, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return NewAPIError("INVALID_EMAIL", "Please enter a valid email address")
	}
	return nil
}
