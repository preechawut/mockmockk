package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"mockapi/internal/config"
	controllerhttp "mockapi/internal/controller/http"
	endpointhttp "mockapi/internal/modules/endpoint/controller/http"
	"mockapi/internal/modules/endpoint/repository/postgres"
	"mockapi/internal/modules/endpoint/service"
	projecthttp "mockapi/internal/modules/project/controller/http"
	projectpostgres "mockapi/internal/modules/project/repository/postgres"
	projectservice "mockapi/internal/modules/project/service"
	pkgpostgres "mockapi/internal/platform/postgres"
	"mockapi/internal/web"
)

// dependency injection
func buildRoutes(pool *pgxpool.Pool) []controllerhttp.RouteRegistrar {
	repo := postgres.NewRepository(pool)
	svc := service.NewService(repo)
	projects := projectpostgres.NewRepository(pool)
	lookup := projectLookup{repo: projects}
	handler := endpointhttp.NewHandler(svc, lookup)
	return []controllerhttp.RouteRegistrar{
		projecthttp.NewHandler(projectservice.NewService(projects)),
		handler,
	}
}

func NewRouter(cfg config.Config, logger *slog.Logger, pool *pgxpool.Pool, ready func() bool) *fiber.App {
	routes := buildRoutes(pool)
	routes = append(routes, web.NewHandler())

	var ping func(context.Context) error
	var poolStats func() any
	if pool != nil {
		ping = pool.Ping
		poolStats = func() any { return pkgpostgres.Stats(pool) }
	} else {
		ping = func(context.Context) error { return fmt.Errorf("database pool is unavailable") }
	}

	if ready == nil {
		ready = func() bool { return true }
	}

	return controllerhttp.NewRouterWithOptions(logger, ping, controllerhttp.Options{
		RequestTimeout:     cfg.RequestTimeout,
		ReadTimeout:        cfg.HTTPReadTimeout,
		WriteTimeout:       cfg.HTTPWriteTimeout,
		Ready:              ready,
		PoolStats:          poolStats,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
	}, routes...)
}

// graceful shutdown
func Run(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pkgpostgres.New(ctx, cfg.DatabaseURL,
		pkgpostgres.WithMaxConns(cfg.DBMaxConns),
		pkgpostgres.WithMinConns(cfg.DBMinConns),
		pkgpostgres.WithMaxConnLifetime(cfg.DBMaxConnLifetime),
		pkgpostgres.WithMaxConnIdleTime(cfg.DBMaxConnIdleTime),
	)
	if err != nil {
		return err
	}
	defer pool.Close()

	router := NewRouter(cfg, logger, pool, func() bool { return ctx.Err() == nil })

	errChan := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", cfg.HTTPAddr)
		if err := router.Listen(cfg.HTTPAddr); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		logger.Info("shutting down server")
		return router.ShutdownWithTimeout(cfg.ShutdownTimeout)
	}
}
