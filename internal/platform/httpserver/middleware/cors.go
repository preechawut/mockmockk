package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"

	"mockapi/internal/shared/apperror"
)

func CORS(origins []string) fiber.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		allowed[origin] = true
	}

	handler := cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return allowed[origin]
		},
		AllowMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodHead,
			fiber.MethodOptions,
		},
		AllowHeaders: []string{
			"Content-Type",
			"X-Request-ID",
			"Idempotency-Key",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
			"Retry-After",
		},
		MaxAge: 600,
	})

	return func(c fiber.Ctx) error {
		origin := c.Get(fiber.HeaderOrigin)
		if origin != "" && !allowed[origin] {
			c.Vary(fiber.HeaderOrigin)
			return apperror.New(apperror.Forbidden, "origin is not allowed")
		}

		return handler(c)
	}
}
