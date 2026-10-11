package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"mockapi/internal/config"
	"mockapi/internal/platform/postgres"
	"mockapi/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", "error", err)
		os.Exit(1)
	}

	slog.Info("migrations applied")
}

func run() error {
	// config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// db
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	return migrations.Apply(ctx, db)
}
