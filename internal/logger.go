package internal

import (
	"go.uber.org/zap"
	"os"
)

func NewZapLogger() (*zap.Logger, func(logger *zap.Logger)) {
	var logger *zap.Logger
	environment := os.Getenv("LOGGER")
	if environment == "JSON" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	zap.ReplaceGlobals(logger)
	syncFunc := func(logger *zap.Logger) {
		_ = logger.Sync()
	}
	return logger, syncFunc
}
