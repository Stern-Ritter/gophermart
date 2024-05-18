package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/Stern-Ritter/gophermart/internal/app"
	"github.com/Stern-Ritter/gophermart/internal/config"
	"github.com/Stern-Ritter/gophermart/internal/logger"
)

func main() {
	cfg, err := app.GetConfig(config.ServerConfig{
		LoggerLvl: "debug",
	})
	if err != nil {
		log.Fatalf("%+v", err)
	}

	logger, err := logger.Initialize(cfg.LoggerLvl)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	err = app.Run(&cfg, logger)
	if err != nil {
		logger.Fatal("Error starting server", zap.String("event", "start server"), zap.Error(err))
	}
}
