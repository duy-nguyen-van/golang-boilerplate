package monitoring

import (
	"context"
	"os"
	"testing"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}

func restoreOTelGlobals(t *testing.T) {
	t.Helper()
	tp := otel.GetTracerProvider()
	mp := otel.GetMeterProvider()
	lp := global.GetLoggerProvider()
	prop := otel.GetTextMapPropagator()
	t.Cleanup(func() {
		otel.SetTracerProvider(tp)
		otel.SetMeterProvider(mp)
		global.SetLoggerProvider(lp)
		otel.SetTextMapPropagator(prop)
	})
}

func shutdownQuick(t *testing.T, provider *OTelProvider) {
	t.Helper()
	if provider == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_ = provider.Shutdown(ctx)
}

func enabledOTelConfig(protocol string, insecure bool) config.Config {
	return config.Config{
		AppName:              "golang-boilerplate",
		AppVersion:           "1.0.0",
		AppEnv:               config.EnvironmentTest,
		OTelExporterEndpoint: "http://127.0.0.1:9",
		OTelExporterProtocol: protocol,
		OTelExporterInsecure: insecure,
		OTelTracesEnabled:    true,
		OTelMetricsEnabled:   true,
		OTelLogsEnabled:      true,
	}
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

func TestInitOpenTelemetry_AllSignalsHTTPAndGRPC(t *testing.T) {
	restoreOTelGlobals(t)

	tests := []struct {
		name     string
		protocol string
		insecure bool
	}{
		{name: "http insecure", protocol: "http/protobuf", insecure: true},
		{name: "http secure", protocol: "http", insecure: false},
		{name: "grpc insecure", protocol: "grpc", insecure: true},
		{name: "grpc secure", protocol: "grpc", insecure: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := InitOpenTelemetry(enabledOTelConfig(tt.protocol, tt.insecure))
			require.NoError(t, err)
			require.NotNil(t, provider)
			require.NotNil(t, provider.LoggerProvider())
			shutdownQuick(t, provider)
		})
	}
}

func TestInitOpenTelemetry_SignalCombinations(t *testing.T) {
	restoreOTelGlobals(t)

	tests := []struct {
		name    string
		traces  bool
		metrics bool
		logs    bool
	}{
		{name: "none", traces: false, metrics: false, logs: false},
		{name: "traces only", traces: true},
		{name: "metrics only", metrics: true},
		{name: "logs only", logs: true},
		{name: "traces and metrics", traces: true, metrics: true},
		{name: "traces and logs", traces: true, logs: true},
		{name: "metrics and logs", metrics: true, logs: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := enabledOTelConfig("http/protobuf", true)
			cfg.OTelTracesEnabled = tt.traces
			cfg.OTelMetricsEnabled = tt.metrics
			cfg.OTelLogsEnabled = tt.logs

			provider, err := InitOpenTelemetry(cfg)
			require.NoError(t, err)
			require.NotNil(t, provider)
			if tt.logs {
				assert.NotNil(t, provider.LoggerProvider())
			} else {
				assert.Nil(t, provider.LoggerProvider())
			}
			shutdownQuick(t, provider)
		})
	}
}

func TestLoggerProvider(t *testing.T) {
	t.Parallel()

	var nilProvider *OTelProvider
	assert.Nil(t, nilProvider.LoggerProvider())
	assert.Nil(t, (&OTelProvider{}).LoggerProvider())
}

func TestOTelProvider_Shutdown(t *testing.T) {
	restoreOTelGlobals(t)

	t.Run("nil provider", func(t *testing.T) {
		var provider *OTelProvider
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	t.Run("empty provider", func(t *testing.T) {
		require.NoError(t, (&OTelProvider{}).Shutdown(context.Background()))
	})

	t.Run("initialized then second shutdown returns error", func(t *testing.T) {
		provider := &OTelProvider{
			tracerProvider: sdktrace.NewTracerProvider(),
			meterProvider:  metric.NewMeterProvider(),
			loggerProvider: sdklog.NewLoggerProvider(),
		}
		require.NoError(t, provider.Shutdown(context.Background()))
		require.Error(t, provider.Shutdown(context.Background()))
	})

	t.Run("logger exporter flush error", func(t *testing.T) {
		cfg := enabledOTelConfig("http/protobuf", true)
		cfg.OTelTracesEnabled = false
		cfg.OTelMetricsEnabled = false
		provider, err := InitOpenTelemetry(cfg)
		require.NoError(t, err)
		require.NotNil(t, provider)

		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(time.Millisecond)
		require.Error(t, provider.Shutdown(ctx))
	})
}

func TestIsOTelEnabled(t *testing.T) {
	t.Parallel()

	assert.False(t, IsOTelEnabled(config.Config{}))
	assert.False(t, IsOTelEnabled(config.Config{OTelExporterEndpoint: "   "}))
	assert.True(t, IsOTelEnabled(config.Config{OTelExporterEndpoint: "localhost:4317"}))
}

func TestUsesHTTPExporter(t *testing.T) {
	t.Parallel()

	assert.False(t, usesHTTPExporter(""))
	assert.False(t, usesHTTPExporter("grpc"))
	assert.True(t, usesHTTPExporter("http"))
	assert.True(t, usesHTTPExporter("HTTP"))
	assert.True(t, usesHTTPExporter("http/protobuf"))
	assert.True(t, usesHTTPExporter("https"))
}

func TestNormalizeGRPCEndpoint(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "localhost:4317", normalizeGRPCEndpoint("  http://localhost:4317/  "))
	assert.Equal(t, "collector:4317", normalizeGRPCEndpoint("https://collector:4317"))
	assert.Equal(t, "localhost:4317", normalizeGRPCEndpoint("localhost:4317"))
}

func TestNewExporters(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tests := []struct {
		name string
		cfg  config.Config
	}{
		{name: "http insecure", cfg: enabledOTelConfig("http/protobuf", true)},
		{name: "http secure", cfg: enabledOTelConfig("https", false)},
		{name: "grpc insecure", cfg: enabledOTelConfig("grpc", true)},
		{name: "grpc secure", cfg: enabledOTelConfig("grpc", false)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traceExp, err := newTraceExporter(ctx, tt.cfg)
			require.NoError(t, err)
			require.NotNil(t, traceExp)
			_ = traceExp.Shutdown(ctx)

			metricExp, err := newMetricExporter(ctx, tt.cfg)
			require.NoError(t, err)
			require.NotNil(t, metricExp)
			_ = metricExp.Shutdown(ctx)

			logExp, err := newLogExporter(ctx, tt.cfg)
			require.NoError(t, err)
			require.NotNil(t, logExp)
			_ = logExp.Shutdown(ctx)
		})
	}
}
