package monitoring

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const otelEchoScope = "golang-boilerplate/internal/monitoring/otelecho"

// NewOTelEcho returns Echo v5 middleware that creates a server span per request.
// Official otelecho still targets Echo v4; this is a lightweight local adapter.
func NewOTelEcho(serviceName string) echo.MiddlewareFunc {
	tracer := otel.GetTracerProvider().Tracer(otelEchoScope)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			request := c.Request()
			savedCtx := request.Context()
			defer func() {
				c.SetRequest(request.WithContext(savedCtx))
			}()

			ctx := propagator.Extract(savedCtx, propagation.HeaderCarrier(request.Header))

			attrs := []attribute.KeyValue{
				attribute.String("http.request.method", request.Method),
				attribute.String("url.path", request.URL.Path),
				attribute.String("server.address", serviceName),
			}

			spanName := fmt.Sprintf("%s %s", request.Method, request.URL.Path)
			if path := c.Path(); path != "" {
				spanName = fmt.Sprintf("%s %s", request.Method, path)
				attrs = append(attrs, attribute.String("http.route", path))
			}

			ctx, span := tracer.Start(ctx, spanName,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
				oteltrace.WithAttributes(attrs...),
			)
			defer span.End()

			c.SetRequest(request.WithContext(ctx))

			err := next(c)

			status := http.StatusOK
			if resp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && resp != nil && resp.Status != 0 {
				status = resp.Status
			}
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
				if he, ok := errors.AsType[*echo.HTTPError](err); ok {
					status = he.Code
				} else if status < http.StatusBadRequest {
					status = http.StatusInternalServerError
				}
			}
			span.SetAttributes(attribute.Int("http.response.status_code", status))
			if status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, http.StatusText(status))
			}

			return err
		}
	}
}
