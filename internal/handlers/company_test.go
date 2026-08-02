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
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockCompanyService struct {
	mock.Mock
}

func (m *mockCompanyService) Create(ctx context.Context, req *dtos.CreateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *mockCompanyService) GetOneByID(ctx context.Context, companyID string) (*models.Company, error) {
	args := m.Called(ctx, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *mockCompanyService) Update(ctx context.Context, companyID string, req *dtos.UpdateCompanyRequest) (*models.Company, error) {
	args := m.Called(ctx, companyID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Company), args.Error(1)
}

func (m *mockCompanyService) Delete(ctx context.Context, companyID string) error {
	return m.Called(ctx, companyID).Error(0)
}

func (m *mockCompanyService) List(ctx context.Context, pageableRequest *dtos.CompanyPageableRequest) (*dtos.DataResponse[models.Company], error) {
	args := m.Called(ctx, pageableRequest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dtos.DataResponse[models.Company]), args.Error(1)
}

func sampleCompany() *models.Company {
	return &models.Company{
		BaseModel:  models.BaseModel{ID: "co-1", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		Name:       "Acme",
		KeycloakID: "kc-1",
	}
}

func newCompanyHandler(svc *mockCompanyService) *CompanyHandler {
	return ProvideCompanyHandler(svc, &config.Config{KeycloakKeyClaim: "claims"}, validator.New())
}

func withCompanyClaims(c *echo.Context) {
	c.Set("claims", &auth.TokenClaims{Sub: "sub-1"})
}

func TestProvideCompanyHandler(t *testing.T) {
	h := ProvideCompanyHandler(new(mockCompanyService), &config.Config{KeycloakKeyClaim: "claims"}, validator.New())
	require.NotNil(t, h)
}

func TestCompanyHandler_CreateCompany(t *testing.T) {
	validBody := `{"name":"Acme","keycloak_id":"kc-1"}`

	tests := []struct {
		name       string
		auth       bool
		body       string
		setup      func(*mockCompanyService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{name: "invalid json", auth: true, body: `{`, wantStatus: http.StatusBadRequest},
		{name: "validation error", auth: true, body: `{"name":"A"}`, wantStatus: http.StatusBadRequest},
		{
			name: "service error",
			auth: true,
			body: validBody,
			setup: func(m *mockCompanyService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateCompanyRequest")).
					Return(nil, errors.New("create failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			body: validBody,
			setup: func(m *mockCompanyService) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*dtos.CreateCompanyRequest")).
					Return(sampleCompany(), nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockCompanyService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := ProvideCompanyHandler(svc, &config.Config{KeycloakKeyClaim: "claims"}, validator.New())
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/companies", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.auth {
				withCompanyClaims(c)
			}
			require.NoError(t, h.CreateCompany(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestCompanyHandler_GetOneByID(t *testing.T) {
	tests := []struct {
		name       string
		auth       bool
		setup      func(*mockCompanyService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name: "service error",
			auth: true,
			setup: func(m *mockCompanyService) {
				m.On("GetOneByID", mock.Anything, "co-1").Return(nil, errors.New("missing"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			setup: func(m *mockCompanyService) {
				m.On("GetOneByID", mock.Anything, "co-1").Return(sampleCompany(), nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockCompanyService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := newCompanyHandler(svc)
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/companies/co-1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "co-1"}})
			if tt.auth {
				withCompanyClaims(c)
			}
			require.NoError(t, h.GetOneByID(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestCompanyHandler_UpdateCompany(t *testing.T) {
	validBody := `{"name":"Acme","keycloak_id":"kc-1"}`

	tests := []struct {
		name       string
		auth       bool
		body       string
		setup      func(*mockCompanyService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{name: "invalid json", auth: true, body: `{`, wantStatus: http.StatusBadRequest},
		{name: "validation error", auth: true, body: `{"name":"A"}`, wantStatus: http.StatusBadRequest},
		{
			name: "service error",
			auth: true,
			body: validBody,
			setup: func(m *mockCompanyService) {
				m.On("Update", mock.Anything, "co-1", mock.AnythingOfType("*dtos.UpdateCompanyRequest")).
					Return(nil, errors.New("update failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			body: validBody,
			setup: func(m *mockCompanyService) {
				m.On("Update", mock.Anything, "co-1", mock.AnythingOfType("*dtos.UpdateCompanyRequest")).
					Return(sampleCompany(), nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockCompanyService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := ProvideCompanyHandler(svc, &config.Config{KeycloakKeyClaim: "claims"}, validator.New())
			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/companies/co-1", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "co-1"}})
			if tt.auth {
				withCompanyClaims(c)
			}
			require.NoError(t, h.UpdateCompany(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestCompanyHandler_DeleteCompany(t *testing.T) {
	tests := []struct {
		name       string
		auth       bool
		setup      func(*mockCompanyService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name: "service error",
			auth: true,
			setup: func(m *mockCompanyService) {
				m.On("Delete", mock.Anything, "co-1").Return(errors.New("delete failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "success",
			auth: true,
			setup: func(m *mockCompanyService) {
				m.On("Delete", mock.Anything, "co-1").Return(nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockCompanyService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := newCompanyHandler(svc)
			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/companies/co-1", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "co-1"}})
			if tt.auth {
				withCompanyClaims(c)
			}
			require.NoError(t, h.DeleteCompany(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}

func TestCompanyHandler_GetCompanies(t *testing.T) {
	tests := []struct {
		name       string
		auth       bool
		query      string
		setup      func(*mockCompanyService)
		wantStatus int
	}{
		{name: "unauthorized", wantStatus: http.StatusUnauthorized},
		{
			name:       "invalid date",
			auth:       true,
			query:      "?end_date=bad",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "service error",
			auth:  true,
			query: "?page=x&page_size=y",
			setup: func(m *mockCompanyService) {
				m.On("List", mock.Anything, mock.AnythingOfType("*dtos.CompanyPageableRequest")).
					Return(nil, errors.New("list failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:  "success",
			auth:  true,
			query: "?page=1&page_size=10&q=Ac",
			setup: func(m *mockCompanyService) {
				m.On("List", mock.Anything, mock.AnythingOfType("*dtos.CompanyPageableRequest")).
					Return(&dtos.DataResponse[models.Company]{
						Data:     []models.Company{*sampleCompany()},
						Pageable: &dtos.Pageable{Page: 1, PageSize: 10, Total: 1},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "defaults for bad pagination",
			auth:  true,
			query: "?page=-1&page_size=-5",
			setup: func(m *mockCompanyService) {
				m.On("List", mock.Anything, mock.MatchedBy(func(pr *dtos.CompanyPageableRequest) bool {
					return pr.Page == 1 && pr.PageSize == 10
				})).Return(&dtos.DataResponse[models.Company]{Data: nil}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(mockCompanyService)
			if tt.setup != nil {
				tt.setup(svc)
			}
			h := newCompanyHandler(svc)
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/companies"+tt.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			if tt.auth {
				withCompanyClaims(c)
			}
			require.NoError(t, h.GetCompanies(c))
			assert.Equal(t, tt.wantStatus, rec.Code)
			svc.AssertExpectations(t)
		})
	}
}
