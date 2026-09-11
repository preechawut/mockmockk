package http

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"

	"mockapi/internal/controller/http/middleware"
)

type Options struct {
	RequestTimeout     time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	Ready              func() bool
	PoolStats          func() any
	CORSAllowedOrigins string
}

type RouteRegistrar interface {
	RegisterRoutes(router fiber.Router)
}

type RootRouteRegistrar interface {
	RegisterRootRoutes(router fiber.Router)
}

func NewRouter(logger *slog.Logger, ping func(context.Context) error, routes ...RouteRegistrar) *fiber.App {
	return NewRouterWithOptions(logger, ping, Options{}, routes...)
}

func NewRouterWithOptions(logger *slog.Logger, ping func(context.Context) error, options Options, routes ...RouteRegistrar) *fiber.App {
	if options.RequestTimeout <= 0 {
		options.RequestTimeout = 10 * time.Second
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          HTTPErrorHandler(logger),
		BodyLimit:             MaxBodySize,
		ReadTimeout:           options.ReadTimeout,
		WriteTimeout:          options.WriteTimeout,
	})

	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.CORS(options.CORSAllowedOrigins))
	app.Use(middleware.RequestID())
	app.Use(middleware.AccessLog(logger))
	app.Use(middleware.RecoverPanic(logger))
	app.Use(middleware.RequestTimeout(options.RequestTimeout))

	app.All("/healthz", func(c *fiber.Ctx) error {
		if c.Method() != fiber.MethodGet {
			return c.SendStatus(fiber.StatusMethodNotAllowed)
		}
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/readyz", func(c *fiber.Ctx) error {
		if options.Ready != nil && !options.Ready() {
			return ErrNotReady
		}
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if ping != nil {
			if err := ping(ctx); err != nil {
				return ErrNotReady
			}
		}
		resp := fiber.Map{"status": "ready"}
		if options.PoolStats != nil {
			if stats := options.PoolStats(); stats != nil {
				resp["pool"] = stats
			}
		}
		return c.JSON(resp)
	})

	apiV1 := app.Group("/api/v1")
	for _, route := range routes {
		if route != nil {
			route.RegisterRoutes(apiV1)
			if rootRegistrar, ok := route.(RootRouteRegistrar); ok {
				rootRegistrar.RegisterRootRoutes(app)
			}
		}
	}

	return app
}
