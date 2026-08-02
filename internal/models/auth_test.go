package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthToken_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired",
			expiresAt: time.Now().Add(-time.Minute),
			expected:  true,
		},
		{
			name:      "not expired",
			expiresAt: time.Now().Add(time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &AuthToken{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.expected, token.IsExpired())
		})
	}
}

func TestUserSession_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "expired",
			expiresAt: time.Now().Add(-time.Second),
			expected:  true,
		},
		{
			name:      "not expired",
			expiresAt: time.Now().Add(time.Hour),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &UserSession{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.expected, session.IsExpired())
		})
	}
}

func TestUserSession_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		session  *UserSession
		expected bool
	}{
		{
			name: "valid session",
			session: &UserSession{
				UserID:    "user-1",
				Token:     "tok",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: true,
		},
		{
			name: "missing user id",
			session: &UserSession{
				UserID:    "",
				Token:     "tok",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "missing token",
			session: &UserSession{
				UserID:    "user-1",
				Token:     "",
				ExpiresAt: time.Now().Add(time.Hour),
			},
			expected: false,
		},
		{
			name: "expired",
			session: &UserSession{
				UserID:    "user-1",
				Token:     "tok",
				ExpiresAt: time.Now().Add(-time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.session.IsValid())
		})
	}
}
