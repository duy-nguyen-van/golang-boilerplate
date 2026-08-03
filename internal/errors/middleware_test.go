package errors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/request"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.Config
		panicValue any
		withHub    bool
		wantStatus int
	}{
		{
			name:       "development exposes panic details for error panic",
			cfg:        &config.Config{AppEnv: config.EnvironmentDevelopment},
			panicValue: fmt.Errorf("dev panic"),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "nil config hides panic details for string panic",
			cfg:        nil,
			panicValue: "string panic",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "production hides panic details with sentry hub",
			cfg:        &config.Config{AppEnv: config.EnvironmentProduction},
			panicValue: fmt.Errorf("prod panic"),
			withHub:    true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/panic", nil)

			ctx := req.Context()
			ctx = request.NewCorrelationIDContext(ctx, "corr-123")
			ctx = request.NewLanguageCodeContext(ctx, "en")
			if tt.withHub {
				hub := sentry.CurrentHub().Clone()
				ctx = sentry.SetHubOnContext(ctx, hub)
			}
			req = req.WithContext(ctx)

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mw := RecoveryMiddleware(tt.cfg)
			handler := mw(func(_ *echo.Context) error {
				panic(tt.panicValue)
			})

			err := handler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), "INTERNAL_ERROR")
		})
	}

	t.Run("passes through without panic", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ok", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := RecoveryMiddleware(&config.Config{AppEnv: config.EnvironmentTest})
		handler := mw(func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})

		require.NoError(t, handler(c))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}

func TestErrorMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		nextErr    error
		wantStatus int
		wantPass   bool
	}{
		{
			name:     "no error",
			path:     "/api/v1/ok",
			nextErr:  nil,
			wantPass: true,
		},
		{
			name:       "handles app error",
			path:       "/api/v1/users",
			nextErr:    NotFoundError("User", nil),
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "skips swagger route",
			path:     "/swagger/index.html",
			nextErr:  fmt.Errorf("swagger err"),
			wantPass: true,
		},
		{
			name:     "skips docs route",
			path:     "/docs/openapi",
			nextErr:  fmt.Errorf("docs err"),
			wantPass: true,
		},
		{
			name:     "skips favicon.ico",
			path:     "/favicon.ico",
			nextErr:  fmt.Errorf("favicon err"),
			wantPass: true,
		},
		{
			name:     "skips favicon path fragment",
			path:     "/assets/favicon-32.png",
			nextErr:  fmt.Errorf("favicon err"),
			wantPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mw := ErrorMiddleware()
			handler := mw(func(_ *echo.Context) error {
				return tt.nextErr
			})

			err := handler(c)
			if tt.wantPass {
				assert.Equal(t, tt.nextErr, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestRouteHelpers(t *testing.T) {
	t.Parallel()

	assert.True(t, isSwaggerRoute("/swagger/index.html"))
	assert.True(t, isSwaggerRoute("/api/docs"))
	assert.False(t, isSwaggerRoute("/api/v1/users"))

	assert.True(t, isFaviconRoute("/favicon.ico"))
	assert.True(t, isFaviconRoute("/static/favicon.png"))
	assert.False(t, isFaviconRoute("/api/v1/users"))
}
