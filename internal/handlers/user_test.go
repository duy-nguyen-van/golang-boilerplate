package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/integration/auth"
	"golang-boilerplate/internal/models"

	"github.com/go-playground/validator/v10"
	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Create(ctx context.Context, req *dtos.CreateUserRequest) (*models.User, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserService) GetOneByID(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserService) Update(ctx context.Context, userID string, req *dtos.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserService) Delete(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserService) List(ctx context.Context, pageableRequest *dtos.UserPageableRequest) (*dtos.DataResponse[models.User], error) {
	args := m.Called(ctx, pageableRequest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.DataResponse[models.User]), args.Error(1)
}

type mockRestClient struct {
	mock.Mock
}

func (m *mockRestClient) Post(endpoint string, body, okResult, failedResult interface{}, headers map[string]string) (*resty.Response, error) {
	args := m.Called(endpoint, body, okResult, failedResult, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*resty.Response), args.Error(1)
}

func (m *mockRestClient) Put(endpoint string, body, okResult, failedResult interface{}, headers map[string]string) (*resty.Response, error) {
	args := m.Called(endpoint, body, okResult, failedResult, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*resty.Response), args.Error(1)
}

func (m *mockRestClient) Get(endpoint string, result interface{}, headers map[string]string, queryParams string) (*resty.Response, error) {
	args := m.Called(endpoint, result, headers, queryParams)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*resty.Response), args.Error(1)
}

func (m *mockRestClient) Patch(endpoint string, body, okResult, failedResult interface{}, headers map[string]string) (*resty.Response, error) {
	args := m.Called(endpoint, body, okResult, failedResult, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*resty.Response), args.Error(1)
}

func testUserCfg() *config.Config {
	return &config.Config{KeycloakKeyClaim: "claims"}
}

func sampleUser() *models.User {
	return &models.User{
		BaseModel: models.BaseModel{ID: "user-1", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}
}

func newUserHandler(svc *mockUserService, rest *mockRestClient) *UserHandler {
	if rest == nil {
		rest = new(mockRestClient)
	}
	return ProvideUserHandler(svc, testUserCfg(), validator.New(), rest)
}

func withClaims(c *echo.Context) {
	c.Set("claims", &auth.TokenClaims{Sub: "sub-1"})
}

func TestProvideUserHandler(t *testing.T) {
	h := ProvideUserHandler(new(mockUserService), testUserCfg(), validator.New(), new(mockRestClient))
	require.NotNil(t, h)
	require.NotNil(t, h.userService)
}

func TestUserHandler_CreateUser(t *testing.T) {
	validBody := `{"email":"john@example.com","first_name":"John","last_name":"Doe"}`

	tests := []struct {
		name       string
		auth       bool
		body       string
		setup      func(*mockUserService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name:       "invalid json",
			auth:       true,
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "validation error",
			auth:       true,
			body:       `{"email":"not-an-email"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			auth: true,
			body: validBody,
			setup: func(m *mockUserService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateUserRequest")).
					Return(nil, errors.New("create failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			body: validBody,
			setup: func(m *mockUserService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateUserRequest")).
					Return(sampleUser(), nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockUserService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := ProvideUserHandler(svc, testUserCfg(), validator.New(), new(mockRestClient))
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.auth {
				withClaims(c)
			}
			require.NoError(t, h.CreateUser(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetOneByID(t *testing.T) {
	tests := []struct {
		name       string
		auth       bool
		setup      func(*mockUserService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name: "service error",
			auth: true,
			setup: func(m *mockUserService) {
				m.On("GetOneByID", mock.Anything, "user-1").Return(nil, errors.New("not found"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			setup: func(m *mockUserService) {
				m.On("GetOneByID", mock.Anything, "user-1").Return(sampleUser(), nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockUserService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := newUserHandler(svc, nil)
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/users/user-1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "user-1"}})
			if tt.auth {
				withClaims(c)
			}
			require.NoError(t, h.GetOneByID(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	validBody := `{"email":"john@example.com","first_name":"John","last_name":"Doe"}`

	tests := []struct {
		name       string
		auth       bool
		body       string
		setup      func(*mockUserService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{name: "invalid json", auth: true, body: `{`, wantStatus: http.StatusBadRequest},
		{name: "validation error", auth: true, body: `{"email":"bad"}`, wantStatus: http.StatusBadRequest},
		{
			name: "service error",
			auth: true,
			body: validBody,
			setup: func(m *mockUserService) {
				m.On("Update", mock.Anything, "user-1", mock.AnythingOfType("*dtos.UpdateUserRequest")).
					Return(nil, errors.New("update failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			body: validBody,
			setup: func(m *mockUserService) {
				m.On("Update", mock.Anything, "user-1", mock.AnythingOfType("*dtos.UpdateUserRequest")).
					Return(sampleUser(), nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockUserService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := ProvideUserHandler(svc, testUserCfg(), validator.New(), new(mockRestClient))
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/users/user-1", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "user-1"}})
			if tt.auth {
				withClaims(c)
			}
			require.NoError(t, h.UpdateUser(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		auth       bool
		setup      func(*mockUserService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name: "service error",
			auth: true,
			setup: func(m *mockUserService) {
				m.On("Delete", mock.Anything, "user-1").Return(errors.New("delete failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			setup: func(m *mockUserService) {
				m.On("Delete", mock.Anything, "user-1").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockUserService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := newUserHandler(svc, nil)
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/users/user-1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "user-1"}})
			if tt.auth {
				withClaims(c)
			}
			require.NoError(t, h.DeleteUser(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUsers(t *testing.T) {
	tests := []struct {
		name       string
		auth       bool
		query      string
		setup      func(*mockUserService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name:       "invalid sort",
			auth:       true,
			query:      "?sort=unknown",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid date",
			auth:       true,
			query:      "?start_date=not-a-date",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "service error",
			auth:  true,
			query: "?page=abc&page_size=abc",
			setup: func(m *mockUserService) {
				m.On("List", mock.Anything, mock.AnythingOfType("*dtos.UserPageableRequest")).
					Return(nil, errors.New("list failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:  "success with defaults and sort",
			auth:  true,
			query: "?page=2&page_size=5&q=Jo&sort=name&sort=-created_at",
			setup: func(m *mockUserService) {
				m.On("List", mock.Anything, mock.AnythingOfType("*dtos.UserPageableRequest")).
					Return(&dtos.DataResponse[models.User]{
						Data:     []models.User{*sampleUser()},
						Pageable: &dtos.Pageable{Page: 2, PageSize: 5, Total: 1},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success page size negative defaults",
			auth:  true,
			query: "?page=0&page_size=-1",
			setup: func(m *mockUserService) {
				m.On("List", mock.Anything, mock.MatchedBy(func(pr *dtos.UserPageableRequest) bool {
					return pr.Page == 1 && pr.PageSize == 10
				})).Return(&dtos.DataResponse[models.User]{Data: []models.User{}}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockUserService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := newUserHandler(svc, nil)
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/users"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.auth {
				withClaims(c)
			}
			require.NoError(t, h.GetUsers(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestUserHandler_TestRestClient(t *testing.T) {
	tests := []struct {
		name       string
		auth       bool
		setup      func(*mockRestClient)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name: "rest error",
			auth: true,
			setup: func(m *mockRestClient) {
				m.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("network"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			setup: func(m *mockRestClient) {
				m.On("Get", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						result := args.Get(1).(*[]map[string]interface{})
						*result = []map[string]interface{}{{"id": float64(1)}}
					}).
					Return(&resty.Response{}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rest := new(mockRestClient)
			if tt.setup != nil {
				tt.setup(rest)
			}
			h := newUserHandler(new(mockUserService), rest)
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/users/test-rest-client", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.auth {
				withClaims(c)
			}
			require.NoError(t, h.TestRestClient(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			rest.AssertExpectations(t)
		})
	}
}
