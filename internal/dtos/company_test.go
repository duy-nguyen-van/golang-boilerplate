package dtos

import (
	"testing"
	"time"

	"golang-boilerplate/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompanyResponse(t *testing.T) {
	now := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	company := &models.Company{
		BaseModel: models.BaseModel{
			ID:        "company-1",
			CreatedAt: now,
			UpdatedAt: now.Add(time.Hour),
		},
		Name:       "Acme",
		KeycloakID: "kc-1",
	}

	resp := NewCompanyResponse(company)
	require.NotNil(t, resp)
	assert.Equal(t, "company-1", resp.ID)
	assert.Equal(t, "Acme", resp.Name)
	assert.Equal(t, "kc-1", resp.KeycloakID)
	assert.Equal(t, now, resp.CreatedAt)
	assert.Equal(t, now.Add(time.Hour), resp.UpdatedAt)
}
