package middlewares

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/integration/auth"

	"golang-boilerplate/internal/logger"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ory/viper"
	"go.uber.org/zap"
)

// RequestLogging provides structured logs to Sentry and zap
// and mirrors the rich request logging previously configured in the router.
func RequestLogging(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogUserAgent: true,
		LogMethod:    true,
		LogRemoteIP:  true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			// Create Sentry logger with request context
			sentryLogger := sentry.NewLogger(c.Request().Context())
			stdLogger := log.New(sentryLogger, "", log.LstdFlags)

			// Extract user info from token if available
			userLoginInfo := ""
			userID := ""
			organizationID := ""
			if len(c.Request().Header["Authorization"]) > 0 {
				tmp := strings.Split(c.Request().Header["Authorization"][0], ".")
				if len(tmp) == 3 {
					sDesc, _ := base64.RawStdEncoding.DecodeString(tmp[1])
					userLoginInfo = string(sDesc)
				}
			}

			// Get user and organization IDs from context if available
			if u := c.Get(cfg.KeycloakKeyClaim); u != nil {
				if claims, ok := u.(*auth.TokenClaims); ok {
					userID = claims.Sub
				}
			}
			if org := c.Get("organization_id"); org != nil {
				organizationID = org.(string)
			}

			// Get request body (captured by upstream middleware)
			body := c.Get("log_body")
			if body == nil {
				body = ""
			}

			// Get query parameters
			query := echo.Map{}
			_ = (&echo.DefaultBinder{}).BindQueryParams(c, &query)
			jsonQueryStr, _ := json.Marshal(query)

			// Get path parameters
			param := echo.Map{}
			_ = (&echo.DefaultBinder{}).BindPathParams(c, &param)
			jsonParamStr, _ := json.Marshal(param)

			// Get request headers (excluding sensitive ones)
			headers := make(map[string]string)
			for k, v := range c.Request().Header {
				// Skip sensitive headers
				if strings.EqualFold(k, "Authorization") ||
					strings.EqualFold(k, "Cookie") ||
					strings.EqualFold(k, "Set-Cookie") {
					continue
				}
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}
			jsonHeadersStr, _ := json.Marshal(headers)

			// Format log message with all fields
			logMsg := fmt.Sprintf("Request: %s %s (status=%d, latency=%v, request_id=%s, remote_ip=%s, user_agent=%s, user_id=%s, org_id=%s, payload=%s, query=%s, path_params=%s, headers=%s, environment=%s, service=fast-ai, version=%s, timestamp=%d, hostname=%s, protocol=%s)",
				values.Method,
				values.URI,
				values.Status,
				values.Latency,
				values.RequestID,
				values.RemoteIP,
				values.UserAgent,
				userID,
				organizationID,
				body,
				string(jsonQueryStr),
				string(jsonParamStr),
				string(jsonHeadersStr),
				viper.GetString("APP_ENV"),
				viper.GetString("APP_VERSION"),
				time.Now().UnixMilli(),
				c.Request().Host,
				c.Request().Proto,
			)

			// Log to Sentry
			stdLogger.Println(logMsg)

			// Also log to zap for structured logging
			logger.Log.Info("Request: "+values.Method+" "+values.URI,
				zap.String("uri", values.URI),
				zap.String("method", values.Method),
				zap.Int("status", values.Status),
				zap.Duration("latency", values.Latency),
				zap.String("request_id", values.RequestID),
				zap.String("remote_ip", values.RemoteIP),
				zap.String("user_agent", values.UserAgent),
				zap.String("user_id", userID),
				zap.String("user_login", userLoginInfo),
				zap.String("org_id", organizationID),
				zap.String("payload", fmt.Sprintf("%v", body)),
				zap.String("query", string(jsonQueryStr)),
				zap.String("path_params", string(jsonParamStr)),
				zap.String("headers", string(jsonHeadersStr)),
				zap.String("environment", viper.GetString("APP_ENV")),
				zap.String("service", "fast-ai"),
				zap.String("version", viper.GetString("APP_VERSION")),
				zap.Int64("timestamp", time.Now().UnixMilli()),
				zap.String("hostname", c.Request().Host),
				zap.String("protocol", c.Request().Proto),
			)

			return nil
		},
	})
}

func LogBodyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		data, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		c.Request().Body = io.NopCloser(bytes.NewReader(data))
		c.Set("log_body", string(data))
		return next(c)
	}
}
