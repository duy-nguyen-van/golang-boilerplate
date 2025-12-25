package monitoring

import (
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/ory/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NRCore is a zap core that sends logs to NewRelic
type NRCore struct {
	app    *newrelic.Application
	levels []zapcore.Level
}

// NewNRCore creates a new zap core for NewRelic
func NewNRCore(app *newrelic.Application) *NRCore {
	return &NRCore{
		app:    app,
		levels: []zapcore.Level{zapcore.DebugLevel, zapcore.InfoLevel, zapcore.WarnLevel, zapcore.ErrorLevel, zapcore.FatalLevel, zapcore.PanicLevel},
	}
}

// Enabled returns whether the core should process this entry
func (c *NRCore) Enabled(level zapcore.Level) bool {
	if c.app == nil {
		return false
	}
	for _, l := range c.levels {
		if l == level {
			return true
		}
	}
	return false
}

// With adds structured context to the core
func (c *NRCore) With(fields []zap.Field) zapcore.Core {
	return c
}

// Check determines whether the supplied entry should be logged
func (c *NRCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

// Write sends the log entry to NewRelic
func (c *NRCore) Write(entry zapcore.Entry, fields []zap.Field) error {
	if c.app == nil {
		return nil
	}

	// Map zap levels to NewRelic severity levels
	severity := "INFO"
	switch entry.Level {
	case zapcore.DebugLevel:
		severity = "DEBUG"
	case zapcore.InfoLevel:
		severity = "INFO"
	case zapcore.WarnLevel:
		severity = "WARN"
	case zapcore.ErrorLevel:
		severity = "ERROR"
	case zapcore.FatalLevel, zapcore.PanicLevel:
		severity = "CRITICAL"
	}

	// Create a map of attributes from fields
	attributes := make(map[string]interface{})
	for _, field := range fields {
		// Convert latency to microseconds if present
		if field.Key == "latency" {
			if duration, ok := field.Interface.(time.Duration); ok {
				attributes[field.Key] = duration.Microseconds()
				continue
			}
		}
		attributes[field.Key] = field.Interface
	}

	// Add standard fields
	attributes["app"] = viper.GetString("NEWRELIC_APP_NAME")
	attributes["level"] = entry.Level.String()

	// Record the log to NewRelic
	c.app.RecordLog(newrelic.LogData{
		Timestamp:  entry.Time.UnixMilli(),
		Severity:   severity,
		Message:    entry.Message,
		Attributes: attributes,
	})

	return nil
}

// Sync flushes any buffered logs
func (c *NRCore) Sync() error {
	return nil
}
