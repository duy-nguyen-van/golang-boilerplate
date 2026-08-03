package services

import (
	"context"
	"testing"

	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/integration/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockEmailSender is a mock implementation of EmailSender
type MockEmailSender struct {
	mock.Mock
}

func (m *MockEmailSender) SendEmail(ctx context.Context, message email.EmailRequest) (*email.EmailResponse, error) {
	args := m.Called(ctx, message)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*email.EmailResponse), args.Error(1)
}

func (m *MockEmailSender) SendRawEmail(ctx context.Context, rawData []byte) (*email.EmailResponse, error) {
	args := m.Called(ctx, rawData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*email.EmailResponse), args.Error(1)
}

func TestEmailService_SendWelcomeEmail(t *testing.T) {
	tests := []struct {
		name          string
		userEmail     string
		userName      string
		setupMock     func(*MockEmailSender)
		expectedError bool
		errorType     string
	}{
		{
			name:      "success - send welcome email",
			userEmail: "john.doe@example.com",
			userName:  "John Doe",
			setupMock: func(m *MockEmailSender) {
				response := &email.EmailResponse{
					MessageID: "msg-123",
					Provider:  "ses",
					Status:    "sent",
				}
				m.On("SendEmail", mock.Anything, mock.MatchedBy(func(req email.EmailRequest) bool {
					return req.Subject == "Welcome to My Echo App!" &&
						len(req.To) == 1 &&
						req.To[0] == "john.doe@example.com" &&
						req.TextBody != "" &&
						req.HTMLBody != ""
				})).Return(response, nil)
			},
			expectedError: false,
		},
		{
			name:      "error - email sender fails",
			userEmail: "john.doe@example.com",
			userName:  "John Doe",
			setupMock: func(m *MockEmailSender) {
				m.On("SendEmail", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "ExternalServiceError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEmailSender := new(MockEmailSender)
			if tt.setupMock != nil {
				tt.setupMock(mockEmailSender)
			}

			service := &EmailService{
				emailSender: mockEmailSender,
			}

			ctx := context.Background()
			err := service.SendWelcomeEmail(ctx, tt.userEmail, tt.userName)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					appErr, ok := err.(*errors.AppError)
					require.True(t, ok, "Expected AppError")
					assert.Equal(t, errors.ErrorTypeExternal, appErr.Type)
				}
			} else {
				require.NoError(t, err)
			}

			mockEmailSender.AssertExpectations(t)
		})
	}
}

func TestEmailService_SendPasswordResetEmail(t *testing.T) {
	tests := []struct {
		name          string
		userEmail     string
		resetToken    string
		setupMock     func(*MockEmailSender)
		expectedError bool
		errorType     string
	}{
		{
			name:       "success - send password reset email",
			userEmail:  "john.doe@example.com",
			resetToken: "reset-token-123",
			setupMock: func(m *MockEmailSender) {
				response := &email.EmailResponse{
					MessageID: "msg-456",
					Provider:  "ses",
					Status:    "sent",
				}
				m.On("SendEmail", mock.Anything, mock.MatchedBy(func(req email.EmailRequest) bool {
					return req.Subject == "Password Reset Request" &&
						len(req.To) == 1 &&
						req.To[0] == "john.doe@example.com" &&
						req.TextBody != "" &&
						req.HTMLBody != "" &&
						contains(req.TextBody, "reset-token-123")
				})).Return(response, nil)
			},
			expectedError: false,
		},
		{
			name:       "error - email sender fails",
			userEmail:  "john.doe@example.com",
			resetToken: "reset-token-123",
			setupMock: func(m *MockEmailSender) {
				m.On("SendEmail", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "ExternalServiceError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEmailSender := new(MockEmailSender)
			if tt.setupMock != nil {
				tt.setupMock(mockEmailSender)
			}

			service := &EmailService{
				emailSender: mockEmailSender,
			}

			ctx := context.Background()
			err := service.SendPasswordResetEmail(ctx, tt.userEmail, tt.resetToken)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					appErr, ok := err.(*errors.AppError)
					require.True(t, ok, "Expected AppError")
					assert.Equal(t, errors.ErrorTypeExternal, appErr.Type)
				}
			} else {
				require.NoError(t, err)
			}

			mockEmailSender.AssertExpectations(t)
		})
	}
}

func TestEmailService_SendNotificationEmail(t *testing.T) {
	tests := []struct {
		name          string
		userEmail     string
		subject       string
		message       string
		setupMock     func(*MockEmailSender)
		expectedError bool
		errorType     string
	}{
		{
			name:      "success - send notification email",
			userEmail: "john.doe@example.com",
			subject:   "Test Notification",
			message:   "This is a test notification message",
			setupMock: func(m *MockEmailSender) {
				response := &email.EmailResponse{
					MessageID: "msg-789",
					Provider:  "ses",
					Status:    "sent",
				}
				m.On("SendEmail", mock.Anything, mock.MatchedBy(func(req email.EmailRequest) bool {
					return req.Subject == "Test Notification" &&
						len(req.To) == 1 &&
						req.To[0] == "john.doe@example.com" &&
						req.TextBody == "This is a test notification message" &&
						req.HTMLBody != ""
				})).Return(response, nil)
			},
			expectedError: false,
		},
		{
			name:      "error - email sender fails",
			userEmail: "john.doe@example.com",
			subject:   "Test Notification",
			message:   "This is a test notification message",
			setupMock: func(m *MockEmailSender) {
				m.On("SendEmail", mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedError: true,
			errorType:     "ExternalServiceError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEmailSender := new(MockEmailSender)
			if tt.setupMock != nil {
				tt.setupMock(mockEmailSender)
			}

			service := &EmailService{
				emailSender: mockEmailSender,
			}

			ctx := context.Background()
			err := service.SendNotificationEmail(ctx, tt.userEmail, tt.subject, tt.message)

			if tt.expectedError {
				require.Error(t, err)
				if tt.errorType != "" {
					appErr, ok := err.(*errors.AppError)
					require.True(t, ok, "Expected AppError")
					assert.Equal(t, errors.ErrorTypeExternal, appErr.Type)
				}
			} else {
				require.NoError(t, err)
			}

			mockEmailSender.AssertExpectations(t)
		})
	}
}

func TestProvideEmailService(t *testing.T) {
	svc := ProvideEmailService(new(MockEmailSender))
	assert.NotNil(t, svc.emailSender)
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
