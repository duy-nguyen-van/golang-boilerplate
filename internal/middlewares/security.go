package middlewares

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// Security returns a middleware that sets common security headers to mitigate XSS and related attacks.
// It configures X-XSS-Protection, X-Content-Type-Options, X-Frame-Options, HSTS, and Referrer-Policy.
// Note: We intentionally avoid setting a strict Content-Security-Policy here to prevent breaking Swagger UI.
func Security() echo.MiddlewareFunc {
	return middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "1; mode=block", // Legacy header, still helpful for older browsers
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		HSTSMaxAge:         31536000, // 1 year
		// HSTSExcludeSubdomains: false, // include subdomains by default
		HSTSPreloadEnabled: true,
		ReferrerPolicy:     "no-referrer",
		// Leave ContentSecurityPolicy empty to avoid interfering with Swagger and any embedded UI
	})
}
