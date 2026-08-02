package dtos

import (
	"fmt"
	"net/http"
	"strings"

	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/utils/i18n"

	"github.com/labstack/echo/v5"
)

type BaseResponse[T any] struct {
	Meta Meta `json:"meta"`
	Data T    `json:"data"`
}

func (b *BaseResponse[T]) JSON(ctx *echo.Context) error {
	return ctx.JSON(b.Meta.HttpCode(), b)
}

// Meta represents metadata for paginated responses
// @Description Metadata for pagination
type Meta struct {
	ErrorCode string `json:"error_code" example:"BAD_REQUEST"`
	Message   string `json:"message" example:"Request is invalid"`
	Code      int    `json:"code,omitempty" example:"400"`
	Page      int    `json:"page,omitempty"  example:"1"`
	PageSize  int    `json:"page_size,omitempty" example:"20"`
	Total     int64  `json:"total,omitempty" example:"18"`
}

// PageableRequest is a struct for pagination request. It contains the page number and the page size. Page number starts from 1.
type PageableRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

type Pageable struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type DataResponse[T any] struct {
	Data     []T       `json:"data"`
	Pageable *Pageable `json:"pageable,omitempty"`
}

// NewPageableRequest creates a new PageableRequest with default values
func NewPageableRequest() *PageableRequest {
	return &PageableRequest{
		Page:     constants.DefaultPage,
		PageSize: constants.DefaultPageSize,
	}
}

// NewUnpaginatedRequest creates a new PageableRequest without pagination
func NewUnpaginatedRequest() *PageableRequest {
	return &PageableRequest{
		Page:     constants.DefaultPage,
		PageSize: constants.NoLimit,
	}
}

// ShouldPaginate returns true if pagination should be applied
func (pr *PageableRequest) ShouldPaginate() bool {
	return pr.PageSize > 0
}

// GetLimit returns the limit to be used in the query
func (pr *PageableRequest) GetLimit() int {
	if pr.PageSize <= 0 {
		return constants.NoLimit
	}
	return pr.PageSize
}

// GetOffset returns the offset to be used in the query
func (pr *PageableRequest) GetOffset() int {
	if pr.PageSize <= 0 {
		return 0
	}
	return (pr.Page - 1) * pr.PageSize
}

func (m *Meta) HttpCode() int {
	// If HttpStatus is explicitly set, use it
	if m.Code != 0 {
		return m.Code
	}

	// Fallback to parsing code field for backward compatibility
	const (
		lengthHTTPCode = 3
	)
	if len(m.ErrorCode) < lengthHTTPCode {
		return http.StatusInternalServerError
	}

	switch {
	case strings.HasPrefix(m.ErrorCode, "200"):
		return http.StatusOK
	case strings.HasPrefix(m.ErrorCode, "400"):
		return http.StatusBadRequest
	case strings.HasPrefix(m.ErrorCode, "401"):
		return http.StatusUnauthorized
	case strings.HasPrefix(m.ErrorCode, "403"):
		return http.StatusForbidden
	case strings.HasPrefix(m.ErrorCode, "404"):
		return http.StatusNotFound
	case strings.HasPrefix(m.ErrorCode, "500"):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func GetMeta(c *echo.Context, code string, httpStatus int) Meta {
	return Meta{
		ErrorCode: code,
		Message:   i18n.T(c, fmt.Sprintf("Code_%s", code), nil),
		Code:      httpStatus,
	}
}

func GetMetaPaging(c *echo.Context, code string, pageable *Pageable, httpStatus int) Meta {
	return Meta{
		ErrorCode: code,
		Message:   i18n.T(c, fmt.Sprintf("Code_%s", code), nil),
		Code:      httpStatus,
		PageSize:  pageable.PageSize,
		Page:      pageable.Page,
		Total:     pageable.Total,
	}
}
