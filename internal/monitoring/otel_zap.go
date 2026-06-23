package monitoring

import (
	"golang-boilerplate/internal/logger"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewOTelZapCore creates a zap core that exports logs to OpenTelemetry.
func NewOTelZapCore(provider log.LoggerProvider, levels []zapcore.Level) zapcore.Core {
	if provider == nil {
		return zapcore.NewNopCore()
	}
	if levels == nil {
		levels = []zapcore.Level{
			zapcore.ErrorLevel,
			zapcore.FatalLevel,
			zapcore.PanicLevel,
		}
	}

	return &otelZapLevelCore{
		core: otelzap.NewCore(
			"golang-boilerplate",
			otelzap.WithLoggerProvider(provider),
		),
		levels: levels,
	}
}

// AttachOTelZapLogger tees an OpenTelemetry zap core onto the global logger.
func AttachOTelZapLogger(provider log.LoggerProvider, levels []zapcore.Level) {
	if provider == nil || logger.Log == nil {
		return
	}

	logger.Log = logger.Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, NewOTelZapCore(provider, levels))
	}))
	logger.Sugar = logger.Log.Sugar()
}

type otelZapLevelCore struct {
	core   zapcore.Core
	levels []zapcore.Level
}

func (c *otelZapLevelCore) Enabled(level zapcore.Level) bool {
	return c.levelEnabled(level) && c.core.Enabled(level)
}

func (c *otelZapLevelCore) With(fields []zapcore.Field) zapcore.Core {
	return &otelZapLevelCore{
		core:   c.core.With(fields),
		levels: c.levels,
	}
}

func (c *otelZapLevelCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.levelEnabled(entry.Level) {
		return c.core.Check(entry, checked)
	}
	return checked
}

func (c *otelZapLevelCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return c.core.Write(entry, fields)
}

func (c *otelZapLevelCore) Sync() error {
	return c.core.Sync()
}

func (c *otelZapLevelCore) levelEnabled(level zapcore.Level) bool {
	for _, l := range c.levels {
		if l == level {
			return true
		}
	}
	return false
}
