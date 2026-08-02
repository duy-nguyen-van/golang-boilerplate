package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseModel(t *testing.T) {
	m := NewBaseModel()
	require.NotEmpty(t, m.ID)

	parsed, err := uuid.Parse(m.ID)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), parsed.Version())
	assert.True(t, m.CreatedAt.IsZero())
	assert.True(t, m.UpdatedAt.IsZero())
}

func TestBaseModel_BeforeCreate(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectError bool
		expectIDSet bool
	}{
		{
			name:        "empty id generates uuid",
			id:          "",
			expectError: false,
			expectIDSet: true,
		},
		{
			name:        "valid uuid kept",
			id:          "01900000-0000-7000-8000-000000000001",
			expectError: false,
			expectIDSet: true,
		},
		{
			name:        "invalid uuid returns error",
			id:          "not-a-uuid",
			expectError: true,
			expectIDSet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BaseModel{ID: tt.id}
			err := m.BeforeCreate(nil)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid uuid")
				assert.Equal(t, tt.id, m.ID)
				return
			}

			require.NoError(t, err)
			if tt.expectIDSet {
				require.NotEmpty(t, m.ID)
				_, parseErr := uuid.Parse(m.ID)
				require.NoError(t, parseErr)
			}
			if tt.id != "" {
				assert.Equal(t, tt.id, m.ID)
			}
		})
	}
}
