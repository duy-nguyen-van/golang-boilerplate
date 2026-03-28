package models

import (
	"fmt"
	"time"

	_ "ariga.io/atlas-provider-gorm/gormschema" // required for Atlas GORM schema loading)
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is a base struct that includes common fields for all database models.
// All models should embed this struct to inherit these fields.
type BaseModel struct {
	ID        string         `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamptz;not null;default:now()"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamptz;not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamptz;index"`
}

func NewBaseModel() BaseModel {
	return BaseModel{
		ID: uuid.Must(uuid.NewV7()).String(),
	}
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.Must(uuid.NewV7()).String()
	} else {
		if _, err := uuid.Parse(b.ID); err != nil {
			return fmt.Errorf("invalid uuid: %w", err)
		}
	}

	return nil
}
