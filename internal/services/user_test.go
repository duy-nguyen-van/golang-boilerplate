package services

import (
	"context"
	"testing"
	"time"

	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/models"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ctxWithSentryHub() context.Context {
	hub := sentry.CurrentHub().Clone()
	return sentry.SetHubOnContext(context.Background(), hub)
}

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

func TestProvideUserService(t *testing.T) {
	svc := ProvideUserService(new(MockUserRepository), new(MockCompanyRepository), new(MockCache))
	require.NotNil(t, svc)
	assert.Implements(t, (*UserService)(nil), svc)
}

func TestUserService_Create_WithSentryHub(t *testing.T) {
	t.Run("company not found reports to sentry", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		mockCache := new(MockCache)
		mockCompanyRepo.On("GetOneByID", "missing").Return(nil, assert.AnError)

		service := &userService{userRepo: mockUserRepo, companyRepo: mockCompanyRepo, cache: mockCache}
		result, err := service.Create(ctxWithSentryHub(), &dtos.CreateUserRequest{
			UserRequest: dtos.UserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Companies: []dtos.UpdateCompanyRequest{{ID: "missing"}},
			},
		})
		require.Error(t, err)
		assert.Nil(t, result)
		mockCompanyRepo.AssertExpectations(t)
	})

	t.Run("database error reports to sentry", func(t *testing.T) {
		mockUserRepo := new(MockUserRepository)
		mockCompanyRepo := new(MockCompanyRepository)
		mockCache := new(MockCache)
		mockUserRepo.On("Create", mock.AnythingOfType("*models.User")).Return(nil, assert.AnError)

		service := &userService{userRepo: mockUserRepo, companyRepo: mockCompanyRepo, cache: mockCache}
		result, err := service.Create(ctxWithSentryHub(), &dtos.CreateUserRequest{
			UserRequest: dtos.UserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
		})
		require.Error(t, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetOneByID_WithSentryHub(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockCompanyRepo := new(MockCompanyRepository)
	mockCache := new(MockCache)
	mockUserRepo.On("GetOneByID", "missing", mock.AnythingOfType("[]string")).Return(nil, assert.AnError)

	service := &userService{userRepo: mockUserRepo, companyRepo: mockCompanyRepo, cache: mockCache}
	result, err := service.GetOneByID(ctxWithSentryHub(), "missing")
	require.Error(t, err)
	assert.Nil(t, result)
	mockUserRepo.AssertExpectations(t)
}

func TestUserService_Update(t *testing.T) {
	existingCompanyID := uuid.New().String()
	newCompanyID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		req           *dtos.UpdateUserRequest
		withSentry    bool
		setupMocks    func(*MockUserRepository, *MockCompanyRepository)
		expectedError bool
		errorType     string
		assertResult  func(*testing.T, *models.User)
	}{
		{
			name:   "success - update fields and keep existing company",
			userID: userID,
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{
					FirstName:  "Updated",
					LastName:   "Name",
					Email:      "updated@example.com",
					KeycloakID: "kc-updated",
					Companies:  []dtos.UpdateCompanyRequest{{ID: existingCompanyID}},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				user := &models.User{
					BaseModel:  models.BaseModel{ID: userID},
					FirstName:  "Old",
					LastName:   "User",
					Email:      "old@example.com",
					KeycloakID: "kc-old",
					Companies: []models.Company{
						{BaseModel: models.BaseModel{ID: existingCompanyID}, Name: "Existing Co"},
					},
				}
				userRepo.On("GetOneByID", userID, []string{"Companies"}).Return(user, nil)
				userRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
					return u.FirstName == "Updated" &&
						u.LastName == "Name" &&
						u.Email == "updated@example.com" &&
						u.KeycloakID == "kc-updated" &&
						len(u.Companies) == 1 &&
						u.Companies[0].ID == existingCompanyID
				})).Return(nil)
			},
			expectedError: false,
			assertResult: func(t *testing.T, u *models.User) {
				assert.Equal(t, "Updated", u.FirstName)
				assert.Equal(t, "kc-updated", u.KeycloakID)
				require.Len(t, u.Companies, 1)
			},
		},
		{
			name:   "success - associate new company",
			userID: userID,
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{
					Companies: []dtos.UpdateCompanyRequest{{ID: newCompanyID}, {ID: ""}},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				user := &models.User{
					BaseModel: models.BaseModel{ID: userID},
					FirstName: "John",
					Companies: []models.Company{},
				}
				newCompany := &models.Company{
					BaseModel: models.BaseModel{ID: newCompanyID},
					Name:      "New Co",
				}
				userRepo.On("GetOneByID", userID, []string{"Companies"}).Return(user, nil)
				companyRepo.On("GetOneByID", newCompanyID).Return(newCompany, nil)
				userRepo.On("Update", mock.MatchedBy(func(u *models.User) bool {
					return len(u.Companies) == 1 && u.Companies[0].ID == newCompanyID
				})).Return(nil)
			},
			expectedError: false,
			assertResult: func(t *testing.T, u *models.User) {
				require.Len(t, u.Companies, 1)
				assert.Equal(t, newCompanyID, u.Companies[0].ID)
			},
		},
		{
			name:   "success - no company or field changes",
			userID: userID,
			req:    &dtos.UpdateUserRequest{},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				user := &models.User{
					BaseModel: models.BaseModel{ID: userID},
					FirstName: "John",
				}
				userRepo.On("GetOneByID", userID, []string{"Companies"}).Return(user, nil)
				userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
			},
			expectedError: false,
		},
		{
			name:   "error - user not found",
			userID: "missing",
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{FirstName: "X"},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				userRepo.On("GetOneByID", "missing", []string{"Companies"}).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:       "error - user not found with sentry hub",
			userID:     "missing",
			withSentry: true,
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{FirstName: "X"},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				userRepo.On("GetOneByID", "missing", []string{"Companies"}).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:   "error - company not found on update",
			userID: userID,
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{
					Companies: []dtos.UpdateCompanyRequest{{ID: "missing-co"}},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				user := &models.User{BaseModel: models.BaseModel{ID: userID}}
				userRepo.On("GetOneByID", userID, []string{"Companies"}).Return(user, nil)
				companyRepo.On("GetOneByID", "missing-co").Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:       "error - company not found with sentry hub",
			userID:     userID,
			withSentry: true,
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{
					Companies: []dtos.UpdateCompanyRequest{{ID: "missing-co"}},
				},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				user := &models.User{BaseModel: models.BaseModel{ID: userID}}
				userRepo.On("GetOneByID", userID, []string{"Companies"}).Return(user, nil)
				companyRepo.On("GetOneByID", "missing-co").Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:   "error - database error on update",
			userID: userID,
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{FirstName: "X"},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				user := &models.User{BaseModel: models.BaseModel{ID: userID}, FirstName: "Old"}
				userRepo.On("GetOneByID", userID, []string{"Companies"}).Return(user, nil)
				userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(assert.AnError)
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
		{
			name:       "error - database error with sentry hub",
			userID:     userID,
			withSentry: true,
			req: &dtos.UpdateUserRequest{
				UserRequest: dtos.UserRequest{FirstName: "X"},
			},
			setupMocks: func(userRepo *MockUserRepository, companyRepo *MockCompanyRepository) {
				user := &models.User{BaseModel: models.BaseModel{ID: userID}, FirstName: "Old"}
				userRepo.On("GetOneByID", userID, []string{"Companies"}).Return(user, nil)
				userRepo.On("Update", mock.AnythingOfType("*models.User")).Return(assert.AnError)
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockCompanyRepo := new(MockCompanyRepository)
			mockCache := new(MockCache)
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockCompanyRepo)
			}

			service := &userService{
				userRepo:    mockUserRepo,
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			if tt.withSentry {
				ctx = ctxWithSentryHub()
			}

			result, err := service.Update(ctx, tt.userID, tt.req)
			if tt.expectedError {
				require.Error(t, err)
				appErr, ok := err.(*errors.AppError)
				require.True(t, ok)
				switch tt.errorType {
				case "NotFoundError":
					assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
				case "DatabaseError":
					assert.Equal(t, errors.ErrorTypeDatabase, appErr.Type)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.assertResult != nil {
					tt.assertResult(t, result)
				}
			}

			mockUserRepo.AssertExpectations(t)
			mockCompanyRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	userID := uuid.New().String()

	tests := []struct {
		name          string
		userID        string
		withSentry    bool
		setupMocks    func(*MockUserRepository)
		expectedError bool
		errorType     string
	}{
		{
			name:   "success",
			userID: userID,
			setupMocks: func(userRepo *MockUserRepository) {
				user := &models.User{BaseModel: models.BaseModel{ID: userID}}
				userRepo.On("GetOneByID", userID, mock.AnythingOfType("[]string")).Return(user, nil)
				userRepo.On("Delete", user).Return(nil)
			},
		},
		{
			name:   "error - user not found",
			userID: "missing",
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("GetOneByID", "missing", mock.AnythingOfType("[]string")).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:       "error - user not found with sentry hub",
			userID:     "missing",
			withSentry: true,
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("GetOneByID", "missing", mock.AnythingOfType("[]string")).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "NotFoundError",
		},
		{
			name:   "error - database error on delete",
			userID: userID,
			setupMocks: func(userRepo *MockUserRepository) {
				user := &models.User{BaseModel: models.BaseModel{ID: userID}}
				userRepo.On("GetOneByID", userID, mock.AnythingOfType("[]string")).Return(user, nil)
				userRepo.On("Delete", user).Return(assert.AnError)
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
		{
			name:       "error - database error with sentry hub",
			userID:     userID,
			withSentry: true,
			setupMocks: func(userRepo *MockUserRepository) {
				user := &models.User{BaseModel: models.BaseModel{ID: userID}}
				userRepo.On("GetOneByID", userID, mock.AnythingOfType("[]string")).Return(user, nil)
				userRepo.On("Delete", user).Return(assert.AnError)
			},
			expectedError: true,
			errorType:     "DatabaseError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockCompanyRepo := new(MockCompanyRepository)
			mockCache := new(MockCache)
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo)
			}

			service := &userService{
				userRepo:    mockUserRepo,
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			if tt.withSentry {
				ctx = ctxWithSentryHub()
			}

			err := service.Delete(ctx, tt.userID)
			if tt.expectedError {
				require.Error(t, err)
				appErr, ok := err.(*errors.AppError)
				require.True(t, ok)
				switch tt.errorType {
				case "NotFoundError":
					assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
				case "DatabaseError":
					assert.Equal(t, errors.ErrorTypeDatabase, appErr.Type)
				}
			} else {
				require.NoError(t, err)
			}
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_List(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		req           *dtos.UserPageableRequest
		withSentry    bool
		setupMocks    func(*MockUserRepository)
		expectedError bool
	}{
		{
			name: "success - with nil dates",
			req: &dtos.UserPageableRequest{
				PageableRequest: dtos.PageableRequest{Page: 1, PageSize: 10},
				Q:               "john",
				Sort:            []string{"-created_at"},
			},
			setupMocks: func(userRepo *MockUserRepository) {
				users := &dtos.DataResponse[models.User]{
					Data: []models.User{{BaseModel: models.BaseModel{ID: uuid.New().String()}, FirstName: "John"}},
				}
				userRepo.On("Get", mock.AnythingOfType("*dtos.UserPageableRequest"), []string{"Companies"}).Return(users, nil)
			},
		},
		{
			name: "success - with start and end dates",
			req: &dtos.UserPageableRequest{
				PageableRequest: dtos.PageableRequest{Page: 1, PageSize: 10},
				StartDate:       &start,
				EndDate:         &end,
			},
			setupMocks: func(userRepo *MockUserRepository) {
				users := &dtos.DataResponse[models.User]{
					Data: []models.User{{BaseModel: models.BaseModel{ID: uuid.New().String()}}},
				}
				userRepo.On("Get", mock.AnythingOfType("*dtos.UserPageableRequest"), []string{"Companies"}).Return(users, nil)
			},
		},
		{
			name: "error - database error",
			req: &dtos.UserPageableRequest{
				PageableRequest: dtos.PageableRequest{Page: 1, PageSize: 10},
			},
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("Get", mock.AnythingOfType("*dtos.UserPageableRequest"), []string{"Companies"}).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
		{
			name:       "error - database error with sentry hub",
			withSentry: true,
			req: &dtos.UserPageableRequest{
				PageableRequest: dtos.PageableRequest{Page: 1, PageSize: 10},
			},
			setupMocks: func(userRepo *MockUserRepository) {
				userRepo.On("Get", mock.AnythingOfType("*dtos.UserPageableRequest"), []string{"Companies"}).Return(nil, assert.AnError)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := new(MockUserRepository)
			mockCompanyRepo := new(MockCompanyRepository)
			mockCache := new(MockCache)
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo)
			}

			service := &userService{
				userRepo:    mockUserRepo,
				companyRepo: mockCompanyRepo,
				cache:       mockCache,
			}

			ctx := context.Background()
			if tt.withSentry {
				ctx = ctxWithSentryHub()
			}

			result, err := service.List(ctx, tt.req)
			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Data)
			}
			mockUserRepo.AssertExpectations(t)
		})
	}
}
