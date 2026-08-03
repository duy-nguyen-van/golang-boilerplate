package handlers

import (
	"golang-boilerplate/internal/dtos"
	"golang-boilerplate/internal/errors"

	"github.com/labstack/echo/v5"
)

type BaseHandler struct {
	errorHandler *errors.ErrorHandler
}

// NewBaseHandler creates a new base handler with error handling
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		errorHandler: errors.NewErrorHandler(),
	}
}

// HandleError processes an error and returns appropriate HTTP response
func (b *BaseHandler) HandleError(c *echo.Context, err error) error {
	return b.errorHandler.HandleError(c, err)
}

// SuccessResponse creates a structured success response
func (b *BaseHandler) SuccessResponse(c *echo.Context, message string, data any, page *dtos.Pageable) error {
	return b.errorHandler.SuccessResponse(c, message, data, page)
}

// ValidationErrorResponse creates a validation error response
func (b *BaseHandler) ValidationErrorResponse(c *echo.Context, message string, validationErrors map[string]string) error {
	return b.errorHandler.ValidationErrorResponse(c, message, validationErrors)
}

// NotFoundErrorResponse creates a not found error response
func (b *BaseHandler) NotFoundErrorResponse(c *echo.Context, resource string) error {
	return b.errorHandler.NotFoundErrorResponse(c, resource)
}

// UnauthorizedErrorResponse creates an unauthorized error response
func (b *BaseHandler) UnauthorizedErrorResponse(c *echo.Context, message string) error {
	return b.errorHandler.UnauthorizedErrorResponse(c, message)
}

// ForbiddenErrorResponse creates a forbidden error response
func (b *BaseHandler) ForbiddenErrorResponse(c *echo.Context, message string) error {
	return b.errorHandler.ForbiddenErrorResponse(c, message)
}

// InternalErrorResponse creates an internal error response
func (b *BaseHandler) InternalErrorResponse(c *echo.Context, message string, cause error) error {
	return b.errorHandler.InternalErrorResponse(c, message, cause)
}
