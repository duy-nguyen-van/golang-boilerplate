package dtos

import (
	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/models"
	"time"
)

// UserResponse represents a user response DTO
type UserResponse struct {
	ID        string            `json:"id" example:"123"`
	Email     string            `json:"email" example:"john.doe@example.com"`
	FirstName string            `json:"first_name" example:"John"`
	LastName  string            `json:"last_name" example:"Doe"`
	CreatedAt time.Time         `json:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt time.Time         `json:"updated_at" example:"2021-01-01T00:00:00Z"`
	Companies []CompanyResponse `json:"companies"`
}

// UserPageableRequest represents the request structure for a user
type UserPageableRequest struct {
	PageableRequest
	StartDate *time.Time `json:"start_date" example:"2025-02-26 08:36:23.886089+00"`
	EndDate   *time.Time `json:"end_date" example:"2025-02-26 08:36:23.886089+00"`
	Q         string     `json:"q" example:"A"`
	Sort      []string   `json:"sort" example:"[-created_at,name]" enums:"created_at,-created_at,name,-name"`
}

// UserRequest represents a user request DTO
type UserRequest struct {
	Email      string                 `json:"email,omitempty" example:"john.doe@example.com" validate:"omitempty,email"`
	FirstName  string                 `json:"first_name,omitempty" example:"John" validate:"omitempty,min=2,max=100"`
	LastName   string                 `json:"last_name,omitempty" example:"Doe" validate:"omitempty,min=2,max=100"`
	KeycloakID string                 `json:"keycloak_id,omitempty" example:"123" validate:"omitempty,min=2,max=100"`
	Companies  []UpdateCompanyRequest `json:"companies,omitempty"`
}

// CreateCompanyRequest represents the request structure for a company
type CreateUserRequest struct {
	UserRequest
}

// UpdateCompanyRequest represents the request structure for a company
type UpdateUserRequest struct {
	UserRequest
	Status constants.UserStatus `json:"status" example:"active" enums:"active,inactive"`
}

func NewUserResponse(user *models.User) *UserResponse {
	result := &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	companies := make([]CompanyResponse, len(user.Companies))
	for i, company := range user.Companies {
		companies[i] = *NewCompanyResponse(&company)
	}
	result.Companies = companies

	return result
}
