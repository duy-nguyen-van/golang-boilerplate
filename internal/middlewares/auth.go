package middlewares

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/integration/auth"
	"golang-boilerplate/internal/monitoring"

	"golang-boilerplate/internal/logger"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthMiddleware creates middleware for JWT authentication
func AuthMiddleware(cfg *config.Config, authService auth.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header required",
				})
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid authorization header format",
				})
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Token required",
				})
			}

			// Validate token
			user, err := authService.ValidateToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token",
				})
			}

			if !*user.Active {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Token is not active",
				})
			}

			// Parse claims into our TokenClaims struct
			var tokenClaims auth.TokenClaims
			_, err = authService.DecodeAccessToken(context.Background(), token, authService.GetRealm(), &tokenClaims)
			if err != nil {
				// Capture invalid claims error in Sentry
				if hub := monitoring.GetSentryHub(c.Request().Context()); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						scope.SetTag("auth_error", "invalid_claims")
						scope.SetTag("service", "fast-ai")
						scope.SetTag("environment", cfg.AppEnv.String())
						scope.SetExtra("path", c.Request().URL.Path)
						scope.SetExtra("method", c.Request().Method)
						scope.SetExtra("ip", c.RealIP())
						scope.SetExtra("user_agent", c.Request().UserAgent())
						scope.SetExtra("error_details", err.Error())
						hub.CaptureException(err)
					})
				}
				logger.Log.Error("Invalid token claims",
					zap.String("path", c.Request().URL.Path),
					zap.String("method", c.Request().Method),
					zap.String("ip", c.RealIP()),
					zap.String("user_agent", c.Request().UserAgent()),
					zap.Error(err),
				)

				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token claims",
				})
			}

			// Store claims in context
			c.Set(authService.GetClaimsKey(), &tokenClaims)

			return next(c)
		}
	}
}

// RequireRole creates middleware that requires specific roles
// It checks both realm-level roles and client-level roles from TokenClaims
func RequireRole(cfg *config.Config, roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := c.Get(cfg.KeycloakKeyClaim).(*auth.TokenClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "User not authenticated",
				})
			}

			// Extract all roles from TokenClaims
			userRoles := extractRolesFromClaims(claims, cfg.KeycloakClientID)

			// Check if user has any of the required roles
			hasRole := false
			for _, requiredRole := range roles {
				if slices.Contains(userRoles, requiredRole) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
			}

			return next(c)
		}
	}
}

// extractRolesFromClaims extracts all roles from TokenClaims (realm + client roles)
func extractRolesFromClaims(claims *auth.TokenClaims, clientID string) []string {
	var roles []string

	// Add realm-level roles
	roles = append(roles, claims.RealmAccess.Roles...)

	// Add client-level roles for the specific client
	if clientRoles, exists := claims.ResourceAccess[clientID]; exists {
		roles = append(roles, clientRoles.Roles...)
	}

	return roles
}

// RequireUMA enforces resource/scope via Keycloak Authorization Services (RPT)
// It exchanges the user's access token for an RPT and checks authorization.permissions.
func RequirePermission(cfg *config.Config, authService auth.AuthService, resource string, scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract access token
			authHeader := c.Request().Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid authorization header format"})
			}
			accessToken := strings.TrimPrefix(authHeader, "Bearer ")
			if accessToken == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Token required"})
			}

			// Request RPT for the desired permission: resource#scope
			perm := resource + "#" + scope
			rpt, err := authService.GetRequestingPartyToken(c.Request().Context(), accessToken, auth.RequestingPartyTokenOptions{
				Audience:    &cfg.KeycloakClientID,
				Permissions: &[]string{perm},
			})
			if err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Permission evaluation failed"})
			}

			// Decode RPT to read authorization.permissions (use our claims)
			var rptClaims auth.TokenClaims
			_, err = authService.DecodeAccessToken(c.Request().Context(), rpt.AccessToken, authService.GetRealm(), &rptClaims)
			if err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "Invalid RPT claims"})
			}

			// Check authorization.permissions contains resource and scope
			for _, p := range rptClaims.Authorization.Permissions {
				if p.ResourceName == resource {
					for _, s := range p.Scopes {
						if s == scope {
							return next(c)
						}
					}
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{"error": "Insufficient permissions"})
		}
	}
}
