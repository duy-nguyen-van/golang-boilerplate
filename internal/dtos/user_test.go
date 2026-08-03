package dtos

import (
	"testing"
	"time"

	"golang-boilerplate/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserResponse(t *testing.T) {
	now := time.Date(2025, 2, 1, 12, 0, 0, 0, time.UTC)

	t.Run("without companies", func(t *testing.T) {
		user := &models.User{
			BaseModel: models.BaseModel{
				ID:        "user-1",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Email:     "a@example.com",
			FirstName: "Ada",
			LastName:  "Lovelace",
		}

		resp := NewUserResponse(user)
		require.NotNil(t, resp)
		assert.Equal(t, "user-1", resp.ID)
		assert.Equal(t, "a@example.com", resp.Email)
		assert.Equal(t, "Ada", resp.FirstName)
		assert.Equal(t, "Lovelace", resp.LastName)
		assert.Equal(t, now, resp.CreatedAt)
		assert.Equal(t, now, resp.UpdatedAt)
		assert.Empty(t, resp.Companies)
	})

	t.Run("with companies", func(t *testing.T) {
		user := &models.User{
			BaseModel: models.BaseModel{ID: "user-2", CreatedAt: now, UpdatedAt: now},
			Email:     "b@example.com",
			FirstName: "Grace",
			LastName:  "Hopper",
			Companies: []models.Company{
				{
					BaseModel:  models.BaseModel{ID: "c1", CreatedAt: now, UpdatedAt: now},
					Name:       "Navy",
					KeycloakID: "kc-n",
				},
				{
					BaseModel:  models.BaseModel{ID: "c2", CreatedAt: now, UpdatedAt: now},
					Name:       "COBOL Inc",
					KeycloakID: "kc-c",
				},
			},
		}

		resp := NewUserResponse(user)
		require.NotNil(t, resp)
		require.Len(t, resp.Companies, 2)
		assert.Equal(t, "c1", resp.Companies[0].ID)
		assert.Equal(t, "Navy", resp.Companies[0].Name)
		assert.Equal(t, "c2", resp.Companies[1].ID)
		assert.Equal(t, "COBOL Inc", resp.Companies[1].Name)
	})
}
