package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironment_String(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "development", EnvironmentDevelopment.String())
	assert.Equal(t, "production", EnvironmentProduction.String())
}

func TestEnvironment_IsDevelopment(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		env  Environment
		want bool
	}{
		{name: "development", env: EnvironmentDevelopment, want: true},
		{name: "test", env: EnvironmentTest, want: true},
		{name: "staging", env: EnvironmentStaging, want: false},
		{name: "production", env: EnvironmentProduction, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.env.IsDevelopment())
		})
	}
}

func TestEnvironment_IsProduction(t *testing.T) {
	t.Parallel()
	assert.True(t, EnvironmentProduction.IsProduction())
	assert.False(t, EnvironmentDevelopment.IsProduction())
}

func TestGetEnvHelpers(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "getEnv uses value when set",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV", "hello")
				assert.Equal(t, "hello", getEnv("TEST_GET_ENV", "fallback"))
			},
		},
		{
			name: "getEnv uses fallback when empty",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV_EMPTY", "")
				_ = os.Unsetenv("TEST_GET_ENV_EMPTY")
				assert.Equal(t, "fallback", getEnv("TEST_GET_ENV_EMPTY", "fallback"))
			},
		},
		{
			name: "getEnvAsInt valid",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV_INT", "42")
				assert.Equal(t, 42, getEnvAsInt("TEST_GET_ENV_INT", 1))
			},
		},
		{
			name: "getEnvAsInt invalid falls back",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV_INT_BAD", "not-a-number")
				assert.Equal(t, 7, getEnvAsInt("TEST_GET_ENV_INT_BAD", 7))
			},
		},
		{
			name: "getEnvAsInt empty falls back",
			run: func(t *testing.T) {
				_ = os.Unsetenv("TEST_GET_ENV_INT_MISSING")
				assert.Equal(t, 9, getEnvAsInt("TEST_GET_ENV_INT_MISSING", 9))
			},
		},
		{
			name: "getEnvAsBool valid",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV_BOOL", "true")
				assert.True(t, getEnvAsBool("TEST_GET_ENV_BOOL", false))
			},
		},
		{
			name: "getEnvAsBool invalid falls back",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV_BOOL_BAD", "not-bool")
				assert.False(t, getEnvAsBool("TEST_GET_ENV_BOOL_BAD", false))
			},
		},
		{
			name: "getEnvAsBool empty falls back",
			run: func(t *testing.T) {
				_ = os.Unsetenv("TEST_GET_ENV_BOOL_MISSING")
				assert.True(t, getEnvAsBool("TEST_GET_ENV_BOOL_MISSING", true))
			},
		},
		{
			name: "getEnvAsDuration valid",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV_DUR", "2s")
				assert.Equal(t, 2*time.Second, getEnvAsDuration("TEST_GET_ENV_DUR", time.Second))
			},
		},
		{
			name: "getEnvAsDuration invalid falls back",
			run: func(t *testing.T) {
				t.Setenv("TEST_GET_ENV_DUR_BAD", "not-a-duration")
				assert.Equal(t, 3*time.Second, getEnvAsDuration("TEST_GET_ENV_DUR_BAD", 3*time.Second))
			},
		},
		{
			name: "getEnvAsDuration empty falls back",
			run: func(t *testing.T) {
				_ = os.Unsetenv("TEST_GET_ENV_DUR_MISSING")
				assert.Equal(t, 4*time.Second, getEnvAsDuration("TEST_GET_ENV_DUR_MISSING", 4*time.Second))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestLoad_DefaultsAndOverrides(t *testing.T) {
	// Isolate from repo .env (godotenv.Load looks in cwd) and ambient process env.
	t.Chdir(t.TempDir())

	keys := []string{
		"APP_ENV", "APP_NAME", "APP_VERSION", "APP_REQUEST_TIMEOUT",
		"DATABASE_DEBUG", "DATABASE_MAX_OPEN_CONNS", "DATABASE_CONN_MAX_LIFETIME",
		"RATE_LIMIT", "RATE_LIMIT_DURATION", "OTEL_TRACES_ENABLED",
		"HTTP_CLIENT_TLS_INSECURE_SKIP_TLS", "BASIC_AUTH_USER",
		"OTEL_SERVICE_NAME", "NEWRELIC_APP_NAME",
	}
	for _, k := range keys {
		t.Setenv(k, "")
		require.NoError(t, os.Unsetenv(k))
	}

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, EnvironmentDevelopment, cfg.AppEnv)
	assert.Equal(t, "1.0.0", cfg.AppVersion)
	assert.Equal(t, 30, cfg.AppRequestTimeout)
	assert.False(t, cfg.DatabaseEnableDebug)
	assert.Equal(t, 25, cfg.DatabaseMaxOpenConns)
	assert.Equal(t, 5*time.Minute, cfg.DatabaseConnMaxLifetime)
	assert.Equal(t, 20, cfg.RateLimit)
	assert.Equal(t, time.Second, cfg.RateLimitDuration)
	assert.True(t, cfg.OTelTracesEnabled)
	assert.False(t, cfg.HTTPClientTLSInsecureSkipTLS)

	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("APP_VERSION", "9.9.9")
	t.Setenv("APP_REQUEST_TIMEOUT", "15")
	t.Setenv("DATABASE_DEBUG", "true")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "50")
	t.Setenv("DATABASE_CONN_MAX_LIFETIME", "10m")
	t.Setenv("RATE_LIMIT", "100")
	t.Setenv("RATE_LIMIT_DURATION", "2s")
	t.Setenv("OTEL_TRACES_ENABLED", "false")
	t.Setenv("HTTP_CLIENT_TLS_INSECURE_SKIP_TLS", "1")
	t.Setenv("BASIC_AUTH_USER", "admin")
	// Ensure OTel service name falls back to APP_NAME, not a leftover OTEL_SERVICE_NAME.
	require.NoError(t, os.Unsetenv("OTEL_SERVICE_NAME"))

	cfg2, err := Load()
	require.NoError(t, err)
	assert.Equal(t, EnvironmentProduction, cfg2.AppEnv)
	assert.Equal(t, "test-app", cfg2.AppName)
	assert.Equal(t, "9.9.9", cfg2.AppVersion)
	assert.Equal(t, 15, cfg2.AppRequestTimeout)
	assert.True(t, cfg2.DatabaseEnableDebug)
	assert.Equal(t, 50, cfg2.DatabaseMaxOpenConns)
	assert.Equal(t, 10*time.Minute, cfg2.DatabaseConnMaxLifetime)
	assert.Equal(t, 100, cfg2.RateLimit)
	assert.Equal(t, 2*time.Second, cfg2.RateLimitDuration)
	assert.False(t, cfg2.OTelTracesEnabled)
	assert.True(t, cfg2.HTTPClientTLSInsecureSkipTLS)
	assert.Equal(t, "admin", cfg2.BasicAuthUsername)
	assert.Equal(t, "test-app", cfg2.OTelServiceName) // falls back via APP_NAME
}

func TestConfig_ConnectionString(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		DatabaseHost:           "db.local",
		DatabasePort:           "5432",
		DatabaseUsername:       "user",
		DatabasePassword:       "secret",
		DatabaseName:           "app",
		DatabaseSSLMode:        "require",
		DatabaseTimezone:       "UTC",
		DatabaseConnectTimeout: 30 * time.Second,
	}
	got := cfg.ConnectionString()
	assert.Contains(t, got, "host=db.local")
	assert.Contains(t, got, "port=5432")
	assert.Contains(t, got, "user=user")
	assert.Contains(t, got, "password=secret")
	assert.Contains(t, got, "dbname=app")
	assert.Contains(t, got, "sslmode=require")
	assert.Contains(t, got, "timezone=UTC")
	assert.Contains(t, got, "connect_timeout=30")
}

func TestConfig_IsDebugMode(t *testing.T) {
	t.Parallel()
	assert.True(t, (&Config{DatabaseEnableDebug: true}).IsDebugMode())
	assert.False(t, (&Config{DatabaseEnableDebug: false}).IsDebugMode())
}

func TestConfig_PopulateFromJSON(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		cfg := &Config{}
		err := cfg.PopulateFromJSON(filepath.Join(t.TempDir(), "missing.json"))
		require.Error(t, err)
	})

	t.Run("valid file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "sa.json")
		require.NoError(t, os.WriteFile(path, []byte(`{
			"client_email": "svc@example.com",
			"private_key": "-----BEGIN PRIVATE KEY-----\nABC\n-----END PRIVATE KEY-----\n",
			"project_id": "proj"
		}`), 0o600))
		cfg := &Config{}
		require.NoError(t, cfg.PopulateFromJSON(path))
	})

	t.Run("invalid json bytes", func(t *testing.T) {
		cfg := &Config{}
		err := cfg.PopulateFromJSONBytes([]byte(`{not-json`))
		require.Error(t, err)
	})

	t.Run("valid json bytes", func(t *testing.T) {
		cfg := &Config{}
		require.NoError(t, cfg.PopulateFromJSONBytes([]byte(`{"client_email":"a@b.com","private_key":"k","project_id":"p"}`)))
	})
}
