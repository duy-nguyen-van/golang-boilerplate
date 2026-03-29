package services

import (
	"context"
	"testing"

	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCompanyRepository is already defined in user_test.go but we need it here too
type MockCompanyRepositoryForCompanyService struct {
	mock.Mock
}

func (m *MockCompanyRepositoryForCompanyService) Create(company *models.Company) (*models.Company, error) {
	args := m.Called(company)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepositoryForCompanyService) GetOneByID(id string) (*models.Company, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepositoryForCompanyService) Update(company *models.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *MockCompanyRepositoryForCompanyService) Delete(company *models.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *MockCompanyRepositoryForCompanyService) Get(pr *dtos.CompanyPageableRequest, preloads ...string) (*dtos.DataResponse[models.Company], error) {
	args := m.Called(pr, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.DataResponse[models.Company]), args.Error(1)
}

func TestCompanyService_Create(t *testing.T) {
	tests := []struct {
		name          string
		req           *dtos.CreateCompanyRequest
		setupMocks    func(*MockCompanyRepositoryForCompanyService, *MockCache)
		expectedError bool
		errorType     string
	}{
		{
			name: "success - create company",
			req: &dtos.CreateCompanyRequest{
				CompanyRequest: dtos.CompanyRequest{
					Name:       "Acme Corp",
					KeycloakID: "keycloak-123",
				},
			},
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				createdCompany := &models.Company{
					BaseModel: models.BaseModel{
						ID: uuid.New().String(),
					},
					Name:       "Acme Corp",
					KeycloakID: "keycloak-123",
				}

				companyRepo.On("Create", mock.AnythingOfType("*models.Company")).Return(createdCompany, nil)
			},
			expectedError: false,
		},
		{
			name: "error - database error on create",
			req: &dtos.CreateCompanyRequest{
				CompanyRequest: dtos.CompanyRequest{
					Name:       "Acme Corp",
					KeycloakID: "keycloak-123",
				},
			},
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				companyRepo.On("Create", mock.AnythingOfType("*models.Company")).Return(nil, errors.DatabaseError("Failed to create company", nil))
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompanyRepo := new(MockCompanyRepositoryForCompanyService)
			mockCache := new(MockCache)

			if tt.setupMocks != nil {
				tt.setupMocks(mockCompanyRepo, mockCache)
			}

			service := &companyService{
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			result, err := service.Create(ctx, tt.req)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					appErr, ok := err.(*errors.AppError)
					require.True(t, ok, "Expected AppError")
					switch tt.errorType {
					case "DatabaseError":
						assert.Equal(t, errors.ErrorTypeDatabase, appErr.Type)
					}
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.req.Name, result.Name)
				assert.Equal(t, tt.req.KeycloakID, result.KeycloakID)
			}

			mockCompanyRepo.AssertExpectations(t)
		})
	}
}

func TestCompanyService_GetOneByID(t *testing.T) {
	tests := []struct {
		name          string
		companyID     string
		setupMocks    func(*MockCompanyRepositoryForCompanyService, *MockCache)
		expectedError bool
	}{
		{
			name:      "success",
			companyID: uuid.New().String(),
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				company := &models.Company{
					BaseModel: models.BaseModel{
						ID: uuid.New().String(),
					},
					Name:       "Acme Corp",
					KeycloakID: "keycloak-123",
				}
				companyRepo.On("GetOneByID", mock.AnythingOfType("string")).Return(company, nil)
			},
			expectedError: false,
		},
		{
			name:      "error - company not found",
			companyID: "non-existent",
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				companyRepo.On("GetOneByID", "non-existent").Return(nil, errors.NotFoundError("Company", nil))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompanyRepo := new(MockCompanyRepositoryForCompanyService)
			mockCache := new(MockCache)

			if tt.setupMocks != nil {
				tt.setupMocks(mockCompanyRepo, mockCache)
			}

			service := &companyService{
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			result, err := service.GetOneByID(ctx, tt.companyID)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.ID)
			}

			mockCompanyRepo.AssertExpectations(t)
		})
	}
}

func TestCompanyService_Update(t *testing.T) {
	tests := []struct {
		name          string
		companyID     string
		req           *dtos.UpdateCompanyRequest
		setupMocks    func(*MockCompanyRepositoryForCompanyService, *MockCache)
		expectedError bool
		errorType     string
	}{
		{
			name:      "success - update company name",
			companyID: uuid.New().String(),
			req: &dtos.UpdateCompanyRequest{
				CompanyRequest: dtos.CompanyRequest{
					Name: "Updated Company Name",
				},
			},
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				companyID := uuid.New()
				company := &models.Company{
					BaseModel: models.BaseModel{
						ID: companyID.String(),
					},
					Name:       "Old Name",
					KeycloakID: "keycloak-123",
				}

				companyRepo.On("GetOneByID", mock.AnythingOfType("string")).Return(company, nil)
				companyRepo.On("Update", mock.AnythingOfType("*models.Company")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "error - company not found",
			companyID: "non-existent",
			req: &dtos.UpdateCompanyRequest{
				CompanyRequest: dtos.CompanyRequest{
					Name: "Updated Name",
				},
			},
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				companyRepo.On("GetOneByID", "non-existent").Return(nil, errors.NotFoundError("Company", nil))
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:      "error - database error on update",
			companyID: uuid.New().String(),
			req: &dtos.UpdateCompanyRequest{
				CompanyRequest: dtos.CompanyRequest{
					Name: "Updated Name",
				},
			},
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				company := &models.Company{
					BaseModel: models.BaseModel{
						ID: uuid.New().String(),
					},
					Name: "Old Name",
				}
				companyRepo.On("GetOneByID", mock.AnythingOfType("string")).Return(company, nil)
				companyRepo.On("Update", mock.AnythingOfType("*models.Company")).Return(errors.DatabaseError("Failed to update company", nil))
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompanyRepo := new(MockCompanyRepositoryForCompanyService)
			mockCache := new(MockCache)

			if tt.setupMocks != nil {
				tt.setupMocks(mockCompanyRepo, mockCache)
			}

			service := &companyService{
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			result, err := service.Update(ctx, tt.companyID, tt.req)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					appErr, ok := err.(*errors.AppError)
					require.True(t, ok, "Expected AppError")
					switch tt.errorType {
					case "NotFoundError":
						assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
					case "DatabaseError":
						assert.Equal(t, errors.ErrorTypeDatabase, appErr.Type)
					}
				}
				if tt.errorType != "NotFoundError" {
					assert.Nil(t, result)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.req.Name != "" {
					assert.Equal(t, tt.req.Name, result.Name)
				}
			}

			mockCompanyRepo.AssertExpectations(t)
		})
	}
}

func TestCompanyService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		companyID     string
		setupMocks    func(*MockCompanyRepositoryForCompanyService, *MockCache)
		expectedError bool
		errorType     string
	}{
		{
			name:      "success",
			companyID: uuid.New().String(),
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				company := &models.Company{
					BaseModel: models.BaseModel{
						ID: uuid.New().String(),
					},
					Name: "Acme Corp",
				}
				companyRepo.On("GetOneByID", mock.AnythingOfType("string")).Return(company, nil)
				companyRepo.On("Delete", mock.AnythingOfType("*models.Company")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:      "error - company not found",
			companyID: "non-existent",
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				companyRepo.On("GetOneByID", "non-existent").Return(nil, errors.NotFoundError("Company", nil))
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:      "error - database error on delete",
			companyID: uuid.New().String(),
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				company := &models.Company{
					BaseModel: models.BaseModel{
						ID: uuid.New().String(),
					},
					Name: "Acme Corp",
				}
				companyRepo.On("GetOneByID", mock.AnythingOfType("string")).Return(company, nil)
				companyRepo.On("Delete", mock.AnythingOfType("*models.Company")).Return(errors.DatabaseError("Failed to delete company", nil))
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompanyRepo := new(MockCompanyRepositoryForCompanyService)
			mockCache := new(MockCache)

			if tt.setupMocks != nil {
				tt.setupMocks(mockCompanyRepo, mockCache)
			}

			service := &companyService{
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			err := service.Delete(ctx, tt.companyID)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					appErr, ok := err.(*errors.AppError)
					require.True(t, ok, "Expected AppError")
					switch tt.errorType {
					case "NotFoundError":
						assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
					case "DatabaseError":
						assert.Equal(t, errors.ErrorTypeDatabase, appErr.Type)
					}
				}
			} else {
				require.NoError(t, err)
			}

			mockCompanyRepo.AssertExpectations(t)
		})
	}
}

func TestCompanyService_List(t *testing.T) {
	tests := []struct {
		name          string
		pageableReq   *dtos.CompanyPageableRequest
		setupMocks    func(*MockCompanyRepositoryForCompanyService, *MockCache)
		expectedError bool
	}{
		{
			name: "success",
			pageableReq: &dtos.CompanyPageableRequest{
				PageableRequest: dtos.PageableRequest{
					Page:     1,
					PageSize: 10,
				},
			},
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				companies := &dtos.DataResponse[models.Company]{
					Data: []models.Company{
						{
							BaseModel: models.BaseModel{ID: uuid.New().String()},
							Name:      "Company 1",
						},
						{
							BaseModel: models.BaseModel{ID: uuid.New().String()},
							Name:      "Company 2",
						},
					},
				}
				companyRepo.On("Get", mock.AnythingOfType("*dtos.CompanyPageableRequest"), mock.AnythingOfType("[]string")).Return(companies, nil)
			},
			expectedError: false,
		},
		{
			name: "error - database error",
			pageableReq: &dtos.CompanyPageableRequest{
				PageableRequest: dtos.PageableRequest{
					Page:     1,
					PageSize: 10,
				},
			},
			setupMocks: func(companyRepo *MockCompanyRepositoryForCompanyService, cache *MockCache) {
				companyRepo.On("Get", mock.AnythingOfType("*dtos.CompanyPageableRequest"), mock.AnythingOfType("[]string")).Return(nil, errors.DatabaseError("Failed to get companies", nil))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCompanyRepo := new(MockCompanyRepositoryForCompanyService)
			mockCache := new(MockCache)

			if tt.setupMocks != nil {
				tt.setupMocks(mockCompanyRepo, mockCache)
			}

			service := &companyService{
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			result, err := service.List(ctx, tt.pageableReq)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Data)
			}

			mockCompanyRepo.AssertExpectations(t)
		})
	}
}
