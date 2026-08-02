package services

import (
	"context"
	"testing"

	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/integration/auth"

	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthProvider is a mock implementation of auth.AuthService
type MockAuthProvider struct {
	mock.Mock
}

func (m *MockAuthProvider) GetRealm() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAuthProvider) DecodeAccessToken(ctx context.Context, token string, realm string, claims *auth.TokenClaims) (*auth.TokenClaims, error) {
	args := m.Called(ctx, token, realm, claims)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenClaims), args.Error(1)
}

func (m *MockAuthProvider) ValidateToken(token string) (*gocloak.IntroSpectTokenResult, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.IntroSpectTokenResult), args.Error(1)
}

func (m *MockAuthProvider) ClientLogin() (*auth.TokenInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenInfo), args.Error(1)
}

func (m *MockAuthProvider) GetUserInfo(token string) (*auth.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthProvider) GetClaimsKey() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAuthProvider) GetRequestingPartyToken(ctx context.Context, accessToken string, opts auth.RequestingPartyTokenOptions) (*auth.JWT, error) {
	args := m.Called(ctx, accessToken, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.JWT), args.Error(1)
}

func (m *MockAuthProvider) CreateUser(ctx context.Context, adminToken string, userDto *dtos.CreateUserRequest) (*auth.User, error) {
	args := m.Called(ctx, adminToken, userDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *MockAuthProvider) SetPassword(ctx context.Context, adminToken string, userID string, password string, temporary bool) error {
	args := m.Called(ctx, adminToken, userID, password, temporary)
	return args.Error(0)
}

func (m *MockAuthProvider) SendVerificationMail(ctx context.Context, adminToken string, userID string, params auth.SendVerificationMailParams) error {
	args := m.Called(ctx, adminToken, userID, params)
	return args.Error(0)
}

func (m *MockAuthProvider) GetClientID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAuthProvider) GetRedirectURI() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAuthProvider) GetOrganization(userClaims *auth.TokenClaims) (auth.Organization, error) {
	args := m.Called(userClaims)
	if args.Get(0) == nil {
		return auth.Organization{}, args.Error(1)
	}
	return args.Get(0).(auth.Organization), args.Error(1)
}

func (m *MockAuthProvider) AddUserToOrganization(ctx context.Context, adminToken string, userID string, organizationID string) error {
	args := m.Called(ctx, adminToken, userID, organizationID)
	return args.Error(0)
}

func (m *MockAuthProvider) AddClientRolesToUser(ctx context.Context, adminToken string, userID string, clientID string, role string) error {
	args := m.Called(ctx, adminToken, userID, clientID, role)
	return args.Error(0)
}

func (m *MockAuthProvider) UpdateUser(ctx context.Context, adminToken string, userID string, userDto *dtos.UpdateUserRequest) error {
	args := m.Called(ctx, adminToken, userID, userDto)
	return args.Error(0)
}

func TestAuthService_ValidateUserToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		setupMock     func(*MockAuthProvider)
		expectedError bool
		errorType     string
	}{
		{
			name:  "success - valid token",
			token: "valid-token-123",
			setupMock: func(m *MockAuthProvider) {
				introspectResult := &gocloak.IntroSpectTokenResult{
					Active: boolPtr(true),
				}
				m.On("ValidateToken", "valid-token-123").Return(introspectResult, nil)
			},
			expectedError: false,
		},
		{
			name:  "error - invalid token",
			token: "invalid-token",
			setupMock: func(m *MockAuthProvider) {
				m.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "UnauthorizedError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthProvider := new(MockAuthProvider)
			if tt.setupMock != nil {
				tt.setupMock(mockAuthProvider)
			}

			service := &AuthService{
				authProvider: mockAuthProvider,
			}

			result, err := service.ValidateUserToken(tt.token)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					appErr, ok := err.(*errors.AppError)
					require.True(t, ok, "Expected AppError")
					assert.Equal(t, errors.ErrorTypeUnauthorized, appErr.Type)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.True(t, *result.Active)
			}

			mockAuthProvider.AssertExpectations(t)
		})
	}
}

func TestAuthService_HasRole(t *testing.T) {
	tests := []struct {
		name     string
		user     *auth.User
		role     string
		expected bool
	}{
		{
			name: "success - user has role",
			user: &auth.User{
				Roles: []string{"admin", "user"},
			},
			role:     "admin",
			expected: true,
		},
		{
			name: "success - user does not have role",
			user: &auth.User{
				Roles: []string{"user"},
			},
			role:     "admin",
			expected: false,
		},
		{
			name: "success - empty roles",
			user: &auth.User{
				Roles: []string{},
			},
			role:     "admin",
			expected: false,
		},
		{
			name: "success - multiple roles, has target role",
			user: &auth.User{
				Roles: []string{"admin", "manager", "user"},
			},
			role:     "manager",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &AuthService{}

			result := service.HasRole(tt.user, tt.role)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthService_HasAnyRole(t *testing.T) {
	tests := []struct {
		name     string
		user     *auth.User
		roles    []string
		expected bool
	}{
		{
			name: "success - user has one of the roles",
			user: &auth.User{
				Roles: []string{"admin", "user"},
			},
			roles:    []string{"admin", "manager"},
			expected: true,
		},
		{
			name: "success - user does not have any of the roles",
			user: &auth.User{
				Roles: []string{"user"},
			},
			roles:    []string{"admin", "manager"},
			expected: false,
		},
		{
			name: "success - user has multiple matching roles",
			user: &auth.User{
				Roles: []string{"admin", "manager", "user"},
			},
			roles:    []string{"admin", "manager"},
			expected: true,
		},
		{
			name: "success - empty roles list",
			user: &auth.User{
				Roles: []string{"admin"},
			},
			roles:    []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &AuthService{}

			result := service.HasAnyRole(tt.user, tt.roles...)

			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for creating pointers to basic types
func boolPtr(b bool) *bool {
	return &b
}
