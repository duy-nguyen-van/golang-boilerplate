package correlationid

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeader(t *testing.T) {
	assert.Equal(t, http.CanonicalHeaderKey("X-Correlation-Id"), Header)
}

func TestNewContext_FromContext(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() context.Context
		expectID      string
		expectPresent bool
	}{
		{
			name: "value present",
			setup: func() context.Context {
				return NewContext(context.Background(), "corr-123")
			},
			expectID:      "corr-123",
			expectPresent: true,
		},
		{
			name: "empty correlation id still present",
			setup: func() context.Context {
				return NewContext(context.Background(), "")
			},
			expectID:      "",
			expectPresent: true,
		},
		{
			name: "missing from background context",
			setup: func() context.Context {
				return context.Background()
			},
			expectID:      "",
			expectPresent: false,
		},
		{
			name: "wrong type in context",
			setup: func() context.Context {
				return context.WithValue(context.Background(), correlationIDCtxKey, 42)
			},
			expectID:      "",
			expectPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := FromContext(tt.setup())
			assert.Equal(t, tt.expectPresent, ok)
			assert.Equal(t, tt.expectID, id)
		})
	}
}

func TestNewContext_Overwrites(t *testing.T) {
	ctx := NewContext(context.Background(), "first")
	ctx = NewContext(ctx, "second")

	id, ok := FromContext(ctx)
	require.True(t, ok)
	assert.Equal(t, "second", id)
}
