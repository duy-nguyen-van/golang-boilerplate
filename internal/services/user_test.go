package services

import (
	"context"
	"testing"
	"time"

	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) (*models.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetOneByID(id string, preloads ...string) (*models.User, error) {
	// Convert variadic preloads to slice for mock matching
	preloadsSlice := []string{}
	if len(preloads) > 0 {
		preloadsSlice = preloads
	}
	args := m.Called(id, preloadsSlice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Get(pr *dtos.UserPageableRequest, preloads ...string) (*dtos.DataResponse[models.User], error) {
	args := m.Called(pr, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.DataResponse[models.User]), args.Error(1)
}

// MockCompanyRepository is a mock implementation of CompanyRepository
type MockCompanyRepository struct {
	mock.Mock
}

func (m *MockCompanyRepository) Create(company *models.Company) (*models.Company, error) {
	args := m.Called(company)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepository) GetOneByID(id string) (*models.Company, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *MockCompanyRepository) Update(company *models.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *MockCompanyRepository) Delete(company *models.Company) error {
	args := m.Called(company)
	return args.Error(0)
}

func (m *MockCompanyRepository) Get(pr *dtos.CompanyPageableRequest, preloads ...string) (*dtos.DataResponse[models.Company], error) {
	args := m.Called(pr, preloads)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.DataResponse[models.Company]), args.Error(1)
}

// MockCache is a mock implementation of cache.Cache
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name          string
		req           *dtos.CreateUserRequest
		setupMocks    func(*MockUserRepository, *MockCompanyRepository, *MockCache)
		expectedError bool
		errorType     string
	}{
		{
			name: "success - create user with companies",
			req: &dtos.CreateUserRequest{
				UserRequest: dtos.UserRequest{
					FirstName:  "John",
					LastName:   "Doe",
					Email:      "john.doe@example.com",
					KeycloakID: "keycloak-123",
					Companies: []dtos.UpdateCompanyRequest{
						{ID: "company-1"},
						{ID: "company-2"},
					},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository, cache *MockCache) {
				companyID1 := uuid.New()
				companyID2 := uuid.New()
				company1 := &models.Company{
					BaseModel: models.BaseModel{
						ID: companyID1.String(),
					},
					Name: "Company 1",
				}
				company2 := &models.Company{
					BaseModel: models.BaseModel{
						ID: companyID2.String(),
					},
					Name: "Company 2",
				}

				companyRepo.On("GetOneByID", "company-1").Return(company1, nil)
				companyRepo.On("GetOneByID", "company-2").Return(company2, nil)

				createdUser := &models.User{
					BaseModel: models.BaseModel{
						ID: uuid.New().String(),
					},
					FirstName:  "John",
					LastName:   "Doe",
					Email:      "john.doe@example.com",
					KeycloakID: "keycloak-123",
					Companies:  []models.Company{*company1, *company2},
				}

				userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(createdUser, nil)
			},
			expectedError: false,
		},
		{
			name: "success - create user without companies",
			req: &dtos.CreateUserRequest{
				UserRequest: dtos.UserRequest{
					FirstName:  "Jane",
					LastName:   "Smith",
					Email:      "jane.smith@example.com",
					KeycloakID: "keycloak-456",
					Companies:  []dtos.UpdateCompanyRequest{},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository, cache *MockCache) {
				createdUser := &models.User{
					BaseModel: models.BaseModel{
						ID: uuid.New().String(),
					},
					FirstName:  "Jane",
					LastName:   "Smith",
					Email:      "jane.smith@example.com",
					KeycloakID: "keycloak-456",
				}

				userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(createdUser, nil)
			},
			expectedError: false,
		},
		{
			name: "error - company not found",
			req: &dtos.CreateUserRequest{
				UserRequest: dtos.UserRequest{
					FirstName:  "John",
					LastName:   "Doe",
					Email:      "john.doe@example.com",
					KeycloakID: "keycloak-123",
					Companies: []dtos.UpdateCompanyRequest{
						{ID: "non-existent"},
					},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository, cache *MockCache) {
				companyRepo.On("GetOneByID", "non-existent").Return(nil, errors.NotFoundError("Company", nil))
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name: "error - database error on create",
			req: &dtos.CreateUserRequest{
				UserRequest: dtos.UserRequest{
					FirstName:  "John",
					LastName:   "Doe",
					Email:      "john.doe@example.com",
					KeycloakID: "keycloak-123",
					Companies:  []dtos.UpdateCompanyRequest{},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository, cache *MockCache) {
				userRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil, errors.DatabaseError("Failed to create user", nil))
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := new(MockUserRepository)
			mockCompanyRepo := new(MockCompanyRepository)
			mockCache := new(MockCache)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockCompanyRepo, mockCache)
			}

			// Create service with mocks
			service := &userService{
				userRepo:    mockUserRepo,
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			// Execute
			ctx := context.Background()
			result, err := service.Create(ctx, tt.req)

			// Assert
			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					// Check error type if needed
					switch tt.errorType {
					case "NotFoundError":
						_, ok := err.(*errors.AppError)
						assert.True(t, ok, "Expected NotFoundError")
					case "DatabaseError":
						_, ok := err.(*errors.AppError)
						assert.True(t, ok, "Expected DatabaseError")
					}
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.req.FirstName, result.FirstName)
				assert.Equal(t, tt.req.LastName, result.LastName)
				assert.Equal(t, tt.req.Email, result.Email)
				assert.Equal(t, tt.req.KeycloakID, result.KeycloakID)
			}

			// Verify all expectations were met
			mockUserRepo.AssertExpectations(t)
			mockCompanyRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetOneByID(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*MockUserRepository, *MockCompanyRepository, *MockCache)
		expectedError bool
	}{
		{
			name:   "success",
			userID: uuid.New().String(),
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository, cache *MockCache) {
				userID := uuid.New()
				user := &models.User{
					BaseModel: models.BaseModel{
						ID: userID.String(),
					},
					FirstName: "John",
					LastName:  "Doe",
					Email:     "john.doe@example.com",
				}
				userRepo.On("GetOneByID", mock.AnythingOfType("string"), mock.AnythingOfType("[]string")).Return(user, nil)
			},
			expectedError: false,
		},
		{
			name:   "error - user not found",
			userID: "non-existent",
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository, cache *MockCache) {
				userRepo.On("GetOneByID", "non-existent", mock.AnythingOfType("[]string")).Return(nil, errors.NotFoundError("User", nil))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := new(MockUserRepository)
			mockCompanyRepo := new(MockCompanyRepository)
			mockCache := new(MockCache)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockCompanyRepo, mockCache)
			}

			// Create service with mocks
			service := &userService{
				userRepo:    mockUserRepo,
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			// Execute
			ctx := context.Background()
			result, err := service.GetOneByID(ctx, tt.userID)

			// Assert
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.ID)
			}

			mockUserRepo.AssertExpectations(t)
		})
	}
}
