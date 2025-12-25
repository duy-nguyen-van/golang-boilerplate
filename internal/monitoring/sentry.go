package monitoring

import (
	"context"
	"fmt"
	"golang-boilerplate/internal/config"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SentryCore is a zap core that sends logs to Sentry
type SentryCore struct {
	ctx    context.Context
	levels []zapcore.Level
}

// NewSentryCore creates a new zap core for Sentry
func NewSentryCore(ctx context.Context, levels []zapcore.Level) *SentryCore {
	if levels == nil {
		levels = []zapcore.Level{
			zapcore.ErrorLevel,
			zapcore.FatalLevel,
			zapcore.PanicLevel,
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return &SentryCore{
		ctx:    ctx,
		levels: levels,
	}
}

// Enabled returns whether the core should process this entry
func (c *SentryCore) Enabled(level zapcore.Level) bool {
	for _, l := range c.levels {
		if l == level {
			return true
		}
	}
	return false
}

// With adds structured context to the core
func (c *SentryCore) With(fields []zap.Field) zapcore.Core {
	return c
}

// Check determines whether the supplied entry should be logged
func (c *SentryCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

// Write sends the log entry to Sentry
func (c *SentryCore) Write(entry zapcore.Entry, fields []zap.Field) error {
	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		// Set log level
		scope.SetLevel(getSentryLevel(entry.Level))

		// Add all fields as extra data
		for _, field := range fields {
			// Special handling for error field
			if field.Key == "error" {
				if err, ok := field.Interface.(error); ok {
					scope.SetExtra("error_details", err.Error())
					hub.CaptureException(err)
					continue
				}
			}
			// Add other fields
			scope.SetExtra(field.Key, field.Interface)
		}

		// Add standard fields
		scope.SetTag("logger", "zap")
		scope.SetTag("log_level", entry.Level.String())

		// Add timestamp
		scope.SetExtra("timestamp", entry.Time.Format(time.RFC3339))

		// Add caller information if available
		if entry.Caller.Defined {
			scope.SetExtra("caller_file", entry.Caller.File)
			scope.SetExtra("caller_line", entry.Caller.Line)
			scope.SetExtra("caller_function", entry.Caller.Function)
		}
	})

	// Format the message with fields
	msg := entry.Message
	if len(fields) > 0 {
		for _, field := range fields {
			if field.Key != "error" {
				msg = msg + " " + field.Key + "=" + formatValue(field.Interface)
			}
		}
	}

	// Log the message using Sentry logger
	sentryLogger := sentry.NewLogger(c.ctx)
	stdLogger := log.New(sentryLogger, "", log.LstdFlags)
	stdLogger.Println(msg)

	return nil
}

// Sync flushes any buffered logs
func (c *SentryCore) Sync() error {
	return nil
}

func InitSentry(config config.Config) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.SentryDSN,
		Environment:      config.AppEnv,
		Release:          config.AppName + "@" + config.AppVersion,
		Debug:            config.AppEnv == "development",
		AttachStacktrace: true,
		EnableTracing:    true,
		EnableLogs:       true,
		TracesSampleRate: 1.0,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Filter out sensitive information
			if event.Request != nil {
				// Remove sensitive headers
				if event.Request.Headers != nil {
					delete(event.Request.Headers, "Authorization")
					delete(event.Request.Headers, "Cookie")
				}
			}
			return event
		},
	})

	if err != nil {
		log.Fatalf("sentry.Init failed: %v", err)
	}
}

// Flush buffered events before shutdown
func FlushSentry() {
	sentry.Flush(2 * time.Second)
}

// GetSentryHub returns a Sentry hub from the context if available, otherwise falls back to the current hub.
// This makes error reporting more robust by ensuring we always try to report errors when possible.
func GetSentryHub(ctx context.Context) *sentry.Hub {
	// Try to get hub from context first
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		// Fallback to current hub if context hub is not available
		hub = sentry.CurrentHub()
	}
	return hub
}

// getSentryLevel converts zap level to sentry level
func getSentryLevel(level zapcore.Level) sentry.Level {
	switch level {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.FatalLevel, zapcore.PanicLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelInfo
	}
}

// formatValue formats a value for logging
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case error:
		return val.Error()
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// SentryWriter implements io.Writer to write logs to Sentry
type SentryWriter struct {
	ctx context.Context
}

// NewSentryWriter creates a new writer that writes to Sentry
func NewSentryWriter(ctx context.Context) *SentryWriter {
	if ctx == nil {
		ctx = context.Background()
	}
	return &SentryWriter{ctx: ctx}
}

// Write implements io.Writer
func (w *SentryWriter) Write(p []byte) (n int, err error) {
	sentryLogger := sentry.NewLogger(w.ctx)
	logger := log.New(sentryLogger, "", log.LstdFlags)
	logger.Print(string(p))
	return len(p), nil
}
