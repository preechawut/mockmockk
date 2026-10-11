package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"mockapi/internal/platform/logger"
)

func RequestID() fiber.Handler {
	return requestid.New()
}

func Logging() fiber.Handler {
	return func(c fiber.Ctx) error {
		started := time.Now()
		log := slog.Default().With("request_id", requestid.FromContext(c))
		c.SetContext(logger.WithContext(c.Context(), log))

		err := c.Next()
		if err != nil {
			err = c.App().ErrorHandler(c, err)
		}

		log.Info("http request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration", time.Since(started),
			"client_ip", c.IP(),
		)
		return err
	}
}
