package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertAccented(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "ascii unchanged except lowercase",
			input:    "Hello World",
			expected: "hello world",
		},
		{
			name:     "vietnamese a variants",
			input:    "àáạãảăắằẳẵặâấầẩẫậ",
			expected: "aaaaaaaaaaaaaaaaa",
		},
		{
			name:     "vietnamese e variants",
			input:    "èẻẽéẹêềểễếệ",
			expected: "eeeeeeeeeee",
		},
		{
			name:     "vietnamese i variants",
			input:    "ìỉĩíị",
			expected: "iiiii",
		},
		{
			name:     "vietnamese o variants",
			input:    "òỏõóọôồổỗốộơờởỡớợ",
			expected: "ooooooooooooooooo",
		},
		{
			name:     "vietnamese u variants",
			input:    "ùủũúụưừửữứự",
			expected: "uuuuuuuuuuu",
		},
		{
			name:     "vietnamese y variants",
			input:    "ỳỷỹýỵ",
			expected: "yyyyy",
		},
		{
			name:     "vietnamese d",
			input:    "đ",
			expected: "d",
		},
		{
			name:     "mixed phrase",
			input:    "Xin Chào Việt Nam",
			expected: "xin chao viet nam",
		},
		{
			name:     "uppercase accents lowered then converted",
			input:    "ĐẶC BIỆT",
			expected: "dac biet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ConvertAccented(tt.input))
		})
	}
}
