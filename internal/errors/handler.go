package errors

import (
	"context"
	"fmt"
	"net/http"

	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/dtos"

	"golang-boilerplate/internal/logger"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ErrorHandler provides utilities for consistent error handling
type ErrorHandler struct{}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// HandleError processes an error and returns appropriate HTTP response
func (h *ErrorHandler) HandleError(c echo.Context, err error) error {
	appErr := h.processError(err)

	// Log the error with context
	h.logError(c, appErr)

	// Report to Sentry if available
	h.reportToSentry(c, appErr)

	// Return structured error response
	return h.errorResponse(c, appErr)
}

// processError converts various error types to AppError
func (h *ErrorHandler) processError(err error) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, return it
	if appErr := GetAppError(err); appErr != nil {
		return appErr
	}

	// Map Echo HTTP errors (e.g., CSRF 403) to proper AppError
	if httpErr, ok := err.(*echo.HTTPError); ok {
		status := httpErr.Code
		msg := fmt.Sprintf("%v", httpErr.Message)
		switch status {
		case http.StatusBadRequest:
			return ValidationError(msg, err)
		case http.StatusUnauthorized:
			return UnauthorizedError(msg, err)
		case http.StatusForbidden:
			return ForbiddenError(msg, err)
		case http.StatusNotFound:
			return NotFoundError("Resource", err)
		default:
			app := NewAppError(constants.InternalError, msg, ErrorTypeInternal, status)
			app.Cause = err
			return app
		}
	}

	// Handle GORM errors
	if gormErr := h.handleGORMError(err); gormErr != nil {
		return gormErr
	}

	// Handle context errors
	if ctxErr := h.handleContextError(err); ctxErr != nil {
		return ctxErr
	}

	// Default to internal error
	return InternalError("An unexpected error occurred", err)
}

// handleGORMError converts GORM errors to AppError
func (h *ErrorHandler) handleGORMError(err error) *AppError {
	switch err {
	case gorm.ErrRecordNotFound:
		return NotFoundError("Resource", err)
	case gorm.ErrInvalidTransaction:
		return DatabaseError("Invalid database transaction", err)
	case gorm.ErrNotImplemented:
		return DatabaseError("Database operation not implemented", err)
	case gorm.ErrMissingWhereClause:
		return DatabaseError("Missing WHERE clause in database operation", err)
	case gorm.ErrUnsupportedDriver:
		return DatabaseError("Unsupported database driver", err)
	case gorm.ErrRegistered:
		return DatabaseError("Database model already registered", err)
	case gorm.ErrInvalidField:
		return DatabaseError("Invalid database field", err)
	case gorm.ErrEmptySlice:
		return DatabaseError("Empty slice provided to database operation", err)
	case gorm.ErrDryRunModeUnsupported:
		return DatabaseError("Dry run mode not supported", err)
	case gorm.ErrInvalidDB:
		return DatabaseError("Invalid database connection", err)
	case gorm.ErrInvalidValue:
		return DatabaseError("Invalid value provided to database", err)
	case gorm.ErrInvalidValueOfLength:
		return DatabaseError("Invalid value length for database field", err)
	case gorm.ErrPreloadNotAllowed:
		return DatabaseError("Preload not allowed for this operation", err)
	default:
		// Check if it's a GORM error by checking the error message
		if isGORMError(err) {
			return DatabaseError("Database operation failed", err)
		}
		return nil
	}
}

// handleContextError converts context errors to AppError
func (h *ErrorHandler) handleContextError(err error) *AppError {
	switch err {
	case context.Canceled:
		return TimeoutError("Request was canceled", err)
	case context.DeadlineExceeded:
		return TimeoutError("Request timeout exceeded", err)
	default:
		return nil
	}
}

// isGORMError checks if an error is a GORM error
func isGORMError(err error) bool {
	errStr := err.Error()
	gormErrorPatterns := []string{
		"database",
		"sql",
		"connection",
		"transaction",
		"constraint",
		"foreign key",
		"unique constraint",
		"duplicate key",
	}

	for _, pattern := range gormErrorPatterns {
		if containsIgnoreCase(errStr, pattern) {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks if a string contains another string (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

// containsSubstring checks if s contains substr (case insensitive)
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// logError logs the error with appropriate level and context
func (h *ErrorHandler) logError(c echo.Context, appErr *AppError) {
	fields := []zap.Field{
		zap.String("error_code", appErr.Code),
		zap.String("error_type", string(appErr.Type)),
		zap.Int("http_status", appErr.HTTPStatus),
		zap.String("operation", appErr.Operation),
		zap.String("resource", appErr.Resource),
		zap.Time("timestamp", appErr.Timestamp),
	}

	// Add request context
	if c != nil {
		fields = append(fields,
			zap.String("path", c.Request().URL.Path),
			zap.String("method", c.Request().Method),
			zap.Any("query", c.QueryParams()),
			zap.String("user_agent", c.Request().UserAgent()),
			zap.String("ip", c.RealIP()),
		)
	}

	// Add error context
	for k, v := range appErr.Context {
		fields = append(fields, zap.Any("context_"+k, v))
	}

	// Add stack trace for internal errors
	if appErr.Type == ErrorTypeInternal || appErr.Type == ErrorTypeDatabase {
		fields = append(fields, zap.String("stack_trace", appErr.StackTrace))
	}

	// Log with appropriate level
	switch appErr.Type {
	case ErrorTypeValidation, ErrorTypeNotFound, ErrorTypeUnauthorized, ErrorTypeForbidden:
		logger.Log.Warn(appErr.Message, fields...)
	case ErrorTypeInternal, ErrorTypeDatabase, ErrorTypeExternal, ErrorTypeCache:
		logger.Log.Error(appErr.Message, fields...)
	default:
		logger.Log.Error(appErr.Message, fields...)
	}
}

// reportToSentry reports the error to Sentry
func (h *ErrorHandler) reportToSentry(c echo.Context, appErr *AppError) {
	if hub := sentry.GetHubFromContext(c.Request().Context()); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			// Set error context
			scope.SetTag("error_code", appErr.Code)
			scope.SetTag("error_type", string(appErr.Type))
			scope.SetTag("http_status", fmt.Sprintf("%d", appErr.HTTPStatus))
			scope.SetTag("operation", appErr.Operation)
			scope.SetTag("resource", appErr.Resource)

			// Set request context
			if c != nil {
				scope.SetExtra("path", c.Request().URL.Path)
				scope.SetExtra("method", c.Request().Method)
				scope.SetExtra("query", c.QueryParams())
				scope.SetExtra("headers", c.Request().Header)
				scope.SetExtra("user_agent", c.Request().UserAgent())
				scope.SetExtra("ip", c.RealIP())
			}

			// Set error context
			for k, v := range appErr.Context {
				scope.SetExtra("error_context_"+k, v)
			}

			// Set stack trace for debugging
			if appErr.StackTrace != "" {
				scope.SetExtra("stack_trace", appErr.StackTrace)
			}

			// Capture the error
			hub.CaptureException(appErr)
		})
	}
}

// errorResponse creates a structured error response
func (h *ErrorHandler) errorResponse(c echo.Context, appErr *AppError) error {
	res := &dtos.BaseResponse[any]{}
	res.Meta = dtos.GetMeta(c, appErr.Code, appErr.HTTPStatus)
	res.Meta.Message = appErr.Message

	// Include field errors in the response if they exist
	if len(appErr.Context) > 0 {
		res.Data = appErr.Context
	} else {
		res.Data = nil
	}

	return res.JSON(c)
}

// SuccessResponse creates a structured success response
func (h *ErrorHandler) SuccessResponse(c echo.Context, message string, data any, page *dtos.Pageable) error {
	res := &dtos.BaseResponse[any]{}
	res.Meta = dtos.GetMeta(c, constants.Success, http.StatusOK)
	res.Meta.Message = message
	if page != nil {
		res.Meta.Page = page.Page
		res.Meta.PageSize = page.PageSize
		res.Meta.Total = page.Total
	}
	res.Data = data
	return res.JSON(c)
}

// ValidationErrorResponse creates a validation error response
func (h *ErrorHandler) ValidationErrorResponse(c echo.Context, message string, validationErrors map[string]string) error {
	appErr := ValidationError(message, nil)
	for field, errMsg := range validationErrors {
		appErr.WithContext("validation_"+field, errMsg)
	}
	return h.errorResponse(c, appErr)
}

// NotFoundErrorResponse creates a not found error response
func (h *ErrorHandler) NotFoundErrorResponse(c echo.Context, resource string) error {
	appErr := NotFoundError(resource, nil)
	return h.errorResponse(c, appErr)
}

// UnauthorizedErrorResponse creates an unauthorized error response
func (h *ErrorHandler) UnauthorizedErrorResponse(c echo.Context, message string) error {
	appErr := UnauthorizedError(message, nil)
	return h.errorResponse(c, appErr)
}

// ForbiddenErrorResponse creates a forbidden error response
func (h *ErrorHandler) ForbiddenErrorResponse(c echo.Context, message string) error {
	appErr := ForbiddenError(message, nil)
	return h.errorResponse(c, appErr)
}

// InternalErrorResponse creates an internal error response
func (h *ErrorHandler) InternalErrorResponse(c echo.Context, message string, cause error) error {
	appErr := InternalError(message, cause)
	return h.errorResponse(c, appErr)
}
