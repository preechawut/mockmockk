package main

import (
	"log/slog"
	"os"

	"mockapi/internal/app"
	"mockapi/internal/config"
	"mockapi/internal/platform/logger"
)

func main() {
	bootstrap := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cfg, err := config.Load()
	if err != nil {
		bootstrap.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	appLogger, err := logger.New(os.Stdout, cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		bootstrap.Error("invalid logging configuration", "error", err)
		os.Exit(1)
	}
	appLogger = appLogger.With("service", cfg.ServiceName, "environment", cfg.Environment)
	if err := app.Run(cfg, appLogger); err != nil {
		appLogger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
