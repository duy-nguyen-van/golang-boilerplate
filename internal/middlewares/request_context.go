package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang-boilerplate/internal/request"

	"github.com/labstack/echo/v4"
)

var (
	// CorrelationIDHeaderKey is the header used to propagate correlation IDs.
	CorrelationIDHeaderKey = http.CanonicalHeaderKey("X-Correlation-Id")
	// LanguageCodeHeaderKey is the header used to propagate the preferred language code.
	LanguageCodeHeaderKey = http.CanonicalHeaderKey("Accept-Language")
)

// RequestContext middleware enriches the request context with correlation ID,
// language code, request timestamp, and request URL.
//
// The generated correlation ID is based on the service name, current time, and
// a small random component when the incoming request does not already provide one.
func RequestContext(serviceName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			correlationID := c.Request().Header.Get(CorrelationIDHeaderKey)
			if correlationID == "" {
				now := time.Now()
				randomPart := now.UnixNano() % 1000
				correlationID = strings.Join([]string{
					serviceName,
					now.Format("20060102150405"),
					formatRandomSuffix(randomPart),
				}, "-")
			}

			languageCode := parseLanguageCode(c.Request().Header.Get(LanguageCodeHeaderKey))
			if languageCode == "" {
				languageCode = "en"
			}

			timestamp := time.Now().UnixMilli()
			requestURL := c.Request().URL.String()

			ctx = request.NewCorrelationIDContext(ctx, correlationID)
			ctx = request.NewLanguageCodeContext(ctx, languageCode)
			ctx = request.NewRequestTimestampContext(ctx, timestamp)
			ctx = request.NewRequestURLContext(ctx, requestURL)

			c.Request().WithContext(ctx)
			c.SetRequest(c.Request().WithContext(ctx))

			c.Response().Header().Set(CorrelationIDHeaderKey, correlationID)

			return next(c)
		}
	}
}

// parseLanguageCode extracts the primary language code from Accept-Language.
func parseLanguageCode(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

// formatRandomSuffix pads the random suffix to 3 digits.
func formatRandomSuffix(n int64) string {
	if n < 0 {
		n = -n
	}
	n = n % 1000
	switch {
	case n < 10:
		return "00" + strconv.FormatInt(n, 10)
	case n < 100:
		return "0" + strconv.FormatInt(n, 10)
	default:
		return strconv.FormatInt(n, 10)
	}
}

