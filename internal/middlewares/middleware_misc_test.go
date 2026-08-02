package middlewares

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/integration/auth"
	"golang-boilerplate/internal/request"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthMiddleware(t *testing.T) {
	cfg := config.Config{BasicAuthUsername: "admin", BasicAuthPassword: "secret"}
	e := echo.New()
	e.Use(BasicAuthMiddleware(cfg))
	e.GET("/docs", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	tests := []struct {
		name       string
		user       string
		pass       string
		wantStatus int
	}{
		{name: "valid", user: "admin", pass: "secret", wantStatus: http.StatusOK},
		{name: "invalid password", user: "admin", pass: "wrong", wantStatus: http.StatusUnauthorized},
		{name: "missing", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/docs", nil)
			if tt.user != "" {
				req.SetBasicAuth(tt.user, tt.pass)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestCORS(t *testing.T) {
	e := echo.New()
	e.Use(CORS())
	e.GET("/ping", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set(echo.HeaderOrigin, "http://example.com")
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowHeaders), "X-CSRF-Token")
}

func TestCSRFAndExposeToken(t *testing.T) {
	tests := []struct {
		name   string
		env    config.Environment
		secure bool
	}{
		{name: "development", env: config.EnvironmentDevelopment, secure: false},
		{name: "production", env: config.EnvironmentProduction, secure: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{AppEnv: tt.env}
			e := echo.New()
			e.Use(CSRF(cfg))
			e.Use(ExposeCSRFToken())
			e.GET("/form", func(c *echo.Context) error {
				return c.NoContent(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/form", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.NotEmpty(t, rec.Header().Get("X-CSRF-Token"))

			cookies := rec.Result().Cookies()
			require.NotEmpty(t, cookies)
			var csrfCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "csrf_token" {
					csrfCookie = c
					break
				}
			}
			require.NotNil(t, csrfCookie)
			assert.Equal(t, tt.secure, csrfCookie.Secure)
		})
	}

	t.Run("expose without token skips header", func(t *testing.T) {
		e := echo.New()
		e.Use(ExposeCSRFToken())
		e.GET("/x", func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("X-CSRF-Token"))
	})

	t.Run("expose empty token skips header", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(echomw.DefaultCSRFConfig.ContextKey, "")
				return next(c)
			}
		})
		e.Use(ExposeCSRFToken())
		e.GET("/x", func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("X-CSRF-Token"))
	})
}

func TestSecurity(t *testing.T) {
	e := echo.New()
	e.Use(Security())
	e.GET("/", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// HSTS is only written for TLS / forwarded HTTPS requests.
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, "1; mode=block", rec.Header().Get(echo.HeaderXXSSProtection))
	assert.Equal(t, "nosniff", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "DENY", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Contains(t, rec.Header().Get(echo.HeaderStrictTransportSecurity), "max-age=31536000")
	assert.Equal(t, "no-referrer", rec.Header().Get(echo.HeaderReferrerPolicy))
}

func TestRateLimitVariants(t *testing.T) {
	variants := []struct {
		name string
		mw   echo.MiddlewareFunc
	}{
		{name: "default", mw: DefaultRateLimit()},
		{name: "strict", mw: StrictRateLimit()},
		{name: "auth", mw: AuthRateLimit()},
		{name: "public", mw: PublicRateLimit()},
		{name: "custom", mw: RateLimit(config.Config{RateLimit: 2, RateLimitDuration: time.Second})},
	}

	for _, tt := range variants {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(tt.mw)
			e.GET("/limited", func(c *echo.Context) error {
				return c.NoContent(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/limited", nil)
			req.RemoteAddr = "127.0.0.1:1234"
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusNoContent, rec.Code)
		})
	}

	t.Run("deny when exceeded", func(t *testing.T) {
		e := echo.New()
		e.Use(RateLimit(config.Config{RateLimit: 1, RateLimitDuration: time.Minute}))
		e.GET("/limited", func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})

		var lastCode int
		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/limited", nil)
			req.RemoteAddr = "10.0.0.1:9999"
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			lastCode = rec.Code
		}
		assert.Equal(t, http.StatusTooManyRequests, lastCode)
	})
}

func TestRequestContext(t *testing.T) {
	tests := []struct {
		name            string
		headers         map[string]string
		wantLang        string
		wantCorrPrefix  string
		wantExistingCorr string
	}{
		{
			name:           "generates correlation id and default language",
			wantLang:       "en",
			wantCorrPrefix: "svc-",
		},
		{
			name: "uses provided correlation and language",
			headers: map[string]string{
				CorrelationIDHeaderKey: "corr-fixed",
				LanguageCodeHeaderKey:  "fr-FR,en;q=0.8",
			},
			wantLang:         "fr-FR",
			wantExistingCorr: "corr-fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(RequestContext("svc"))
			var gotCorr, gotLang string
			e.GET("/ctx", func(c *echo.Context) error {
				gotCorr, _ = request.CorrelationIDFromContext(c.Request().Context())
				gotLang, _ = request.LanguageCodeFromContext(c.Request().Context())
				return c.NoContent(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/ctx?x=1", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, tt.wantLang, gotLang)
			if tt.wantExistingCorr != "" {
				assert.Equal(t, tt.wantExistingCorr, gotCorr)
				assert.Equal(t, tt.wantExistingCorr, rec.Header().Get(CorrelationIDHeaderKey))
			} else {
				assert.True(t, strings.HasPrefix(gotCorr, tt.wantCorrPrefix))
				assert.Equal(t, gotCorr, rec.Header().Get(CorrelationIDHeaderKey))
			}
		})
	}
}

func TestParseLanguageCode(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", parseLanguageCode(""))
	assert.Equal(t, "en", parseLanguageCode("en"))
	assert.Equal(t, "vi", parseLanguageCode("vi,en;q=0.9"))
}

func TestFormatRandomSuffix(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "005", formatRandomSuffix(5))
	assert.Equal(t, "042", formatRandomSuffix(42))
	assert.Equal(t, "999", formatRandomSuffix(999))
	assert.Equal(t, "007", formatRandomSuffix(-7))
	assert.Equal(t, "000", formatRandomSuffix(0))
}

func TestLogBodyMiddleware(t *testing.T) {
	t.Run("captures body", func(t *testing.T) {
		e := echo.New()
		var captured any
		e.Use(LogBodyMiddleware)
		e.POST("/body", func(c *echo.Context) error {
			captured = c.Get("log_body")
			return c.NoContent(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodPost, "/body", strings.NewReader(`{"a":1}`))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, `{"a":1}`, captured)
	})

	t.Run("read error", func(t *testing.T) {
		e := echo.New()
		handler := LogBodyMiddleware(func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodPost, "/body", errReader{})
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		err := handler(c)
		require.Error(t, err)
	})
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestRequestLogging(t *testing.T) {
	cfg := &config.Config{
		AppEnv:           config.EnvironmentTest,
		AppVersion:       "1.2.3",
		KeycloakKeyClaim: "claims",
	}

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			payload := base64.RawStdEncoding.EncodeToString([]byte(`{"sub":"u1"}`))
			c.Request().Header.Set("Authorization", "Bearer hdr."+payload+".sig")
			c.Set(cfg.KeycloakKeyClaim, &auth.TokenClaims{Sub: "user-42"})
			c.Set("organization_id", "org-1")
			c.Set("log_body", `{"hello":"world"}`)
			ctx := request.NewCorrelationIDContext(c.Request().Context(), "corr-1")
			ctx = request.NewLanguageCodeContext(ctx, "en")
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	e.Use(RequestLogging(cfg))
	e.GET("/logged/:id", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/logged/abc?q=1", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Custom", "value")
	req.Header.Set("Cookie", "secret=1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Cover branch where body/claims/org are absent and auth header is not JWT-shaped.
	e2 := echo.New()
	e2.Use(RequestLogging(cfg))
	e2.GET("/plain", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	req2 := httptest.NewRequest(http.MethodGet, "/plain", nil)
	req2.Header.Set("Authorization", "Bearer not-a-jwt")
	rec2 := httptest.NewRecorder()
	e2.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusNoContent, rec2.Code)
}

var _ io.Reader = errReader{}
