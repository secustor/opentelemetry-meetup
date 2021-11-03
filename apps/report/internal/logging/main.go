package logging

import (
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func GetRootLogger() *zap.SugaredLogger {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	return logger.Sugar()
}
