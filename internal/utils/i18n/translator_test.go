package i18n

import (
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestT(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(c *echo.Context)
		messageKey string
		param      map[string]interface{}
		expected   string
	}{
		{
			name:       "no localizer returns key",
			setup:      func(c *echo.Context) {},
			messageKey: "Code_200000",
			param:      nil,
			expected:   "Code_200000",
		},
		{
			name: "wrong type in context returns key",
			setup: func(c *echo.Context) {
				c.Set(LocalizerContext, "not-a-localizer")
			},
			messageKey: "Code_400000",
			param:      nil,
			expected:   "Code_400000",
		},
		{
			name: "successful localize with template data",
			setup: func(c *echo.Context) {
				bundle := goi18n.NewBundle(language.English)
				require.NoError(t, bundle.AddMessages(language.English, &goi18n.Message{
					ID:    "Greeting",
					Other: "Hello, {{.Name}}!",
				}))
				c.Set(LocalizerContext, goi18n.NewLocalizer(bundle, "en"))
			},
			messageKey: "Greeting",
			param:      map[string]interface{}{"Name": "Ada"},
			expected:   "Hello, Ada!",
		},
		{
			name: "localize error returns key",
			setup: func(c *echo.Context) {
				bundle := goi18n.NewBundle(language.English)
				require.NoError(t, bundle.AddMessages(language.English, &goi18n.Message{
					ID:    "BadTemplate",
					Other: "Hello, {{.Name}",
				}))
				c.Set(LocalizerContext, goi18n.NewLocalizer(bundle, "en"))
			},
			messageKey: "MissingKeyWithEmptyDefault",
			param:      nil,
			expected:   "MissingKeyWithEmptyDefault",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newEchoContext()
			tt.setup(c)
			assert.Equal(t, tt.expected, T(c, tt.messageKey, tt.param))
		})
	}
}
