package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailMessage_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		msg      *EmailMessage
		expected bool
	}{
		{
			name: "valid with body",
			msg: &EmailMessage{
				To:      []string{"a@example.com"},
				Subject: "Hi",
				Body:    "text",
			},
			expected: true,
		},
		{
			name: "valid with html body",
			msg: &EmailMessage{
				To:       []string{"a@example.com"},
				Subject:  "Hi",
				HTMLBody: "<p>hi</p>",
			},
			expected: true,
		},
		{
			name: "missing to",
			msg: &EmailMessage{
				To:      nil,
				Subject: "Hi",
				Body:    "text",
			},
			expected: false,
		},
		{
			name: "empty to",
			msg: &EmailMessage{
				To:      []string{},
				Subject: "Hi",
				Body:    "text",
			},
			expected: false,
		},
		{
			name: "missing subject",
			msg: &EmailMessage{
				To:      []string{"a@example.com"},
				Subject: "",
				Body:    "text",
			},
			expected: false,
		},
		{
			name: "missing body and html",
			msg: &EmailMessage{
				To:      []string{"a@example.com"},
				Subject: "Hi",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.msg.IsValid())
		})
	}
}

func TestEmailTemplate_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		tmpl     *EmailTemplate
		expected bool
	}{
		{
			name: "valid with body",
			tmpl: &EmailTemplate{
				Name:    "welcome",
				Subject: "Welcome",
				Body:    "hello",
			},
			expected: true,
		},
		{
			name: "valid with html body",
			tmpl: &EmailTemplate{
				Name:     "welcome",
				Subject:  "Welcome",
				HTMLBody: "<p>hello</p>",
			},
			expected: true,
		},
		{
			name: "missing name",
			tmpl: &EmailTemplate{
				Name:    "",
				Subject: "Welcome",
				Body:    "hello",
			},
			expected: false,
		},
		{
			name: "missing subject",
			tmpl: &EmailTemplate{
				Name:    "welcome",
				Subject: "",
				Body:    "hello",
			},
			expected: false,
		},
		{
			name: "missing body and html",
			tmpl: &EmailTemplate{
				Name:    "welcome",
				Subject: "Welcome",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.tmpl.IsValid())
		})
	}
}
