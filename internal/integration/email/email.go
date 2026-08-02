package email

import (
	"context"
	"fmt"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/errors"
)

// EmailMessage represents an email message
type EmailMessage struct {
	To       []string `json:"to"`
	CC       []string `json:"cc,omitempty"`
	BCC      []string `json:"bcc,omitempty"`
	Subject  string   `json:"subject"`
	Body     string   `json:"body"`
	HTMLBody string   `json:"html_body,omitempty"`
}

// EmailRequest represents a generic email request
type EmailRequest struct {
	To           []string               `json:"to"`
	Cc           []string               `json:"cc,omitempty"`
	Bcc          []string               `json:"bcc,omitempty"`
	Subject      string                 `json:"subject"`
	TemplateID   string                 `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	HTMLBody     string                 `json:"html_body,omitempty"`
	TextBody     string                 `json:"text_body,omitempty"`
	Attachments  []Attachment           `json:"attachments,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	ContentType string `json:"content_type"`
}

// EmailResponse represents the response from email providers
type EmailResponse struct {
	MessageID string `json:"message_id"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// TemplateData represents common template variables
type TemplateData struct {
	UserName     string `json:"user_name"`
	UserEmail    string `json:"user_email"`
	CompanyName  string `json:"company_name"`
	ActionURL    string `json:"action_url"`
	SupportEmail string `json:"support_email"`
	// Add more common fields as needed
}

// EmailSender defines the interface for sending emails
type EmailSender interface {
	SendEmail(ctx context.Context, message EmailRequest) (*EmailResponse, error)
	SendRawEmail(ctx context.Context, rawData []byte) (*EmailResponse, error)
}

func ProvideEmailSender(config config.Config) (EmailSender, error) {
	switch config.EmailProvider {
	case constants.EmailProviderSES:
		sesSender, err := NewSESSender(config)
		if err != nil {
			return nil, errors.ExternalServiceError("Failed to initialize SES email sender", err).
				WithOperation("initialize_email_sender").
				WithResource("email")
		}
		return sesSender, nil
	default:
		return nil, errors.InternalError("Invalid email provider", fmt.Errorf("invalid email provider: %s", config.EmailProvider)).
			WithOperation("initialize_email_sender").
			WithResource("email").
			WithContext("email_provider", config.EmailProvider)
	}
}
