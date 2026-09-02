package monitoring

import (
	"testing"

	"golang-boilerplate/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewOTelZapCore(t *testing.T) {
	t.Run("nil provider returns nop", func(t *testing.T) {
		core := NewOTelZapCore(nil, nil)
		require.NotNil(t, core)
		assert.False(t, core.Enabled(zapcore.ErrorLevel))
	})

	t.Run("default levels", func(t *testing.T) {
		provider := sdklog.NewLoggerProvider()
		t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

		core := NewOTelZapCore(provider, nil)
		wrapped, ok := core.(*otelZapLevelCore)
		require.True(t, ok)
		assert.True(t, wrapped.levelEnabled(zapcore.ErrorLevel))
		assert.True(t, wrapped.levelEnabled(zapcore.FatalLevel))
		assert.True(t, wrapped.levelEnabled(zapcore.PanicLevel))
		assert.False(t, wrapped.levelEnabled(zapcore.InfoLevel))

		_ = core.Enabled(zapcore.ErrorLevel)
		withFields := core.With([]zapcore.Field{zap.String("k", "v")})
		require.NotNil(t, withFields)
		_ = core.Sync()
		_ = core.Check(zapcore.Entry{Level: zapcore.ErrorLevel}, nil)
		assert.Nil(t, core.Check(zapcore.Entry{Level: zapcore.InfoLevel}, nil))
		_ = core.Write(zapcore.Entry{Level: zapcore.ErrorLevel, Message: "err"}, nil)
	})

	t.Run("custom levels", func(t *testing.T) {
		provider := sdklog.NewLoggerProvider()
		t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

		core := NewOTelZapCore(provider, []zapcore.Level{zapcore.WarnLevel})
		wrapped, ok := core.(*otelZapLevelCore)
		require.True(t, ok)
		assert.True(t, wrapped.levelEnabled(zapcore.WarnLevel))
		assert.False(t, wrapped.levelEnabled(zapcore.ErrorLevel))
	})
}

func TestAttachOTelZapLogger(t *testing.T) {
	origLog := logger.Log
	origSugar := logger.Sugar
	t.Cleanup(func() {
		logger.Log = origLog
		logger.Sugar = origSugar
	})

	t.Run("nil provider", func(t *testing.T) {
		logger.Log = zap.NewNop()
		logger.Sugar = logger.Log.Sugar()
		AttachOTelZapLogger(nil, nil)
		assert.NotNil(t, logger.Log)
	})

	t.Run("nil logger", func(t *testing.T) {
		logger.Log = nil
		logger.Sugar = nil
		provider := sdklog.NewLoggerProvider()
		t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })
		AttachOTelZapLogger(provider, nil)
		assert.Nil(t, logger.Log)
	})

	t.Run("attaches core", func(t *testing.T) {
		logger.Log = zap.NewNop()
		logger.Sugar = logger.Log.Sugar()
		provider := sdklog.NewLoggerProvider()
		t.Cleanup(func() { _ = provider.Shutdown(t.Context()) })

		AttachOTelZapLogger(provider, []zapcore.Level{zapcore.ErrorLevel})
		require.NotNil(t, logger.Log)
		require.NotNil(t, logger.Sugar)
	})
}
