package email

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"golang-boilerplate/internal/config"
	appErrors "golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/logger"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}

func newTestSESSender(t *testing.T, cfg config.Config, handler http.HandlerFunc) *SESSender {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &SESSender{
		client: ses.New(ses.Options{
			Region:           "us-east-1",
			Credentials:      credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
			BaseEndpoint:     aws.String(server.URL),
			HTTPClient:       server.Client(),
			RetryMaxAttempts: 1,
		}),
		config: cfg,
	}
}

func sesXMLSuccess(action, messageID string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		_, _ = fmt.Fprintf(w, "<%[1]sResponse><%[1]sResult><MessageId>%s</MessageId></%[1]sResult></%[1]sResponse>", action, messageID)
	}
}

func sesXMLError(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, `<ErrorResponse><Error><Type>Sender</Type><Code>MessageRejected</Code><Message>%s</Message></Error></ErrorResponse>`, message)
	}
}

func parseSESForm(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	values, err := url.ParseQuery(string(body))
	require.NoError(t, err)
	return values
}

func TestNewSESSender(t *testing.T) {
	sender, err := NewSESSender(config.Config{
		AWSSESRegion:    "us-east-1",
		AWSSESAccessKey: "AKIAIOSFODNN7EXAMPLE",
		AWSSESSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	})
	require.NoError(t, err)
	require.NotNil(t, sender)
	assert.NotNil(t, sender.client)
}

func TestSESSender_SendEmail(t *testing.T) {
	cfg := config.Config{AWSSESAccessKey: "noreply@example.com"}

	t.Run("html body only", func(t *testing.T) {
		var form url.Values
		sender := newTestSESSender(t, cfg, func(w http.ResponseWriter, r *http.Request) {
			form = parseSESForm(t, r)
			sesXMLSuccess("SendEmail", "html-id")(w, r)
		})

		resp, err := sender.SendEmail(context.Background(), EmailRequest{
			To:       []string{"user@example.com"},
			Subject:  "Hello",
			HTMLBody: "<p>hi</p>",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "sent", resp.Status)
		assert.Equal(t, "html-id", resp.MessageID)
		assert.Equal(t, "ses", resp.Provider)
		assert.Equal(t, "SendEmail", form.Get("Action"))
		assert.Equal(t, "noreply@example.com", form.Get("Source"))
		assert.Equal(t, "<p>hi</p>", form.Get("Message.Body.Html.Data"))
		assert.Empty(t, form.Get("Message.Body.Text.Data"))
	})

	t.Run("html and text body", func(t *testing.T) {
		var form url.Values
		sender := newTestSESSender(t, cfg, func(w http.ResponseWriter, r *http.Request) {
			form = parseSESForm(t, r)
			sesXMLSuccess("SendEmail", "both-id")(w, r)
		})

		resp, err := sender.SendEmail(context.Background(), EmailRequest{
			To:       []string{"user@example.com"},
			Subject:  "Hello",
			HTMLBody: "<p>hi</p>",
			TextBody: "hi",
		})
		require.NoError(t, err)
		assert.Equal(t, "both-id", resp.MessageID)
		assert.Equal(t, "<p>hi</p>", form.Get("Message.Body.Html.Data"))
		assert.Equal(t, "hi", form.Get("Message.Body.Text.Data"))
	})

	t.Run("text body only with cc and bcc", func(t *testing.T) {
		var form url.Values
		sender := newTestSESSender(t, cfg, func(w http.ResponseWriter, r *http.Request) {
			form = parseSESForm(t, r)
			sesXMLSuccess("SendEmail", "text-id")(w, r)
		})

		resp, err := sender.SendEmail(context.Background(), EmailRequest{
			To:       []string{"user@example.com"},
			Cc:       []string{"cc@example.com"},
			Bcc:      []string{"bcc@example.com"},
			Subject:  "Hello",
			TextBody: "plain",
		})
		require.NoError(t, err)
		assert.Equal(t, "text-id", resp.MessageID)
		assert.Equal(t, "plain", form.Get("Message.Body.Text.Data"))
		assert.Empty(t, form.Get("Message.Body.Html.Data"))
		assert.Equal(t, "cc@example.com", form.Get("Destination.CcAddresses.member.1"))
		assert.Equal(t, "bcc@example.com", form.Get("Destination.BccAddresses.member.1"))
	})

	t.Run("missing body", func(t *testing.T) {
		sender := &SESSender{config: cfg}
		resp, err := sender.SendEmail(context.Background(), EmailRequest{
			To:      []string{"user@example.com"},
			Subject: "Hello",
		})
		require.Error(t, err)
		assert.Nil(t, resp)

		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.ErrorTypeExternal, appErr.Type)
	})

	t.Run("ses send fails", func(t *testing.T) {
		sender := newTestSESSender(t, cfg, sesXMLError("ses unavailable"))
		resp, err := sender.SendEmail(context.Background(), EmailRequest{
			To:       []string{"user@example.com"},
			Subject:  "Hello",
			TextBody: "plain",
		})
		require.Error(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "ses", resp.Provider)
		assert.Equal(t, "failed", resp.Status)
		assert.NotEmpty(t, resp.Error)

		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, appErrors.ErrorTypeExternal, appErr.Type)
		assert.Equal(t, "send_email", appErr.Operation)
	})
}

func TestSESSender_SendRawEmail(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var form url.Values
		sender := newTestSESSender(t, config.Config{}, func(w http.ResponseWriter, r *http.Request) {
			form = parseSESForm(t, r)
			sesXMLSuccess("SendRawEmail", "raw-id")(w, r)
		})

		resp, err := sender.SendRawEmail(context.Background(), []byte("raw-mime"))
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "raw-id", resp.MessageID)
		assert.Equal(t, "ses", resp.Provider)
		assert.Equal(t, "sent", resp.Status)
		assert.Equal(t, "SendRawEmail", form.Get("Action"))
		assert.NotEmpty(t, form.Get("RawMessage.Data"))
	})

	t.Run("ses send fails", func(t *testing.T) {
		sender := newTestSESSender(t, config.Config{}, sesXMLError("raw send failed"))
		resp, err := sender.SendRawEmail(context.Background(), []byte("raw-mime"))
		require.Error(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "ses", resp.Provider)
		assert.Equal(t, "failed", resp.Status)
		assert.NotEmpty(t, resp.Error)

		var appErr *appErrors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, "send_raw_email", appErr.Operation)
		assert.Equal(t, "ses", appErr.Resource)
	})
}
