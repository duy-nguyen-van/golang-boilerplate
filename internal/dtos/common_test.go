package dtos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/utils/i18n"

	"github.com/labstack/echo/v5"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"
)

func newEchoContext() *echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	return echo.NewContext(req, rec, e)
}

func TestNewPageableRequest(t *testing.T) {
	pr := NewPageableRequest()
	require.NotNil(t, pr)
	assert.Equal(t, constants.DefaultPage, pr.Page)
	assert.Equal(t, constants.DefaultPageSize, pr.PageSize)
}

func TestNewUnpaginatedRequest(t *testing.T) {
	pr := NewUnpaginatedRequest()
	require.NotNil(t, pr)
	assert.Equal(t, constants.DefaultPage, pr.Page)
	assert.Equal(t, constants.NoLimit, pr.PageSize)
}

func TestPageableRequest_ShouldPaginate(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		expected bool
	}{
		{name: "positive page size", pageSize: 10, expected: true},
		{name: "zero page size", pageSize: 0, expected: false},
		{name: "negative page size", pageSize: -1, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PageableRequest{PageSize: tt.pageSize}
			assert.Equal(t, tt.expected, pr.ShouldPaginate())
		})
	}
}

func TestPageableRequest_GetLimit(t *testing.T) {
	tests := []struct {
		name     string
		pageSize int
		expected int
	}{
		{name: "positive", pageSize: 25, expected: 25},
		{name: "zero returns no limit", pageSize: 0, expected: constants.NoLimit},
		{name: "negative returns no limit", pageSize: -5, expected: constants.NoLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PageableRequest{PageSize: tt.pageSize}
			assert.Equal(t, tt.expected, pr.GetLimit())
		})
	}
}

func TestPageableRequest_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		expected int
	}{
		{name: "first page", page: 1, pageSize: 10, expected: 0},
		{name: "second page", page: 2, pageSize: 10, expected: 10},
		{name: "zero page size", page: 3, pageSize: 0, expected: 0},
		{name: "negative page size", page: 3, pageSize: -1, expected: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &PageableRequest{Page: tt.page, PageSize: tt.pageSize}
			assert.Equal(t, tt.expected, pr.GetOffset())
		})
	}
}

func TestMeta_HttpCode(t *testing.T) {
	tests := []struct {
		name     string
		meta     Meta
		expected int
	}{
		{name: "explicit code wins", meta: Meta{Code: http.StatusCreated, ErrorCode: "400000"}, expected: http.StatusCreated},
		{name: "short error code", meta: Meta{ErrorCode: "40"}, expected: http.StatusInternalServerError},
		{name: "200 prefix", meta: Meta{ErrorCode: "200000"}, expected: http.StatusOK},
		{name: "400 prefix", meta: Meta{ErrorCode: "400001"}, expected: http.StatusBadRequest},
		{name: "401 prefix", meta: Meta{ErrorCode: "401000"}, expected: http.StatusUnauthorized},
		{name: "403 prefix", meta: Meta{ErrorCode: "403000"}, expected: http.StatusForbidden},
		{name: "404 prefix", meta: Meta{ErrorCode: "404000"}, expected: http.StatusNotFound},
		{name: "500 prefix", meta: Meta{ErrorCode: "500000"}, expected: http.StatusInternalServerError},
		{name: "unknown prefix", meta: Meta{ErrorCode: "999000"}, expected: http.StatusInternalServerError},
		{name: "empty error code", meta: Meta{ErrorCode: ""}, expected: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.meta.HttpCode())
		})
	}
}

func TestGetMeta(t *testing.T) {
	c := newEchoContext()
	bundle := goi18n.NewBundle(language.English)
	require.NoError(t, bundle.AddMessages(language.English, &goi18n.Message{
		ID:    "Code_200000",
		Other: "OK",
	}))
	c.Set(i18n.LocalizerContext, goi18n.NewLocalizer(bundle, "en"))

	meta := GetMeta(c, "200000", http.StatusOK)
	assert.Equal(t, "200000", meta.ErrorCode)
	assert.Equal(t, "OK", meta.Message)
	assert.Equal(t, http.StatusOK, meta.Code)
}

func TestGetMetaPaging(t *testing.T) {
	c := newEchoContext()
	pageable := &Pageable{Page: 2, PageSize: 20, Total: 55}

	meta := GetMetaPaging(c, "200000", pageable, http.StatusOK)
	assert.Equal(t, "200000", meta.ErrorCode)
	assert.Equal(t, "Code_200000", meta.Message) // no localizer → key fallback
	assert.Equal(t, http.StatusOK, meta.Code)
	assert.Equal(t, 2, meta.Page)
	assert.Equal(t, 20, meta.PageSize)
	assert.Equal(t, int64(55), meta.Total)
}

func TestBaseResponse_JSON(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := echo.NewContext(req, rec, e)

	resp := &BaseResponse[string]{
		Meta: Meta{Code: http.StatusOK, ErrorCode: "200000", Message: "ok"},
		Data: "hello",
	}

	require.NoError(t, resp.JSON(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	var decoded BaseResponse[string]
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &decoded))
	assert.Equal(t, "hello", decoded.Data)
	assert.Equal(t, "200000", decoded.Meta.ErrorCode)
}
