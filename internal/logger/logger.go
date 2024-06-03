package logger

import (
	"go.uber.org/zap"
)

type ServerLogger struct {
	*zap.Logger
}

func Initialize(level string) (*ServerLogger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &ServerLogger{logger}, nil
}
