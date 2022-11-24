package logger

import (
	"go.uber.org/zap"
)

func init() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	zap.ReplaceGlobals(logger)
}
