package email

import (
	"testing"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"
	appErrors "golang-boilerplate/internal/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideEmailSender(t *testing.T) {
	t.Run("ses provider", func(t *testing.T) {
		sender, err := ProvideEmailSender(config.Config{
			EmailProvider:   constants.EmailProviderSES,
			AWSSESRegion:    "us-east-1",
			AWSSESAccessKey: "AKIAIOSFODNN7EXAMPLE",
			AWSSESSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		})
		require.NoError(t, err)
		require.NotNil(t, sender)
		_, ok := sender.(*SESSender)
		assert.True(t, ok)
	})

	t.Run("invalid provider", func(t *testing.T) {
		sender, err := ProvideEmailSender(config.Config{EmailProvider: "unknown"})
		require.Error(t, err)
		assert.Nil(t, sender)

		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.ErrorTypeInternal, appErr.Type)
		assert.Equal(t, "initialize_email_sender", appErr.Operation)
	})
}
