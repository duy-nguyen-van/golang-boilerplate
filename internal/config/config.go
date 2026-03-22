package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Environment represents the application environment.
type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "production"
	EnvironmentTest        Environment = "test"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) IsDevelopment() bool {
	return e == EnvironmentDevelopment || e == EnvironmentTest
}

func (e Environment) IsProduction() bool {
	return e == EnvironmentProduction
}

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	AppEnv        Environment
	AppName       string
	AppVersion    string
	Timezone      string
	AppHTTPServer string
	// AppRequestTimeout is the HTTP server read header timeout in seconds.
	AppRequestTimeout int
	AppBaseURL    string

	// Database configuration
	DatabaseHost        string
	DatabasePort        string
	DatabaseUsername    string
	DatabasePassword    string
	DatabaseName        string
	DatabaseEnableDebug bool

	// Database connection management
	DatabaseMaxOpenConns    int
	DatabaseMaxIdleConns    int
	DatabaseConnMaxLifetime time.Duration
	DatabaseConnMaxIdleTime time.Duration
	DatabaseConnectTimeout  time.Duration
	DatabaseQueryTimeout    time.Duration
	DatabaseHealthTimeout   time.Duration
	DatabaseRetryAttempts   int
	DatabaseRetryDelay      time.Duration
	DatabaseSSLMode         string
	DatabaseTimezone        string

	// Cache configuration
	CacheProvider   string
	RedisHost       string
	RedisPort       string
	RedisPassword   string
	RedisDB         int
	PoolSize        int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration

	// Logging configuration
	LogLevel string

	// Authentication configuration
	AuthProvider        string
	KeycloakURL         string
	KeycloakRealm       string
	KeycloakClientID    string
	KeycloakSecret      string
	KeycloakKeyClaim    string
	KeycloakRedirectURI string

	// Email configuration
	EmailProvider   string
	AWSSESRegion    string
	AWSSESAccessKey string
	AWSSESSecretKey string

	// Environment
	Environment string

	// Rate limiting configuration
	DefaultRateLimit  int
	AuthRateLimit     int
	PublicRateLimit   int
	RateLimit         int
	RateLimitDuration time.Duration

	// NewRelic configuration
	NewRelicAppName string
	NewRelicLicense string

	// Sentry configuration
	SentryDSN string

	// Basic Auth configuration
	BasicAuthUsername string
	BasicAuthPassword string

	// HTTP Client configuration
	HTTPClientTimeout            time.Duration
	HTTPClientRetryCount         int
	HTTPClientRetryWaitMin       time.Duration
	HTTPClientRetryWaitMax       time.Duration
	HTTPClientDebug              bool
	HTTPClientTLSInsecureSkipTLS bool

	// Cloud storage configuration
	StorageProvider           string
	GCSBucket                 string
	GCSCredentialsJSONPath    string
	GCSPresignedURLDuration   time.Duration
	GCSPresignedURLExpiration time.Duration
	S3Bucket                  string
	S3Region                  string
	S3AccessKey               string
	S3SecretKey               string
	S3PresignedURLDuration    time.Duration

	// Payment configuration
	PaymentProvider         string
	StripeSecretKey         string
	StripePublicKey         string
	StripeWebhookSecret     string
	StripeSuccessURL        string
	StripeCancelURL         string
	StripeCustomerPortalURL string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't fail if it doesn't exist
	}

	cfg := &Config{
		AppEnv:                       Environment(getEnv("APP_ENV", "development")),
		AppName:                      getEnv("APP_NAME", ""),
		AppVersion:                   getEnv("APP_VERSION", "1.0.0"),
		Timezone:                     getEnv("TIMEZONE", "UTC"),
		AppHTTPServer:                getEnv("APP_HTTP_SERVER", ":3000"),
		AppRequestTimeout:            getEnvAsInt("APP_REQUEST_TIMEOUT", 30),
		AppBaseURL:                   getEnv("APP_BASE_URL", ""),
		DatabaseHost:                 getEnv("POSTGRES_HOST", "localhost"),
		DatabasePort:                 getEnv("POSTGRES_PORT", "5432"),
		DatabaseUsername:             getEnv("POSTGRES_USER", "postgres"),
		DatabasePassword:             getEnv("POSTGRES_PASSWORD", ""),
		DatabaseName:                 getEnv("POSTGRES_DB", ""),
		DatabaseEnableDebug:          getEnvAsBool("DATABASE_DEBUG", false),
		DatabaseMaxOpenConns:         getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
		DatabaseMaxIdleConns:         getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5),
		DatabaseConnMaxLifetime:      getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		DatabaseConnMaxIdleTime:      getEnvAsDuration("DATABASE_CONN_MAX_IDLE_TIME", 1*time.Minute),
		DatabaseConnectTimeout:       getEnvAsDuration("DATABASE_CONNECT_TIMEOUT", 30*time.Second),
		DatabaseQueryTimeout:         getEnvAsDuration("DATABASE_QUERY_TIMEOUT", 30*time.Second),
		DatabaseHealthTimeout:        getEnvAsDuration("DATABASE_HEALTH_TIMEOUT", 5*time.Second),
		DatabaseRetryAttempts:        getEnvAsInt("DATABASE_RETRY_ATTEMPTS", 3),
		DatabaseRetryDelay:           getEnvAsDuration("DATABASE_RETRY_DELAY", 1*time.Second),
		DatabaseSSLMode:              getEnv("DATABASE_SSL_MODE", "disable"),
		DatabaseTimezone:             getEnv("DATABASE_TIMEZONE", "UTC"),
		CacheProvider:                getEnv("CACHE_PROVIDER", "redis"),
		RedisHost:                    getEnv("REDIS_HOST", "localhost"),
		RedisPort:                    getEnv("REDIS_PORT", "6379"),
		RedisPassword:                getEnv("REDIS_PASSWORD", ""),
		RedisDB:                      getEnvAsInt("REDIS_DB", 0),
		PoolSize:                     getEnvAsInt("REDIS_POOL_SIZE", 10),
		DialTimeout:                  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:                  getEnvAsDuration("REDIS_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:                 getEnvAsDuration("REDIS_WRITE_TIMEOUT", 5*time.Second),
		PoolTimeout:                  getEnvAsDuration("REDIS_POOL_TIMEOUT", 5*time.Second),
		MaxRetries:                   getEnvAsInt("REDIS_MAX_RETRIES", 3),
		MinRetryBackoff:              getEnvAsDuration("REDIS_MIN_RETRY_BACKOFF", 1*time.Second),
		MaxRetryBackoff:              getEnvAsDuration("REDIS_MAX_RETRY_BACKOFF", 5*time.Second),
		LogLevel:                     getEnv("LOG_LEVEL", "info"),
		AuthProvider:                 getEnv("AUTH_PROVIDER", "keycloak"),
		KeycloakURL:                  getEnv("KEYCLOAK_URL", ""),
		KeycloakRealm:                getEnv("KEYCLOAK_REALM", ""),
		KeycloakClientID:             getEnv("KEYCLOAK_CLIENT_ID", ""),
		KeycloakSecret:               getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		KeycloakKeyClaim:             getEnv("KEY_CLAIMS", ""),
		KeycloakRedirectURI:          getEnv("KEYCLOAK_REDIRECT_URI", ""),
		EmailProvider:                getEnv("EMAIL_PROVIDER", "ses"),
		AWSSESRegion:                 getEnv("AWS_SES_REGION", ""),
		AWSSESAccessKey:              getEnv("AWS_SES_ACCESS_KEY", ""),
		AWSSESSecretKey:              getEnv("AWS_SES_SECRET_KEY", ""),
		RateLimit:                    getEnvAsInt("RATE_LIMIT", 20),
		RateLimitDuration:            getEnvAsDuration("RATE_LIMIT_DURATION", 1*time.Second),
		Environment:                  getEnv("ENVIRONMENT", "development"),
		DefaultRateLimit:             getEnvAsInt("DEFAULT_RATE_LIMIT", 20),
		AuthRateLimit:                getEnvAsInt("AUTH_RATE_LIMIT", 3),
		PublicRateLimit:              getEnvAsInt("PUBLIC_RATE_LIMIT", 100),
		NewRelicAppName:              getEnv("NEWRELIC_APP_NAME", "golang-boilerplate"),
		NewRelicLicense:              getEnv("NEWRELIC_LICENSE", ""),
		SentryDSN:                    getEnv("SENTRY_DSN", ""),
		BasicAuthUsername:            getEnv("BASIC_AUTH_USER", ""),
		BasicAuthPassword:            getEnv("BASIC_AUTH_SECRET", ""),
		HTTPClientTimeout:            getEnvAsDuration("HTTP_CLIENT_TIMEOUT", 30*time.Second),
		HTTPClientRetryCount:         getEnvAsInt("HTTP_CLIENT_RETRY_COUNT", 2),
		HTTPClientRetryWaitMin:       getEnvAsDuration("HTTP_CLIENT_RETRY_WAIT_MIN", 250*time.Millisecond),
		HTTPClientRetryWaitMax:       getEnvAsDuration("HTTP_CLIENT_RETRY_WAIT_MAX", 2*time.Second),
		HTTPClientDebug:              getEnvAsBool("HTTP_CLIENT_DEBUG", false),
		HTTPClientTLSInsecureSkipTLS: getEnvAsBool("HTTP_CLIENT_TLS_INSECURE_SKIP_TLS", false),
		StorageProvider:              getEnv("STORAGE_PROVIDER", "gcs"),
		GCSBucket:                    getEnv("GCS_BUCKET", ""),
		GCSCredentialsJSONPath:       getEnv("GCS_CREDENTIALS_JSON", ""),
		GCSPresignedURLDuration:      getEnvAsDuration("GCS_PRESIGNED_URL_DURATION", 1*time.Hour),
		GCSPresignedURLExpiration:    getEnvAsDuration("GCS_PRESIGNED_URL_EXPIRATION", 1*time.Hour),
		S3Bucket:                     getEnv("S3_BUCKET", ""),
		S3Region:                     getEnv("S3_REGION", ""),
		S3AccessKey:                  getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:                  getEnv("S3_SECRET_KEY", ""),
		S3PresignedURLDuration:       getEnvAsDuration("S3_PRESIGNED_URL_DURATION", 1*time.Hour),
		PaymentProvider:              getEnv("PAYMENT_PROVIDER", "stripe"),
		StripeSecretKey:              getEnv("STRIPE_SECRET_KEY", ""),
		StripePublicKey:              getEnv("STRIPE_PUBLIC_KEY", ""),
		StripeWebhookSecret:          getEnv("STRIPE_WEBHOOK_SECRET", ""),
		StripeSuccessURL:             getEnv("STRIPE_SUCCESS_URL", ""),
		StripeCancelURL:              getEnv("STRIPE_CANCEL_URL", ""),
		StripeCustomerPortalURL:      getEnv("STRIPE_CUSTOMER_PORTAL_URL", ""),
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// getEnvAsDuration gets an environment variable as duration with a fallback value
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v sslmode=%v timezone=%v connect_timeout=%d",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseUsername,
		c.DatabasePassword,
		c.DatabaseName,
		c.DatabaseSSLMode,
		c.DatabaseTimezone,
		int(c.DatabaseConnectTimeout.Seconds()),
	)
}

func (c *Config) IsDebugMode() bool {
	return c.DatabaseEnableDebug
}

// populateFromJSON reads a service account JSON and fills email and private key
func (c *Config) PopulateFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var payload struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
		ProjectID   string `json:"project_id"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	return nil
}
