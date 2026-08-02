package i18n

import (
	"github.com/labstack/echo/v5"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

const LocalizerContext = "localizer"

func T(c *echo.Context, messageKey string, param map[string]interface{}) string {
	msg := &i18n.Message{
		ID: messageKey,
	}
	lz, ok := c.Get(LocalizerContext).(*i18n.Localizer)
	if !ok {
		return messageKey
	}
	str, err := lz.Localize(&i18n.LocalizeConfig{
		DefaultMessage: msg,
		TemplateData:   param,
	})
	if err != nil {
		return messageKey
	}
	return str
}
