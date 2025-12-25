package main

import (
	"context"
	"errors"
	"fmt"
	"golang-boilerplate/docs"
	"golang-boilerplate/internal/cache"
	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/handlers"
	"golang-boilerplate/internal/httpclient"
	"golang-boilerplate/internal/integration/auth"
	"golang-boilerplate/internal/integration/email"
	"golang-boilerplate/internal/integration/payment"
	"golang-boilerplate/internal/integration/storage"
	"golang-boilerplate/internal/logger"
	"golang-boilerplate/internal/monitoring"
	"golang-boilerplate/internal/repositories"
	"golang-boilerplate/internal/services"
	"net"
	"net/http"
	"os"
	"time"

	"golang-boilerplate/cmd/server/routes"

	"golang-boilerplate/internal/db"

	"github.com/go-playground/validator/v10"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/ory/viper"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewHTTPServer(lc fx.Lifecycle,
	healthHandler *handlers.HealthHandler,
	userHandler *handlers.UserHandler,
	companyHandler *handlers.CompanyHandler,
	authProvider auth.AuthService,
	nrApp *newrelic.Application,
	cfg *config.Config,
	db *db.PostgresDB,
) *http.Server {
	handler := routes.Router(userHandler, companyHandler, healthHandler, authProvider, nrApp, cfg).Server.Handler

	srv := &http.Server{
		Addr: viper.GetString("APP_HTTP_SERVER"), Handler: handler,
		ReadHeaderTimeout: time.Second * time.Duration(viper.GetInt("APP_REQUEST_TIMEOUT")),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			logger.Sugar.Infof("Starting HTTP server at %s", srv.Addr)
			go func() {
				err := srv.Serve(ln)
				if err != nil {
					logger.Sugar.Panicf("HTTP server error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Sugar.Info("Shutting down HTTP server...")

			// Graceful shutdown with timeout
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			// Shutdown HTTP server
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Sugar.Errorf("HTTP server shutdown error: %v", err)
				return err
			}

			// Close database connections
			if err := db.Close(); err != nil {
				logger.Sugar.Errorf("Database shutdown error: %v", err)
				return err
			}

			logger.Sugar.Info("Server shutdown completed")
			return nil
		},
	})

	return srv
}

// @title Golang Boilerplate API
// @version 1.0
// @description This is a backend API for Golang Boilerplate
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.basic  BasicAuth
// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Bearer Token Authentication. Use "Bearer {token}" as the value.
func main() {
	InitConfig(".env")
	// Ensure Swagger spec is registered and optionally override fields at runtime
	docs.SwaggerInfo.BasePath = "/api/v1"
	cfg, err := config.Load()
	if err != nil {
		// Use standard log for fatal errors before logger is initialized
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	// Initialize global logger before any middleware uses it
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	nrApp := monitoring.InitNewRelic(*cfg)
	monitoring.InitSentry(*cfg)

	// Add Sentry core to zap logger
	if logger.Log != nil {
		logger.Log = logger.Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, monitoring.NewSentryCore(context.Background(), []zapcore.Level{
				zapcore.ErrorLevel,
				zapcore.FatalLevel,
				zapcore.PanicLevel,
			}))
		}))
		logger.Sugar = logger.Log.Sugar()
	}

	// Ensure all events are flushed before the program exits
	defer monitoring.FlushSentry()

	// Set application timezone from environment variable, default to UTC if not specified
	timezone := cfg.Timezone
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		logger.Sugar.Warnf("Invalid timezone %s, falling back to UTC", timezone)
		loc = time.UTC
	}
	time.Local = loc

	fx.New(
		fx.Supply(cfg),
		fx.Supply(nrApp),
		fx.Provide(
			NewHTTPServer,
			ProvideGormPostgres,
			ProvideValidator,
			httpclient.ProvideRestClient,
			auth.ProvideAuth,
			cache.ProvideCache,
			email.ProvideEmailSender,
			payment.ProvidePaymentAdapter,
			storage.ProvideStorageAdapter,
			repositories.ProvideUserRepository,
			repositories.ProvideCompanyRepository,
			services.ProvideCompanyService,
			services.ProvideEmailService,
			services.ProvideUserService,
			services.ProvideAuthService,
			handlers.ProvideHealthHandler,
			handlers.ProvideUserHandler,
			handlers.ProvideCompanyHandler,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func ProvideValidator() *validator.Validate {
	return validator.New()
}

func InitConfig(path string) {
	viper.AutomaticEnv()
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		// Use standard log for warnings before logger is initialized
		fmt.Fprintf(os.Stderr, "no config file '%s' not found. Using default values\n", path)
	} else if err != nil { // Handle other errors that occurred while reading the config file
		panic(fmt.Errorf("fatal error while reading the config file: %w", err))
	}
}

func ProvideGormPostgres(cfg *config.Config) *db.PostgresDB {
	appDB := &db.PostgresDB{}
	err := appDB.NewPostgresDB(cfg)
	if err != nil {
		logger.Sugar.Fatalf("Connecting to Database: %v", err)
	}
	return appDB
}
