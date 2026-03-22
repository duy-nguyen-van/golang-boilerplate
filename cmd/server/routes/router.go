package routes

import (
	"net/http"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/handlers"
	"golang-boilerplate/internal/integration/auth"
	middlewares "golang-boilerplate/internal/middlewares"

	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/newrelic/go-agent/v3/newrelic"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Router(
	userHandler *handlers.UserHandler,
	companyHandler *handlers.CompanyHandler,
	healthHandler *handlers.HealthHandler,
	authService auth.AuthService,
	nrApp *newrelic.Application,
	cfg *config.Config,
) *echo.Echo {
	r := echo.New()

	// Once it's done, you can attach the handler as one of your middleware
	r.Use(sentryecho.New(sentryecho.Options{
		Repanic: true,
	}))

	// Custom error handler middleware to capture errors and report to Sentry
	r.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				// Do not report missing routes / bad methods to Sentry (avoids noise when e.g. Swagger UI is not mounted)
				if he, ok := err.(*echo.HTTPError); ok && (he.Code == http.StatusNotFound || he.Code == http.StatusMethodNotAllowed) {
					return err
				}
				// Get the Sentry hub from context
				if hub := sentryecho.GetHubFromContext(c); hub != nil {
					// Capture the error with additional context
					hub.WithScope(func(scope *sentry.Scope) {
						// Add request context
						scope.SetExtra("method", c.Request().Method)
						scope.SetExtra("path", c.Request().URL.Path)
						scope.SetExtra("query", c.QueryParams())
						scope.SetExtra("headers", c.Request().Header)
						scope.SetExtra("body", c.Get("log_body"))

						// Add environment context
						scope.SetTag("environment", cfg.AppEnv.String())
						scope.SetTag("service", cfg.AppName)
						scope.SetTag("handler", c.Path())

						// Add organization context if available
						if orgID := c.Get("organization_id"); orgID != nil {
							scope.SetTag("organization_id", orgID.(string))
						}

						// Add error type tag
						if echoErr, ok := err.(*echo.HTTPError); ok {
							scope.SetTag("error_type", "http_error")
							scope.SetExtra("http_code", echoErr.Code)
						} else {
							scope.SetTag("error_type", "internal_error")
						}

						// Capture the error
						hub.CaptureException(err)
					})
				}
			}
			return err
		}
	})

	r.Use(middlewares.LogBodyMiddleware)
	r.Use(middleware.RequestID())
	r.Use(middlewares.RequestContext(cfg.AppName))
	r.Use(errors.RecoveryMiddleware(cfg)) // Add panic recovery
	r.Use(errors.ErrorMiddleware())       // Add centralized error handling
	r.Use(middlewares.Security())         // Add secure headers (XSS, HSTS, etc.)
	r.Use(middlewares.CORS())
	r.Use(middlewares.CSRF(cfg))
	r.Use(middlewares.ExposeCSRFToken())
	r.Use(middlewares.DefaultRateLimit())
	r.Use(middlewares.RequestLogging(cfg))

	if cfg.AppEnv != config.EnvironmentProduction {
		r.GET("/swagger/*", echoSwagger.WrapHandler, middlewares.BasicAuthMiddleware(*cfg))
	}

	// Base API path
	baseAPI := "api"

	// Version 1 API group
	v1 := r.Group(baseAPI + "/v1")

	// Public routes
	publicGroup := v1.Group("")
	publicGroup.GET("/", healthHandler.HealthCheck)
	publicGroup.GET("/health/database", healthHandler.DatabaseHealthCheck)
	publicGroup.GET("/health/metrics", healthHandler.DatabaseMetrics)

	// User routes
	userGroup := v1.Group("/users")

	userGroup.GET("", userHandler.GetUsers,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.UserViewRoles...),
	)

	userGroup.GET("/test-rest-client", userHandler.TestRestClient, middlewares.AuthMiddleware(cfg, authService))

	userGroup.POST("", userHandler.CreateUser,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserManager),
	)

	userGroup.GET("/:id", userHandler.GetOneByID,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserViewer),
	)

	userGroup.PUT("/:id", userHandler.UpdateUser,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.UserManagementRoles...),
	)

	userGroup.DELETE("/:id", userHandler.DeleteUser,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleUserManager),
	)

	// Company routes
	companyGroup := v1.Group("/companies")

	companyGroup.POST("", companyHandler.CreateCompany,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleCompanyManager, constants.RoleCompanyCreator),
	)

	companyGroup.GET("/:id", companyHandler.GetOneByID,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.CompanyViewRoles...),
	)

	companyGroup.PUT("/:id", companyHandler.UpdateCompany,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.RoleAdmin, constants.RoleCompanyManager, constants.RoleCompanyEditor),
	)

	companyGroup.DELETE("/:id", companyHandler.DeleteCompany,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.RoleAdmin),
	)

	companyGroup.GET("", companyHandler.GetCompanies,
		middlewares.AuthMiddleware(cfg, authService),
		middlewares.RequireRole(cfg, constants.CompanyViewRoles...),
	)

	return r
}
