package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	fiberLimiter "github.com/gofiber/fiber/v3/middleware/limiter"

	"mockapi/internal/platform/httpserver"
)

func RateLimit(limit int, window time.Duration) fiber.Handler {
	return fiberLimiter.New(fiberLimiter.Config{
		Max:        limit,
		Expiration: window,
		Next: func(c fiber.Ctx) bool {
			return c.Method() == fiber.MethodOptions || !strings.HasPrefix(c.Path(), "/api/")
		},
		LimitReached: func(c fiber.Ctx) error {
			return httpserver.ErrorStatus(c, fiber.StatusTooManyRequests, "rate_limited", "too many requests; try again later")
		},
	})
}
