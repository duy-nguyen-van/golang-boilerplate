package monitoring

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestStartSpanAndEndSpan(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	t.Run("with attributes", func(t *testing.T) {
		ctx, span := StartSpan(context.Background(), "test-tracer", "op", attribute.String("k", "v"))
		require.NotNil(t, span)
		assert.NotNil(t, ctx)
		EndSpan(span, nil)
	})

	t.Run("records error", func(t *testing.T) {
		_, span := StartSpan(context.Background(), "test-tracer", "fail")
		EndSpan(span, errors.New("boom"))
	})

	t.Run("nil span is no-op", func(t *testing.T) {
		EndSpan(nil, errors.New("ignored"))
	})
}
