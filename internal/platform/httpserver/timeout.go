package httpserver

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/timeout"
)

func RequestTimeout(duration time.Duration) fiber.Handler {
	return timeout.New(func(c fiber.Ctx) error { return c.Next() }, timeout.Config{
		Timeout: duration,
		OnTimeout: func(c fiber.Ctx) error {
			return ErrorStatus(c, fiber.StatusServiceUnavailable, "request_timeout", "request timed out")
		},
	})
}
