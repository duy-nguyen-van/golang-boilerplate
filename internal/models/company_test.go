package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompany_TableName(t *testing.T) {
	assert.Equal(t, "companies", Company{}.TableName())
}
