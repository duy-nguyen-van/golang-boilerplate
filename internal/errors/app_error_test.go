package errors

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"golang-boilerplate/internal/constants"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_ErrorAndUnwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        *AppError
		wantMsg    string
		wantUnwrap error
	}{
		{
			name:       "without cause",
			err:        NewAppError(constants.InternalError, "boom", ErrorTypeInternal, http.StatusInternalServerError),
			wantMsg:    constants.InternalError + ": boom",
			wantUnwrap: nil,
		},
		{
			name:       "with cause",
			err:        WrapError(fmt.Errorf("root"), constants.DatabaseError, "db failed", ErrorTypeDatabase, http.StatusInternalServerError),
			wantMsg:    constants.DatabaseError + ": db failed (caused by: root)",
			wantUnwrap: fmt.Errorf("root"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantMsg, tt.err.Error())
			if tt.wantUnwrap == nil {
				assert.Nil(t, tt.err.Unwrap())
			} else {
				assert.EqualError(t, tt.err.Unwrap(), tt.wantUnwrap.Error())
			}
		})
	}
}

func TestAppError_WithContextOperationResource(t *testing.T) {
	t.Parallel()

	err := NewAppError(constants.BadRequest, "bad", ErrorTypeValidation, http.StatusBadRequest)
	assert.Nil(t, err.Context)

	got := err.WithContext("field", "invalid").
		WithContext("other", 1).
		WithOperation("create_user").
		WithResource("users")

	assert.Same(t, err, got)
	assert.Equal(t, map[string]interface{}{"field": "invalid", "other": 1}, err.Context)
	assert.Equal(t, "create_user", err.Operation)
	assert.Equal(t, "users", err.Resource)
}

func TestConstructors(t *testing.T) {
	t.Parallel()

	cause := fmt.Errorf("cause")

	tests := []struct {
		name       string
		err        *AppError
		wantCode   string
		wantType   ErrorType
		wantStatus int
		wantCause  bool
	}{
		{
			name:       "NewAppError",
			err:        NewAppError(constants.InternalError, "msg", ErrorTypeInternal, http.StatusInternalServerError),
			wantCode:   constants.InternalError,
			wantType:   ErrorTypeInternal,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "WrapError",
			err:        WrapError(cause, constants.DatabaseError, "wrapped", ErrorTypeDatabase, http.StatusInternalServerError),
			wantCode:   constants.DatabaseError,
			wantType:   ErrorTypeDatabase,
			wantStatus: http.StatusInternalServerError,
			wantCause:  true,
		},
		{
			name:       "ValidationError",
			err:        ValidationError("invalid", cause),
			wantCode:   constants.ValidationError,
			wantType:   ErrorTypeValidation,
			wantStatus: http.StatusBadRequest,
			wantCause:  true,
		},
		{
			name:       "ValidationErrorWithDetails",
			err:        ValidationErrorWithDetails("invalid", cause, map[string]string{"Email": "required"}),
			wantCode:   constants.ValidationError,
			wantType:   ErrorTypeValidation,
			wantStatus: http.StatusBadRequest,
			wantCause:  true,
		},
		{
			name:       "ValidationErrorWithDetails empty map",
			err:        ValidationErrorWithDetails("invalid", nil, nil),
			wantCode:   constants.ValidationError,
			wantType:   ErrorTypeValidation,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "NotFoundError",
			err:        NotFoundError("User", cause),
			wantCode:   constants.UserNotFound,
			wantType:   ErrorTypeNotFound,
			wantStatus: http.StatusNotFound,
			wantCause:  true,
		},
		{
			name:       "UnauthorizedError",
			err:        UnauthorizedError("nope", cause),
			wantCode:   constants.Unauthorized,
			wantType:   ErrorTypeUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantCause:  true,
		},
		{
			name:       "ForbiddenError",
			err:        ForbiddenError("denied", cause),
			wantCode:   constants.Forbidden,
			wantType:   ErrorTypeForbidden,
			wantStatus: http.StatusForbidden,
			wantCause:  true,
		},
		{
			name:       "ConflictError",
			err:        ConflictError("exists", cause),
			wantCode:   constants.BadRequest,
			wantType:   ErrorTypeConflict,
			wantStatus: http.StatusConflict,
			wantCause:  true,
		},
		{
			name:       "InternalError",
			err:        InternalError("oops", cause),
			wantCode:   constants.InternalError,
			wantType:   ErrorTypeInternal,
			wantStatus: http.StatusInternalServerError,
			wantCause:  true,
		},
		{
			name:       "DatabaseError",
			err:        DatabaseError("db", cause),
			wantCode:   constants.DatabaseError,
			wantType:   ErrorTypeDatabase,
			wantStatus: http.StatusInternalServerError,
			wantCause:  true,
		},
		{
			name:       "ExternalServiceError",
			err:        ExternalServiceError("upstream", cause),
			wantCode:   constants.ExternalServiceError,
			wantType:   ErrorTypeExternal,
			wantStatus: http.StatusBadGateway,
			wantCause:  true,
		},
		{
			name:       "CacheError",
			err:        CacheError("redis", cause),
			wantCode:   constants.InternalError,
			wantType:   ErrorTypeCache,
			wantStatus: http.StatusInternalServerError,
			wantCause:  true,
		},
		{
			name:       "TimeoutError",
			err:        TimeoutError("slow", cause),
			wantCode:   constants.InternalError,
			wantType:   ErrorTypeTimeout,
			wantStatus: http.StatusRequestTimeout,
			wantCause:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.NotNil(t, tt.err)
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, tt.wantType, tt.err.Type)
			assert.Equal(t, tt.wantStatus, tt.err.HTTPStatus)
			assert.NotEmpty(t, tt.err.StackTrace)
			assert.False(t, tt.err.Timestamp.IsZero())
			if tt.wantCause {
				assert.Error(t, tt.err.Cause)
			}
		})
	}

	details := ValidationErrorWithDetails("invalid", cause, map[string]string{"Email": "required"})
	assert.Equal(t, "required", details.Context["Email"])
}

func TestParseValidationErrors(t *testing.T) {
	t.Parallel()

	validate := validator.New()

	type sample struct {
		Required  string   `validate:"required"`
		Email     string   `validate:"email"`
		Min       string   `validate:"min=3"`
		Max       string   `validate:"max=2"`
		Len       string   `validate:"len=2"`
		Numeric   string   `validate:"numeric"`
		Alpha     string   `validate:"alpha"`
		Alphanum  string   `validate:"alphanum"`
		URL       string   `validate:"url"`
		UUID      string   `validate:"uuid"`
		OneOf     string   `validate:"oneof=a b"`
		Gte       int      `validate:"gte=10"`
		Lte       int      `validate:"lte=1"`
		Gt        int      `validate:"gt=5"`
		Lt        int      `validate:"lt=1"`
		Eq        int      `validate:"eq=3"`
		Ne        int      `validate:"ne=0"`
		Unique    []string `validate:"unique"`
		Unknown   string   `validate:"startswith=ZZ"`
		SkipEmpty string   `validate:"omitempty,email"`
	}

	err := validate.Struct(sample{
		Required:  "",
		Email:     "bad",
		Min:       "ab",
		Max:       "abcd",
		Len:       "a",
		Numeric:   "x",
		Alpha:     "1",
		Alphanum:  "!",
		URL:       "not-a-url",
		UUID:      "nope",
		OneOf:     "c",
		Gte:       1,
		Lte:       5,
		Gt:        1,
		Lt:        5,
		Eq:        1,
		Ne:        0,
		Unique:    []string{"a", "a"},
		Unknown:   "hello",
		SkipEmpty: "",
	})
	require.Error(t, err)

	got := ParseValidationErrors(err)
	assert.Equal(t, "Required is required", got["Required"])
	assert.Equal(t, "Email must be a valid email address", got["Email"])
	assert.Equal(t, "Min must be at least 3 characters long", got["Min"])
	assert.Equal(t, "Max must be at most 2 characters long", got["Max"])
	assert.Equal(t, "Len must be exactly 2 characters long", got["Len"])
	assert.Equal(t, "Numeric must be a valid number", got["Numeric"])
	assert.Equal(t, "Alpha must contain only letters", got["Alpha"])
	assert.Equal(t, "Alphanum must contain only letters and numbers", got["Alphanum"])
	assert.Equal(t, "URL must be a valid URL", got["URL"])
	assert.Equal(t, "UUID must be a valid UUID", got["UUID"])
	assert.Equal(t, "OneOf must be one of: a b", got["OneOf"])
	assert.Equal(t, "Gte must be greater than or equal to 10", got["Gte"])
	assert.Equal(t, "Lte must be less than or equal to 1", got["Lte"])
	assert.Equal(t, "Gt must be greater than 5", got["Gt"])
	assert.Equal(t, "Lt must be less than 1", got["Lt"])
	assert.Equal(t, "Eq must be equal to 3", got["Eq"])
	assert.Equal(t, "Ne must not be equal to 0", got["Ne"])
	assert.Equal(t, "Unique must be unique", got["Unique"])
	assert.Contains(t, got["Unknown"], "is invalid")
	_, hasSkip := got["SkipEmpty"]
	assert.False(t, hasSkip)

	assert.Empty(t, ParseValidationErrors(fmt.Errorf("not validation")))

	// omitempty is not emitted by the validator as a failure tag; cover via stub FieldError.
	omitErr := validator.ValidationErrors{stubFieldError{field: "Optional", tag: "omitempty", value: ""}}
	assert.Empty(t, ParseValidationErrors(omitErr))
}

func TestIsAppErrorHelpers(t *testing.T) {
	t.Parallel()

	appErr := NotFoundError("User", nil)
	plain := fmt.Errorf("plain")

	tests := []struct {
		name        string
		err         error
		isApp       bool
		wantCode    string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "app error",
			err:         appErr,
			isApp:       true,
			wantCode:    constants.UserNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "User not found",
		},
		{
			name:        "wrapped app error",
			err:         fmt.Errorf("wrap: %w", appErr),
			isApp:       true,
			wantCode:    constants.UserNotFound,
			wantStatus:  http.StatusNotFound,
			wantMessage: "User not found",
		},
		{
			name:        "plain error",
			err:         plain,
			isApp:       false,
			wantCode:    constants.InternalError,
			wantStatus:  http.StatusInternalServerError,
			wantMessage: "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.isApp, IsAppError(tt.err))
			if tt.isApp {
				assert.NotNil(t, GetAppError(tt.err))
			} else {
				assert.Nil(t, GetAppError(tt.err))
			}
			assert.Equal(t, tt.wantCode, GetErrorCode(tt.err))
			assert.Equal(t, tt.wantStatus, GetHTTPStatus(tt.err))
			assert.Equal(t, tt.wantMessage, GetErrorMessage(tt.err))
		})
	}
}

// stubFieldError implements validator.FieldError for edge-case coverage.
type stubFieldError struct {
	field, tag, param string
	value             interface{}
}

func (s stubFieldError) Tag() string                      { return s.tag }
func (s stubFieldError) ActualTag() string                { return s.tag }
func (s stubFieldError) Namespace() string                { return s.field }
func (s stubFieldError) StructNamespace() string          { return s.field }
func (s stubFieldError) Field() string                    { return s.field }
func (s stubFieldError) StructField() string              { return s.field }
func (s stubFieldError) Value() interface{}               { return s.value }
func (s stubFieldError) Param() string                    { return s.param }
func (s stubFieldError) Kind() reflect.Kind               { return reflect.String }
func (s stubFieldError) Type() reflect.Type               { return reflect.TypeOf("") }
func (s stubFieldError) Translate(_ ut.Translator) string { return s.Error() }
func (s stubFieldError) Error() string                    { return s.field + " " + s.tag }
