package monitoring

import (
	"context"
	"strings"
	"time"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// IsOTelEnabled reports whether OpenTelemetry exporters are configured.
func IsOTelEnabled(cfg config.Config) bool {
	return strings.TrimSpace(cfg.OTelExporterEndpoint) != ""
}

// OTelProvider holds OpenTelemetry providers for graceful shutdown.
type OTelProvider struct {
	tracerProvider *trace.TracerProvider
	meterProvider  *metric.MeterProvider
	loggerProvider *logsdk.LoggerProvider
}

// LoggerProvider returns the OpenTelemetry logger provider for zap bridging.
func (p *OTelProvider) LoggerProvider() *logsdk.LoggerProvider {
	if p == nil {
		return nil
	}
	return p.loggerProvider
}

// InitOpenTelemetry configures tracing and metrics exporters when an OTLP endpoint is set.
func InitOpenTelemetry(cfg config.Config) (*OTelProvider, error) {
	if strings.TrimSpace(cfg.OTelExporterEndpoint) == "" {
		logger.Sugar.Warn("OpenTelemetry OTLP endpoint not provided, skipping OpenTelemetry initialization")
		return nil, nil
	}

	ctx := context.Background()

	res, err := newOTelResource(cfg)
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	provider := &OTelProvider{}

	if cfg.OTelTracesEnabled {
		traceExporter, err := newTraceExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}

		provider.tracerProvider = trace.NewTracerProvider(
			trace.WithBatcher(traceExporter),
			trace.WithResource(res),
		)
		otel.SetTracerProvider(provider.tracerProvider)
	}

	if cfg.OTelMetricsEnabled {
		metricExporter, err := newMetricExporter(ctx, cfg)
		if err != nil {
			if provider.tracerProvider != nil {
				_ = provider.tracerProvider.Shutdown(ctx)
			}
			return nil, err
		}

		provider.meterProvider = metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(metricExporter)),
			metric.WithResource(res),
		)
		otel.SetMeterProvider(provider.meterProvider)
	}

	if cfg.OTelLogsEnabled {
		logExporter, err := newLogExporter(ctx, cfg)
		if err != nil {
			if provider.meterProvider != nil {
				_ = provider.meterProvider.Shutdown(ctx)
			}
			if provider.tracerProvider != nil {
				_ = provider.tracerProvider.Shutdown(ctx)
			}
			return nil, err
		}

		provider.loggerProvider = logsdk.NewLoggerProvider(
			logsdk.WithProcessor(logsdk.NewBatchProcessor(logExporter)),
			logsdk.WithResource(res),
		)
		global.SetLoggerProvider(provider.loggerProvider)
	}

	logger.Sugar.Infof(
		"OpenTelemetry initialized (endpoint=%s, protocol=%s, traces=%t, metrics=%t, logs=%t)",
		cfg.OTelExporterEndpoint,
		cfg.OTelExporterProtocol,
		cfg.OTelTracesEnabled,
		cfg.OTelMetricsEnabled,
		cfg.OTelLogsEnabled,
	)

	return provider, nil
}

// Shutdown flushes and shuts down OpenTelemetry providers.
func (p *OTelProvider) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var shutdownErr error

	if p.loggerProvider != nil {
		if err := p.loggerProvider.Shutdown(shutdownCtx); err != nil {
			shutdownErr = err
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(shutdownCtx); err != nil {
			shutdownErr = err
		}
	}

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(shutdownCtx); err != nil {
			shutdownErr = err
		}
	}

	return shutdownErr
}

// newOTelResource builds the process resource. SchemaURL must match
// resource.Default() (semconv v1.43.0 in otel SDK v1.46); a mismatch
// returns "conflicting Schema URL".
func newOTelResource(cfg config.Config) (*resource.Resource, error) {
	serviceName := cfg.OTelServiceName
	if serviceName == "" {
		serviceName = cfg.AppName
	}
	if serviceName == "" {
		serviceName = "golang-boilerplate"
	}

	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(cfg.AppVersion),
			semconv.DeploymentEnvironmentNameKey.String(cfg.AppEnv.String()),
		),
	)
}

func usesHTTPExporter(protocol string) bool {
	return strings.HasPrefix(strings.ToLower(protocol), "http")
}

func newTraceExporter(ctx context.Context, cfg config.Config) (trace.SpanExporter, error) {
	if usesHTTPExporter(cfg.OTelExporterProtocol) {
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpointURL(cfg.OTelExporterEndpoint),
		}
		if cfg.OTelExporterInsecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(ctx, opts...)
	}

	endpoint := normalizeGRPCEndpoint(cfg.OTelExporterEndpoint)
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
	}
	if cfg.OTelExporterInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return otlptracegrpc.New(ctx, opts...)
}

func newLogExporter(ctx context.Context, cfg config.Config) (logsdk.Exporter, error) {
	if usesHTTPExporter(cfg.OTelExporterProtocol) {
		opts := []otlploghttp.Option{
			otlploghttp.WithEndpointURL(cfg.OTelExporterEndpoint),
		}
		if cfg.OTelExporterInsecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		return otlploghttp.New(ctx, opts...)
	}

	endpoint := normalizeGRPCEndpoint(cfg.OTelExporterEndpoint)
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint),
	}
	if cfg.OTelExporterInsecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	return otlploggrpc.New(ctx, opts...)
}

func newMetricExporter(ctx context.Context, cfg config.Config) (metric.Exporter, error) {
	if usesHTTPExporter(cfg.OTelExporterProtocol) {
		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpointURL(cfg.OTelExporterEndpoint),
		}
		if cfg.OTelExporterInsecure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		return otlpmetrichttp.New(ctx, opts...)
	}

	endpoint := normalizeGRPCEndpoint(cfg.OTelExporterEndpoint)
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
	}
	if cfg.OTelExporterInsecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	return otlpmetricgrpc.New(ctx, opts...)
}

func normalizeGRPCEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	return strings.TrimSuffix(endpoint, "/")
}
