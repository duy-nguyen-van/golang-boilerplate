package dtos

import (
	"golang-boilerplate/internal/models"
	"time"
)

// CompanyResponse represents a company response DTO
type CompanyResponse struct {
	ID         string    `json:"id" example:"123"`
	Name       string    `json:"name" example:"John Doe"`
	KeycloakID string    `json:"keycloak_id" example:"123"`
	CreatedAt  time.Time `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt  time.Time `json:"updated_at" example:"2021-01-01T00:00:00Z"`
}

// CompanyPageableRequest represents the request structure for a company
type CompanyPageableRequest struct {
	PageableRequest
	StartDate *time.Time `json:"start_date" example:"2025-02-26 08:36:23.886089+00"`
	EndDate   *time.Time `json:"end_date" example:"2025-02-26 08:36:23.886089+00"`
	Q         string     `json:"q" example:"A"`
	Sort      []string   `json:"sort" example:"[-created_at,name]" enums:"created_at,-created_at,name,-name"`
}

// CompanyRequest represents a company request DTO
type CompanyRequest struct {
	Name       string `json:"name,omitempty" example:"John Doe" validate:"omitempty,min=2,max=100"`
	KeycloakID string `json:"keycloak_id,omitempty" example:"123" validate:"omitempty,min=2,max=100"`
}

// CreateCompanyRequest represents the request structure for a company
type CreateCompanyRequest struct {
	CompanyRequest
}

// UpdateCompanyRequest represents the request structure for a company
type UpdateCompanyRequest struct {
	CompanyRequest
	ID string `json:"id,omitempty" example:"123" validate:"omitempty,min=2,max=100"`
}

func NewCompanyResponse(company *models.Company) *CompanyResponse {
	return &CompanyResponse{
		ID:         company.ID,
		Name:       company.Name,
		KeycloakID: company.KeycloakID,
		CreatedAt:  company.CreatedAt,
		UpdatedAt:  company.UpdatedAt,
	}
}
