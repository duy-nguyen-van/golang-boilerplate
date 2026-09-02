package monitoring

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func withTestTracer(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(t.Context())
		otel.SetTracerProvider(prev)
	})
	return recorder
}

func TestNewOTelEcho(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		route      string
		handler    echo.HandlerFunc
		wantStatus int
		wantError  bool
		spanStatus codes.Code
	}{
		{
			name:  "success uses route path",
			path:  "/users/42",
			route: "/users/:id",
			handler: func(c *echo.Context) error {
				return c.String(http.StatusOK, "ok")
			},
			wantStatus: http.StatusOK,
			spanStatus: codes.Unset,
		},
		{
			name: "http error",
			path: "/missing",
			handler: func(c *echo.Context) error {
				return echo.NewHTTPError(http.StatusNotFound, "missing")
			},
			wantStatus: http.StatusNotFound,
			wantError:  true,
			spanStatus: codes.Error,
		},
		{
			name: "generic error becomes 500",
			path: "/boom",
			handler: func(c *echo.Context) error {
				return errors.New("boom")
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  true,
			spanStatus: codes.Error,
		},
		{
			name: "generic error after 400 keeps status",
			path: "/bad",
			handler: func(c *echo.Context) error {
				if err := c.String(http.StatusBadRequest, "bad"); err != nil {
					return err
				}
				return errors.New("still failed")
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
			spanStatus: codes.Error,
		},
		{
			name: "server error status",
			path: "/fail",
			handler: func(c *echo.Context) error {
				return c.String(http.StatusInternalServerError, "fail")
			},
			wantStatus: http.StatusInternalServerError,
			spanStatus: codes.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := withTestTracer(t)
			e := echo.New()
			e.Use(NewOTelEcho("golang-boilerplate"))
			route := tt.route
			if route == "" {
				route = tt.path
			}
			e.GET(route, tt.handler)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			if !tt.wantError {
				assert.Equal(t, tt.wantStatus, rec.Code)
			}

			spans := recorder.Ended()
			require.NotEmpty(t, spans)
			span := spans[len(spans)-1]
			assert.Equal(t, tt.spanStatus, span.Status().Code)
		})
	}

	t.Run("empty route uses url path", func(t *testing.T) {
		recorder := withTestTracer(t)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/raw", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := NewOTelEcho("golang-boilerplate")
		require.NoError(t, mw(func(c *echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})(c))

		spans := recorder.Ended()
		require.NotEmpty(t, spans)
		assert.Equal(t, "GET /raw", spans[len(spans)-1].Name())
	})
}
