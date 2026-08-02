package request

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCtxKey_String(t *testing.T) {
	k := ctxKey{name: "language_code"}
	assert.Equal(t, "golang-boilerplate context value language_code", k.String())
}

func TestLanguageCodeContext(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() context.Context
		expectCode    string
		expectPresent bool
	}{
		{
			name: "value present",
			setup: func() context.Context {
				return NewLanguageCodeContext(context.Background(), "en")
			},
			expectCode:    "en",
			expectPresent: true,
		},
		{
			name: "missing",
			setup: func() context.Context {
				return context.Background()
			},
			expectCode:    "",
			expectPresent: false,
		},
		{
			name: "wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxKeyLanguageCode, 1)
			},
			expectCode:    "",
			expectPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, ok := LanguageCodeFromContext(tt.setup())
			assert.Equal(t, tt.expectPresent, ok)
			assert.Equal(t, tt.expectCode, code)
		})
	}
}

func TestRequestTimestampContext(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() context.Context
		expectTS      int64
		expectPresent bool
	}{
		{
			name: "value present",
			setup: func() context.Context {
				return NewRequestTimestampContext(context.Background(), 1710000000123)
			},
			expectTS:      1710000000123,
			expectPresent: true,
		},
		{
			name: "missing",
			setup: func() context.Context {
				return context.Background()
			},
			expectTS:      0,
			expectPresent: false,
		},
		{
			name: "wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxKeyRequestTimestamp, "not-int64")
			},
			expectTS:      0,
			expectPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, ok := RequestTimestampFromContext(tt.setup())
			assert.Equal(t, tt.expectPresent, ok)
			assert.Equal(t, tt.expectTS, ts)
		})
	}
}

func TestCorrelationIDContext(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() context.Context
		expectID      string
		expectPresent bool
	}{
		{
			name: "value present",
			setup: func() context.Context {
				return NewCorrelationIDContext(context.Background(), "req-corr-1")
			},
			expectID:      "req-corr-1",
			expectPresent: true,
		},
		{
			name: "missing",
			setup: func() context.Context {
				return context.Background()
			},
			expectID:      "",
			expectPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := CorrelationIDFromContext(tt.setup())
			assert.Equal(t, tt.expectPresent, ok)
			assert.Equal(t, tt.expectID, id)
		})
	}
}

func TestRequestURLContext(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() context.Context
		expectURL     string
		expectPresent bool
	}{
		{
			name: "value present",
			setup: func() context.Context {
				return NewRequestURLContext(context.Background(), "/api/v1/users")
			},
			expectURL:     "/api/v1/users",
			expectPresent: true,
		},
		{
			name: "missing",
			setup: func() context.Context {
				return context.Background()
			},
			expectURL:     "",
			expectPresent: false,
		},
		{
			name: "wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), ctxKeyRequestURL, 99)
			},
			expectURL:     "",
			expectPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, ok := RequestURLFromContext(tt.setup())
			assert.Equal(t, tt.expectPresent, ok)
			assert.Equal(t, tt.expectURL, url)
		})
	}
}

func TestContextHelpers_Compose(t *testing.T) {
	ctx := context.Background()
	ctx = NewLanguageCodeContext(ctx, "vi")
	ctx = NewRequestTimestampContext(ctx, 42)
	ctx = NewCorrelationIDContext(ctx, "cid")
	ctx = NewRequestURLContext(ctx, "/health")

	code, ok := LanguageCodeFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "vi", code)

	ts, ok := RequestTimestampFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, int64(42), ts)

	id, ok := CorrelationIDFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "cid", id)

	url, ok := RequestURLFromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "/health", url)
}
