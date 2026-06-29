package validator_test

import (
	"testing"

	"github.com/nepeta70/ride-hailing/services/user/internal/core/validator"
	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	// Table-Driven Test definition using anonymous structs
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "Valid simple lowercase email",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "Valid email with uppercase characters",
			email:    "Beth_Murphy48@gmail.com",
			expected: true,
		},
		{
			name:     "Valid email with subdomains and allowed symbols",
			email:    "user.name+tag@sub.domain.com",
			expected: true,
		},
		{
			name:     "Valid email with modern long TLD",
			email:    "developer@backend.technology",
			expected: true,
		},
		{
			name:     "Invalid email missing domain",
			email:    "plainaddress",
			expected: false,
		},
		{
			name:     "Invalid email missing @ symbol",
			email:    "missingat.com",
			expected: false,
		},
		{
			name:     "Invalid email with spaces",
			email:    "white space@domain.com",
			expected: false,
		},
		{
			name:     "Invalid email exceeding maximum length configuration",
			email:    string(make([]byte, 321)) + "@domain.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := validator.IsValidEmail(tt.email)
			assert.Equal(t, tt.expected, actual, "Result mismatch for email: %s", tt.email)
		})
	}
}
