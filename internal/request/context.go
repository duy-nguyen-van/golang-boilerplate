package request

import (
	"context"

	"golang-boilerplate/pkg/correlationid"
)

type ctxKey struct {
	name string
}

func (k ctxKey) String() string {
	return "golang-boilerplate context value " + k.name
}

var ctxKeyLanguageCode = ctxKey{"language_code"}

// LanguageCodeFromContext retrieves the language code from the context.
func LanguageCodeFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ctxKeyLanguageCode).(string)
	return val, ok
}

// NewLanguageCodeContext creates a new context with the given language code.
func NewLanguageCodeContext(ctx context.Context, languageCode string) context.Context {
	return context.WithValue(ctx, ctxKeyLanguageCode, languageCode)
}

var ctxKeyRequestTimestamp = ctxKey{"request_timestamp"}

// RequestTimestampFromContext retrieves the request timestamp from the context.
// The timestamp is stored as Unix milliseconds.
func RequestTimestampFromContext(ctx context.Context) (int64, bool) {
	val, ok := ctx.Value(ctxKeyRequestTimestamp).(int64)
	return val, ok
}

// NewRequestTimestampContext creates a new context with the given request timestamp.
func NewRequestTimestampContext(ctx context.Context, requestTimestamp int64) context.Context {
	return context.WithValue(ctx, ctxKeyRequestTimestamp, requestTimestamp)
}

// CorrelationIDFromContext retrieves the correlation ID from the context.
func CorrelationIDFromContext(ctx context.Context) (string, bool) {
	return correlationid.FromContext(ctx)
}

// NewCorrelationIDContext creates a new context with the given correlation ID.
func NewCorrelationIDContext(ctx context.Context, correlationID string) context.Context {
	return correlationid.NewContext(ctx, correlationID)
}

var ctxKeyRequestURL = ctxKey{"request_url"}

// RequestURLFromContext retrieves the request URL from the context.
func RequestURLFromContext(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(ctxKeyRequestURL).(string)
	return val, ok
}

// NewRequestURLContext creates a new context with the given request URL.
func NewRequestURLContext(ctx context.Context, requestURL string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestURL, requestURL)
}
