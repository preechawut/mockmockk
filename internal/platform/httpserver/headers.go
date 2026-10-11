package httpserver

import "github.com/gofiber/fiber/v3"

func ResponseHeaders() fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("Referrer-Policy", "no-referrer")
		c.Set("Cache-Control", "no-store")
		return c.Next()
	}
}
