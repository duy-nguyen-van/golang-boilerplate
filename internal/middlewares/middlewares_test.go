package middlewares

import (
	"os"
	"testing"

	"golang-boilerplate/internal/logger"

	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}
