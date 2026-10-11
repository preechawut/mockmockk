package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/config"
)

func Run(ctx context.Context, cfg config.Config, db *sql.DB) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	server := New(db, cfg)
	shutdownResult := make(chan error, 1)

	server.Hooks().OnListen(func(data fiber.ListenData) error {
		slog.Info("api started", "address", net.JoinHostPort(data.Host, data.Port))
		return nil
	})

	server.Hooks().OnPostShutdown(func(err error) error {
		shutdownResult <- err
		return nil
	})

	// start server with graceful shutdown
	err := server.Listen(cfg.HTTPAddr, fiber.ListenConfig{
		DisableStartupMessage: true,
		GracefulContext:       ctx,
		ShutdownTimeout:       10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.HTTPAddr, err)
	}

	// wait for shutdown complete
	if ctx.Err() != nil {
		if err := <-shutdownResult; err != nil {
			return fmt.Errorf("shutdown api: %w", err)
		}
	}

	slog.Info("api stopped")
	return nil
}
