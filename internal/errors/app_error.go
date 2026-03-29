package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"golang-boilerplate/internal/constants"

	"github.com/go-playground/validator/v10"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeNotFound represents resource not found errors
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeUnauthorized represents authentication errors
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	// ErrorTypeForbidden represents authorization errors
	ErrorTypeForbidden ErrorType = "forbidden"
	// ErrorTypeConflict represents resource conflict errors
	ErrorTypeConflict ErrorType = "conflict"
	// ErrorTypeInternal represents internal server errors
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeExternal represents external service errors
	ErrorTypeExternal ErrorType = "external"
	// ErrorTypeDatabase represents database errors
	ErrorTypeDatabase ErrorType = "database"
	// ErrorTypeCache represents cache errors
	ErrorTypeCache ErrorType = "cache"
	// ErrorTypeTimeout represents timeout errors
	ErrorTypeTimeout ErrorType = "timeout"
)

// AppError represents a structured application error
type AppError struct {
	// Code is the error code from constants
	Code string `json:"code"`
	// Message is the human-readable error message
	Message string `json:"message"`
	// Type categorizes the error
	Type ErrorType `json:"type"`
	// HTTPStatus is the HTTP status code
	HTTPStatus int `json:"http_status"`
	// Cause is the underlying error
	Cause error `json:"-"`
	// Context contains additional context information
	Context map[string]interface{} `json:"context,omitempty"`
	// Timestamp when the error occurred
	Timestamp time.Time `json:"timestamp"`
	// Stack trace for debugging
	StackTrace string `json:"stack_trace,omitempty"`
	// Operation that was being performed when error occurred
	Operation string `json:"operation,omitempty"`
	// Resource that was being accessed
	Resource string `json:"resource,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithOperation sets the operation context
func (e *AppError) WithOperation(operation string) *AppError {
	e.Operation = operation
	return e
}

// WithResource sets the resource context
func (e *AppError) WithResource(resource string) *AppError {
	e.Resource = resource
	return e
}

// NewAppError creates a new AppError
func NewAppError(code, message string, errType ErrorType, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Type:       errType,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
		StackTrace: getStackTrace(),
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, code, message string, errType ErrorType, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Type:       errType,
		Cause:      err,
		HTTPStatus: httpStatus,
		Timestamp:  time.Now(),
		StackTrace: getStackTrace(),
	}
}

// ValidationError creates a validation error
func ValidationError(message string, cause error) *AppError {
	return WrapError(cause, constants.ValidationError, message, ErrorTypeValidation, http.StatusBadRequest)
}

// ValidationErrorWithDetails creates a validation error with detailed field errors
func ValidationErrorWithDetails(message string, cause error, fieldErrors map[string]string) *AppError {
	appErr := WrapError(cause, constants.ValidationError, message, ErrorTypeValidation, http.StatusBadRequest)

	// Add field-specific errors to context
	for field, errMsg := range fieldErrors {
		appErr = appErr.WithContext(field, errMsg)
	}

	return appErr
}

// ParseValidationErrors parses validator.ValidationErrors into a map of field errors
func ParseValidationErrors(err error) map[string]string {
	fieldErrors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, validationErr := range validationErrors {
			fieldName := validationErr.Field()
			fieldTag := validationErr.Tag()
			fieldValue := validationErr.Value()

			// Create user-friendly error messages based on validation tag
			var errorMessage string
			switch fieldTag {
			case "required":
				errorMessage = fmt.Sprintf("%s is required", fieldName)
			case "email":
				errorMessage = fmt.Sprintf("%s must be a valid email address", fieldName)
			case "min":
				errorMessage = fmt.Sprintf("%s must be at least %s characters long", fieldName, validationErr.Param())
			case "max":
				errorMessage = fmt.Sprintf("%s must be at most %s characters long", fieldName, validationErr.Param())
			case "len":
				errorMessage = fmt.Sprintf("%s must be exactly %s characters long", fieldName, validationErr.Param())
			case "numeric":
				errorMessage = fmt.Sprintf("%s must be a valid number", fieldName)
			case "alpha":
				errorMessage = fmt.Sprintf("%s must contain only letters", fieldName)
			case "alphanum":
				errorMessage = fmt.Sprintf("%s must contain only letters and numbers", fieldName)
			case "url":
				errorMessage = fmt.Sprintf("%s must be a valid URL", fieldName)
			case "uuid":
				errorMessage = fmt.Sprintf("%s must be a valid UUID", fieldName)
			case "oneof":
				errorMessage = fmt.Sprintf("%s must be one of: %s", fieldName, validationErr.Param())
			case "gte":
				errorMessage = fmt.Sprintf("%s must be greater than or equal to %s", fieldName, validationErr.Param())
			case "lte":
				errorMessage = fmt.Sprintf("%s must be less than or equal to %s", fieldName, validationErr.Param())
			case "gt":
				errorMessage = fmt.Sprintf("%s must be greater than %s", fieldName, validationErr.Param())
			case "lt":
				errorMessage = fmt.Sprintf("%s must be less than %s", fieldName, validationErr.Param())
			case "eq":
				errorMessage = fmt.Sprintf("%s must be equal to %s", fieldName, validationErr.Param())
			case "ne":
				errorMessage = fmt.Sprintf("%s must not be equal to %s", fieldName, validationErr.Param())
			case "unique":
				errorMessage = fmt.Sprintf("%s must be unique", fieldName)
			case "omitempty":
				// Skip omitempty errors as they are not actual validation failures
				continue
			default:
				errorMessage = fmt.Sprintf("%s is invalid (value: %v)", fieldName, fieldValue)
			}

			fieldErrors[fieldName] = errorMessage
		}
	}

	return fieldErrors
}

// NotFoundError creates a not found error
func NotFoundError(resource string, cause error) *AppError {
	msg := fmt.Sprintf("%s not found", resource)
	return WrapError(cause, constants.UserNotFound, msg, ErrorTypeNotFound, http.StatusNotFound)
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string, cause error) *AppError {
	return WrapError(cause, constants.Unauthorized, message, ErrorTypeUnauthorized, http.StatusUnauthorized)
}

// ForbiddenError creates a forbidden error
func ForbiddenError(message string, cause error) *AppError {
	return WrapError(cause, constants.Forbidden, message, ErrorTypeForbidden, http.StatusForbidden)
}

// ConflictError creates a conflict error
func ConflictError(message string, cause error) *AppError {
	return WrapError(cause, constants.BadRequest, message, ErrorTypeConflict, http.StatusConflict)
}

// InternalError creates an internal server error
func InternalError(message string, cause error) *AppError {
	return WrapError(cause, constants.InternalError, message, ErrorTypeInternal, http.StatusInternalServerError)
}

// DatabaseError creates a database error
func DatabaseError(message string, cause error) *AppError {
	return WrapError(cause, constants.DatabaseError, message, ErrorTypeDatabase, http.StatusInternalServerError)
}

// ExternalServiceError creates an external service error
func ExternalServiceError(message string, cause error) *AppError {
	return WrapError(cause, constants.ExternalServiceError, message, ErrorTypeExternal, http.StatusBadGateway)
}

// CacheError creates a cache error
func CacheError(message string, cause error) *AppError {
	return WrapError(cause, constants.InternalError, message, ErrorTypeCache, http.StatusInternalServerError)
}

// TimeoutError creates a timeout error
func TimeoutError(message string, cause error) *AppError {
	return WrapError(cause, constants.InternalError, message, ErrorTypeTimeout, http.StatusRequestTimeout)
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	// Clean up the stack trace to remove unnecessary lines
	lines := strings.Split(stack, "\n")
	var filteredLines []string

	for i, line := range lines {
		// Skip the first line (goroutine info) and lines containing runtime functions
		if i > 0 && !strings.Contains(line, "runtime.") && !strings.Contains(line, "reflect.") {
			filteredLines = append(filteredLines, line)
		}
	}

	return strings.Join(filteredLines, "\n")
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from an error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) string {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Code
	}
	return constants.InternalError
}

// GetHTTPStatus extracts the HTTP status from an error
func GetHTTPStatus(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// GetErrorMessage extracts the error message from an error
func GetErrorMessage(err error) string {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.Message
	}
	return err.Error()
}
