package middlewares

import (
	"net/http"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// RateLimit creates a rate limiting middleware with custom configuration
func RateLimit(config config.Config) echo.MiddlewareFunc {
	rateLimit := rate.Limit(float64(config.RateLimit) / config.RateLimitDuration.Seconds())
	store := echoMiddleware.NewRateLimiterMemoryStoreWithConfig(
		echoMiddleware.RateLimiterMemoryStoreConfig{
			Rate:      rateLimit,
			Burst:     config.RateLimit,
			ExpiresIn: config.RateLimitDuration,
		},
	)

	return echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store: store,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusForbidden, map[string]interface{}{
				"message_code": constants.RateLimitExceeded,
				"message":      "Rate limit exceeded",
				"limit":        config.RateLimit,
				"window":       config.RateLimitDuration.String(),
			})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error_code": constants.RateLimitExceeded,
				"message":    "Rate limit exceeded",
				"limit":      config.RateLimit,
				"window":     config.RateLimitDuration.String(),
			})
		},
	})
}

// DefaultRateLimit creates a default rate limiting middleware (20 requests per second)
func DefaultRateLimit() echo.MiddlewareFunc {
	return RateLimit(config.Config{
		RateLimit:         20,
		RateLimitDuration: time.Second,
	})
}

// StrictRateLimit creates a strict rate limiting middleware (5 requests per minute)
func StrictRateLimit() echo.MiddlewareFunc {
	return RateLimit(config.Config{
		RateLimit:         5,
		RateLimitDuration: time.Minute,
	})
}

// AuthRateLimit creates a rate limiting middleware for authentication endpoints
func AuthRateLimit() echo.MiddlewareFunc {
	return RateLimit(config.Config{
		RateLimit:         3,
		RateLimitDuration: time.Minute,
	})
}

// PublicRateLimit creates a rate limiting middleware for public endpoints
func PublicRateLimit() echo.MiddlewareFunc {
	return RateLimit(config.Config{
		RateLimit:         100,
		RateLimitDuration: time.Minute,
	})
}
