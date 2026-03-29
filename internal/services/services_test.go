package services

import (
	"os"
	"testing"

	"golang-boilerplate/internal/logger"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// logger.Init is only called from cmd/server; tests need a non-nil logger.
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}
