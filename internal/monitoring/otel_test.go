package monitoring

import (
	"os"
	"testing"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}

func TestNewOTelResource_SchemaURLMatchesSDKDefault(t *testing.T) {
	t.Parallel()

	res, err := newOTelResource(config.Config{
		AppName:    "golang-boilerplate",
		AppVersion: "1.0.0",
		AppEnv:     config.EnvironmentDevelopment,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, resource.Default().SchemaURL(), res.SchemaURL())
	assert.Equal(t, semconv.SchemaURL, res.SchemaURL())
}

func TestNewOTelResource_ServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  config.Config
		want string
	}{
		{
			name: "uses OTelServiceName when set",
			cfg: config.Config{
				OTelServiceName: "custom-service",
				AppName:         "golang-boilerplate",
			},
			want: "custom-service",
		},
		{
			name: "falls back to AppName",
			cfg: config.Config{
				AppName: "from-app-name",
			},
			want: "from-app-name",
		},
		{
			name: "falls back to default when names empty",
			cfg:  config.Config{},
			want: "golang-boilerplate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			res, err := newOTelResource(tt.cfg)
			require.NoError(t, err)

			val, ok := res.Set().Value(semconv.ServiceNameKey)
			require.True(t, ok)
			assert.Equal(t, tt.want, val.AsString())
		})
	}
}

func TestInitOpenTelemetry_SkipsWhenEndpointEmpty(t *testing.T) {
	t.Parallel()

	provider, err := InitOpenTelemetry(config.Config{})
	require.NoError(t, err)
	assert.Nil(t, provider)
}

func TestIsOTelEnabled(t *testing.T) {
	t.Parallel()

	assert.False(t, IsOTelEnabled(config.Config{}))
	assert.False(t, IsOTelEnabled(config.Config{OTelExporterEndpoint: "   "}))
	assert.True(t, IsOTelEnabled(config.Config{OTelExporterEndpoint: "localhost:4317"}))
}
