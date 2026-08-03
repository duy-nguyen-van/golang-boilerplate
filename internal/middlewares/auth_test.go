package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/integration/auth"

	"github.com/Nerzal/gocloak/v13"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) GetRealm() string {
	return m.Called().String(0)
}

func (m *mockAuthService) DecodeAccessToken(ctx context.Context, token string, realm string, claims *auth.TokenClaims) (*auth.TokenClaims, error) {
	args := m.Called(ctx, token, realm, claims)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenClaims), args.Error(1)
}

func (m *mockAuthService) ValidateToken(token string) (*gocloak.IntroSpectTokenResult, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*gocloak.IntroSpectTokenResult), args.Error(1)
}

func (m *mockAuthService) ClientLogin() (*auth.TokenInfo, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenInfo), args.Error(1)
}

func (m *mockAuthService) GetUserInfo(token string) (*auth.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *mockAuthService) GetClaimsKey() string {
	return m.Called().String(0)
}

func (m *mockAuthService) GetRequestingPartyToken(ctx context.Context, accessToken string, opts auth.RequestingPartyTokenOptions) (*auth.JWT, error) {
	args := m.Called(ctx, accessToken, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.JWT), args.Error(1)
}

func (m *mockAuthService) CreateUser(ctx context.Context, adminToken string, userDto *dtos.CreateUserRequest) (*auth.User, error) {
	args := m.Called(ctx, adminToken, userDto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.User), args.Error(1)
}

func (m *mockAuthService) SetPassword(ctx context.Context, adminToken string, userID string, password string, temporary bool) error {
	return m.Called(ctx, adminToken, userID, password, temporary).Error(0)
}

func (m *mockAuthService) SendVerificationMail(ctx context.Context, adminToken string, userID string, params auth.SendVerificationMailParams) error {
	return m.Called(ctx, adminToken, userID, params).Error(0)
}

func (m *mockAuthService) GetClientID() string {
	return m.Called().String(0)
}

func (m *mockAuthService) GetRedirectURI() string {
	return m.Called().String(0)
}

func (m *mockAuthService) GetOrganization(userClaims *auth.TokenClaims) (auth.Organization, error) {
	args := m.Called(userClaims)
	if args.Get(0) == nil {
		return auth.Organization{}, args.Error(1)
	}
	return args.Get(0).(auth.Organization), args.Error(1)
}

func (m *mockAuthService) AddUserToOrganization(ctx context.Context, adminToken string, userID string, organizationID string) error {
	return m.Called(ctx, adminToken, userID, organizationID).Error(0)
}

func (m *mockAuthService) AddClientRolesToUser(ctx context.Context, adminToken string, userID string, clientID string, role string) error {
	return m.Called(ctx, adminToken, userID, clientID, role).Error(0)
}

func (m *mockAuthService) UpdateUser(ctx context.Context, adminToken string, userID string, userDto *dtos.UpdateUserRequest) error {
	return m.Called(ctx, adminToken, userID, userDto).Error(0)
}

func boolPtr(b bool) *bool { return &b }

func newEchoContext(method, path string, headers map[string]string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{AppEnv: config.EnvironmentTest, KeycloakKeyClaim: "claims"}
	nextCalled := false
	next := func(c *echo.Context) error {
		nextCalled = true
		return c.JSON(http.StatusOK, map[string]string{"ok": "true"})
	}

	tests := []struct {
		name       string
		headers    map[string]string
		setup      func(*mockAuthService)
		wantStatus int
		wantNext   bool
		wantBody   string
	}{
		{
			name:       "missing authorization header",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Authorization header required",
		},
		{
			name:       "invalid authorization format",
			headers:    map[string]string{"Authorization": "Basic abc"},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Invalid authorization header format",
		},
		{
			name:       "empty bearer token",
			headers:    map[string]string{"Authorization": "Bearer "},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Token required",
		},
		{
			name:    "invalid token",
			headers: map[string]string{"Authorization": "Bearer bad"},
			setup: func(m *mockAuthService) {
				m.On("ValidateToken", "bad").Return(nil, errors.New("invalid"))
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Invalid token",
		},
		{
			name:    "inactive token",
			headers: map[string]string{"Authorization": "Bearer inactive"},
			setup: func(m *mockAuthService) {
				m.On("ValidateToken", "inactive").Return(&gocloak.IntroSpectTokenResult{Active: boolPtr(false)}, nil)
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Token is not active",
		},
		{
			name:    "invalid claims",
			headers: map[string]string{"Authorization": "Bearer claims-bad"},
			setup: func(m *mockAuthService) {
				m.On("ValidateToken", "claims-bad").Return(&gocloak.IntroSpectTokenResult{Active: boolPtr(true)}, nil)
				m.On("GetRealm").Return("realm")
				m.On("DecodeAccessToken", mock.Anything, "claims-bad", "realm", mock.Anything).
					Return(nil, errors.New("decode failed"))
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "Invalid token claims",
		},
		{
			name:    "success",
			headers: map[string]string{"Authorization": "Bearer good"},
			setup: func(m *mockAuthService) {
				m.On("ValidateToken", "good").Return(&gocloak.IntroSpectTokenResult{Active: boolPtr(true)}, nil)
				m.On("GetRealm").Return("realm")
				m.On("DecodeAccessToken", mock.Anything, "good", "realm", mock.Anything).
					Run(func(args mock.Arguments) {
						claims := args.Get(3).(*auth.TokenClaims)
						claims.Sub = "user-1"
					}).
					Return(&auth.TokenClaims{Sub: "user-1"}, nil)
				m.On("GetClaimsKey").Return("claims")
			},
			wantStatus: http.StatusOK,
			wantNext:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled = false
			authSvc := new(mockAuthService)
			if tt.setup != nil {
				tt.setup(authSvc)
			}

			c, rec := newEchoContext(http.MethodGet, "/secure", tt.headers)
			err := AuthMiddleware(cfg, authSvc)(next)(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantNext, nextCalled)
			assert.Contains(t, rec.Body.String(), tt.wantBody)
			authSvc.AssertExpectations(t)
		})
	}
}

func TestRequireRole(t *testing.T) {
	cfg := &config.Config{KeycloakKeyClaim: "claims", KeycloakClientID: "app-client"}
	next := func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}

	tests := []struct {
		name       string
		claims     any
		roles      []string
		wantStatus int
	}{
		{
			name:       "missing claims",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "insufficient permissions",
			claims: &auth.TokenClaims{
				RealmAccess: struct {
					Roles []string `json:"roles"`
				}{Roles: []string{"viewer"}},
			},
			roles:      []string{"admin"},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "realm role match",
			claims: &auth.TokenClaims{
				RealmAccess: struct {
					Roles []string `json:"roles"`
				}{Roles: []string{"admin"}},
			},
			roles:      []string{"admin"},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "client role match",
			claims: &auth.TokenClaims{
				ResourceAccess: map[string]struct {
					Roles []string `json:"roles"`
				}{
					"app-client": {Roles: []string{"editor"}},
				},
			},
			roles:      []string{"editor"},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, rec := newEchoContext(http.MethodGet, "/role", nil)
			if tt.claims != nil {
				c.Set(cfg.KeycloakKeyClaim, tt.claims)
			}
			err := RequireRole(cfg, tt.roles...)(next)(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestExtractRolesFromClaims(t *testing.T) {
	t.Parallel()
	claims := &auth.TokenClaims{
		RealmAccess: struct {
			Roles []string `json:"roles"`
		}{Roles: []string{"realm-admin"}},
		ResourceAccess: map[string]struct {
			Roles []string `json:"roles"`
		}{
			"client-a": {Roles: []string{"client-role"}},
			"other":    {Roles: []string{"ignored"}},
		},
	}
	roles := extractRolesFromClaims(claims, "client-a")
	assert.Equal(t, []string{"realm-admin", "client-role"}, roles)

	rolesMissingClient := extractRolesFromClaims(claims, "missing")
	assert.Equal(t, []string{"realm-admin"}, rolesMissingClient)
}

func TestRequirePermission(t *testing.T) {
	cfg := &config.Config{KeycloakClientID: "app-client"}
	next := func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	}

	tests := []struct {
		name       string
		headers    map[string]string
		setup      func(*mockAuthService)
		wantStatus int
	}{
		{
			name:       "invalid auth header",
			headers:    map[string]string{"Authorization": "Basic x"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty token",
			headers:    map[string]string{"Authorization": "Bearer "},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "rpt evaluation failed",
			headers: map[string]string{"Authorization": "Bearer access"},
			setup: func(m *mockAuthService) {
				m.On("GetRequestingPartyToken", mock.Anything, "access", mock.Anything).
					Return(nil, errors.New("uma failed"))
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "invalid rpt claims",
			headers: map[string]string{"Authorization": "Bearer access"},
			setup: func(m *mockAuthService) {
				m.On("GetRequestingPartyToken", mock.Anything, "access", mock.Anything).
					Return(&auth.JWT{AccessToken: "rpt"}, nil)
				m.On("GetRealm").Return("realm")
				m.On("DecodeAccessToken", mock.Anything, "rpt", "realm", mock.Anything).
					Return(nil, errors.New("bad rpt"))
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "insufficient permissions",
			headers: map[string]string{"Authorization": "Bearer access"},
			setup: func(m *mockAuthService) {
				m.On("GetRequestingPartyToken", mock.Anything, "access", mock.Anything).
					Return(&auth.JWT{AccessToken: "rpt"}, nil)
				m.On("GetRealm").Return("realm")
				m.On("DecodeAccessToken", mock.Anything, "rpt", "realm", mock.Anything).
					Run(func(args mock.Arguments) {
						claims := args.Get(3).(*auth.TokenClaims)
						claims.Authorization.Permissions = []struct {
							ResourceName string   `json:"rsname"`
							Scopes       []string `json:"scopes"`
						}{
							{ResourceName: "users", Scopes: []string{"read"}},
						}
					}).
					Return(&auth.TokenClaims{}, nil)
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:    "success",
			headers: map[string]string{"Authorization": "Bearer access"},
			setup: func(m *mockAuthService) {
				m.On("GetRequestingPartyToken", mock.Anything, "access", mock.Anything).
					Return(&auth.JWT{AccessToken: "rpt"}, nil)
				m.On("GetRealm").Return("realm")
				m.On("DecodeAccessToken", mock.Anything, "rpt", "realm", mock.Anything).
					Run(func(args mock.Arguments) {
						claims := args.Get(3).(*auth.TokenClaims)
						claims.Authorization.Permissions = []struct {
							ResourceName string   `json:"rsname"`
							Scopes       []string `json:"scopes"`
						}{
							{ResourceName: "users", Scopes: []string{"write"}},
						}
					}).
					Return(&auth.TokenClaims{}, nil)
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc := new(mockAuthService)
			if tt.setup != nil {
				tt.setup(authSvc)
			}
			c, rec := newEchoContext(http.MethodGet, "/perm", tt.headers)
			err := RequirePermission(cfg, authSvc, "users", "write")(next)(c)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			authSvc.AssertExpectations(t)
		})
	}
}
