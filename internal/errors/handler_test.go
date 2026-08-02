package errors

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/dtos"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTestContext(method, path string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func newTestContextWithHub(method, path string) (*echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	hub := sentry.CurrentHub().Clone()
	ctx := sentry.SetHubOnContext(req.Context(), hub)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestNewErrorHandler(t *testing.T) {
	t.Parallel()
	assert.NotNil(t, NewErrorHandler())
}

func TestHandleError_Paths(t *testing.T) {
	h := NewErrorHandler()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		withHub    bool
	}{
		{
			name:       "app error",
			err:        ValidationError("bad input", nil).WithContext("email", "required"),
			wantStatus: http.StatusBadRequest,
			wantCode:   constants.ValidationError,
		},
		{
			name:       "echo bad request empty message",
			err:        echo.NewHTTPError(http.StatusBadRequest, ""),
			wantStatus: http.StatusBadRequest,
			wantCode:   constants.ValidationError,
		},
		{
			name:       "echo unauthorized",
			err:        echo.NewHTTPError(http.StatusUnauthorized, "auth required"),
			wantStatus: http.StatusUnauthorized,
			wantCode:   constants.Unauthorized,
		},
		{
			name:       "echo forbidden",
			err:        echo.NewHTTPError(http.StatusForbidden, "forbidden"),
			wantStatus: http.StatusForbidden,
			wantCode:   constants.Forbidden,
		},
		{
			name:       "echo not found",
			err:        echo.NewHTTPError(http.StatusNotFound, "missing"),
			wantStatus: http.StatusNotFound,
			wantCode:   constants.UserNotFound,
		},
		{
			name:       "echo default status",
			err:        echo.NewHTTPError(http.StatusTeapot, "teapot"),
			wantStatus: http.StatusTeapot,
			wantCode:   constants.InternalError,
		},
		{
			name:       "gorm record not found",
			err:        gorm.ErrRecordNotFound,
			wantStatus: http.StatusNotFound,
			wantCode:   constants.UserNotFound,
		},
		{
			name:       "context canceled",
			err:        context.Canceled,
			wantStatus: http.StatusRequestTimeout,
			wantCode:   constants.InternalError,
		},
		{
			name:       "context deadline exceeded",
			err:        context.DeadlineExceeded,
			wantStatus: http.StatusRequestTimeout,
			wantCode:   constants.InternalError,
		},
		{
			name:       "generic error",
			err:        fmt.Errorf("unexpected boom"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   constants.InternalError,
		},
		{
			name:       "gorm-like pattern error with sentry hub",
			err:        fmt.Errorf("duplicate key value violates unique constraint"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   constants.DatabaseError,
			withHub:    true,
		},
		{
			name:       "conflict app error logs default branch",
			err:        ConflictError("already exists", nil),
			wantStatus: http.StatusConflict,
			wantCode:   constants.BadRequest,
			withHub:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c *echo.Context
			var rec *httptest.ResponseRecorder
			if tt.withHub {
				c, rec = newTestContextWithHub(http.MethodGet, "/api/v1/test")
			} else {
				c, rec = newTestContext(http.MethodGet, "/api/v1/test")
			}

			err := h.HandleError(c, tt.err)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantCode)
		})
	}
}

func TestProcessError_NilAndGORMVariants(t *testing.T) {
	t.Parallel()
	h := NewErrorHandler()

	assert.Nil(t, h.processError(nil))

	gormCases := []struct {
		name string
		err  error
	}{
		{"invalid transaction", gorm.ErrInvalidTransaction},
		{"not implemented", gorm.ErrNotImplemented},
		{"missing where", gorm.ErrMissingWhereClause},
		{"unsupported driver", gorm.ErrUnsupportedDriver},
		{"registered", gorm.ErrRegistered},
		{"invalid field", gorm.ErrInvalidField},
		{"empty slice", gorm.ErrEmptySlice},
		{"dry run", gorm.ErrDryRunModeUnsupported},
		{"invalid db", gorm.ErrInvalidDB},
		{"invalid value", gorm.ErrInvalidValue},
		{"invalid value length", gorm.ErrInvalidValueOfLength},
		{"preload not allowed", gorm.ErrPreloadNotAllowed},
	}

	for _, tt := range gormCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			appErr := h.processError(tt.err)
			require.NotNil(t, appErr)
			assert.Equal(t, ErrorTypeDatabase, appErr.Type)
			assert.Equal(t, constants.DatabaseError, appErr.Code)
		})
	}

	t.Run("non gorm returns nil from handleGORMError", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, h.handleGORMError(fmt.Errorf("totally unrelated")))
	})
}

func TestIsGORMErrorAndContainsHelpers(t *testing.T) {
	t.Parallel()

	patterns := []string{
		"database offline",
		"sql: connection refused",
		"lost connection",
		"transaction aborted",
		"constraint failed",
		"foreign key mismatch",
		"unique constraint violated",
		"duplicate key detected",
	}
	for _, msg := range patterns {
		assert.True(t, isGORMError(fmt.Errorf("%s", msg)), msg)
	}
	assert.False(t, isGORMError(fmt.Errorf("nothing matching")))

	assert.True(t, containsIgnoreCase("database", "database"))
	assert.True(t, containsIgnoreCase("database error", "database"))
	assert.True(t, containsIgnoreCase("error database", "database"))
	assert.True(t, containsIgnoreCase("xx database yy", "database"))
	assert.False(t, containsIgnoreCase("abc", "abcd"))
	assert.False(t, containsIgnoreCase("hello", "xyz"))
	assert.False(t, containsIgnoreCase("abc", "xyz"))

	assert.True(t, containsSubstring("hello", ""))
	assert.True(t, containsSubstring("", ""))
	assert.False(t, containsSubstring("ab", "abc"))
	assert.False(t, containsSubstring("abcdef", "zzz"))
	assert.True(t, containsSubstring("abcdef", "cde"))
}

func TestLogError_LevelsAndNilContext(t *testing.T) {
	t.Parallel()
	h := NewErrorHandler()

	types := []ErrorType{
		ErrorTypeValidation,
		ErrorTypeNotFound,
		ErrorTypeUnauthorized,
		ErrorTypeForbidden,
		ErrorTypeInternal,
		ErrorTypeDatabase,
		ErrorTypeExternal,
		ErrorTypeCache,
		ErrorTypeConflict,
		ErrorTypeTimeout,
	}

	for _, typ := range types {
		appErr := NewAppError(constants.InternalError, "msg", typ, http.StatusInternalServerError).
			WithContext("k", "v").
			WithOperation("op").
			WithResource("res")
		h.logError(nil, appErr)
	}

	c, _ := newTestContext(http.MethodPost, "/api/v1/users?q=1")
	h.logError(c, ValidationError("bad", nil))
}

func TestReportToSentry_WithAndWithoutHub(t *testing.T) {
	t.Parallel()
	h := NewErrorHandler()

	appErr := InternalError("boom", fmt.Errorf("root")).
		WithContext("k", "v").
		WithOperation("op").
		WithResource("res")

	c, _ := newTestContext(http.MethodGet, "/api/v1/x")
	h.reportToSentry(c, appErr) // nil hub path

	cHub, _ := newTestContextWithHub(http.MethodGet, "/api/v1/x")
	h.reportToSentry(cHub, appErr)

	emptyTrace := NewAppError(constants.InternalError, "no stack", ErrorTypeInternal, 500)
	emptyTrace.StackTrace = ""
	h.reportToSentry(cHub, emptyTrace)
}

func TestErrorAndSuccessResponses(t *testing.T) {
	h := NewErrorHandler()

	t.Run("errorResponse with context and empty message", func(t *testing.T) {
		c, rec := newTestContext(http.MethodGet, "/api/v1/x")
		appErr := NewAppError(constants.ValidationError, "", ErrorTypeValidation, http.StatusBadRequest)
		require.NoError(t, h.errorResponse(c, appErr))
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		c2, rec2 := newTestContext(http.MethodGet, "/api/v1/x")
		appErr2 := ValidationError("bad", nil).WithContext("email", "required")
		require.NoError(t, h.errorResponse(c2, appErr2))
		assert.Equal(t, http.StatusBadRequest, rec2.Code)
		assert.Contains(t, rec2.Body.String(), "email")
	})

	t.Run("SuccessResponse without page", func(t *testing.T) {
		c, rec := newTestContext(http.MethodGet, "/api/v1/x")
		require.NoError(t, h.SuccessResponse(c, "ok", map[string]string{"a": "b"}, nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "ok")
	})

	t.Run("SuccessResponse with page", func(t *testing.T) {
		c, rec := newTestContext(http.MethodGet, "/api/v1/x")
		page := &dtos.Pageable{Page: 2, PageSize: 10, Total: 42}
		require.NoError(t, h.SuccessResponse(c, "listed", []int{1}, page))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"page":2`)
	})

	t.Run("typed error response helpers", func(t *testing.T) {
		helpers := []struct {
			name string
			fn   func(*echo.Context) error
			code int
		}{
			{
				name: "ValidationErrorResponse",
				fn: func(c *echo.Context) error {
					return h.ValidationErrorResponse(c, "invalid", map[string]string{"name": "required"})
				},
				code: http.StatusBadRequest,
			},
			{
				name: "NotFoundErrorResponse",
				fn: func(c *echo.Context) error {
					return h.NotFoundErrorResponse(c, "Company")
				},
				code: http.StatusNotFound,
			},
			{
				name: "UnauthorizedErrorResponse",
				fn: func(c *echo.Context) error {
					return h.UnauthorizedErrorResponse(c, "login required")
				},
				code: http.StatusUnauthorized,
			},
			{
				name: "ForbiddenErrorResponse",
				fn: func(c *echo.Context) error {
					return h.ForbiddenErrorResponse(c, "no access")
				},
				code: http.StatusForbidden,
			},
			{
				name: "InternalErrorResponse",
				fn: func(c *echo.Context) error {
					return h.InternalErrorResponse(c, "failed", fmt.Errorf("root"))
				},
				code: http.StatusInternalServerError,
			},
		}

		for _, tt := range helpers {
			t.Run(tt.name, func(t *testing.T) {
				c, rec := newTestContext(http.MethodGet, "/api/v1/x")
				require.NoError(t, tt.fn(c))
				assert.Equal(t, tt.code, rec.Code)
			})
		}
	})
}
