package middlewares

import (
	"net/http"

	"golang-boilerplate/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CSRF returns a configured CSRF middleware.
func CSRF(cfg *config.Config) echo.MiddlewareFunc {
	//nolint:gosec // G101: cookie name is not a secret
	return middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookiePath:     "/",
		CookieSameSite: http.SameSiteLaxMode,
		CookieSecure:   cfg.AppEnv == config.EnvironmentProduction,
		CookieHTTPOnly: true,
	})
}

// ExposeCSRFToken adds the current CSRF token to the response header for clients.
func ExposeCSRFToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if token, ok := c.Get(middleware.DefaultCSRFConfig.ContextKey).(string); ok && token != "" {
				c.Response().Header().Set("X-CSRF-Token", token)
			}
			return next(c)
		}
	}
}
