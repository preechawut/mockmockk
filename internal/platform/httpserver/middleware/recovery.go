package middleware

import (
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
	fiberRecover "github.com/gofiber/fiber/v3/middleware/recover"

	"mockapi/internal/platform/logger"
)

func logPanic(c fiber.Ctx, recovered any) {
	log := logger.FromContext(c.Context())
	log.Error("request panic",
		"panic", recovered,
		"method", c.Method(),
		"path", c.Path(),
		"client_ip", c.IP(),
		"stack", string(debug.Stack()),
	)
}

func Recovery() fiber.Handler {
	return fiberRecover.New(fiberRecover.Config{
		EnableStackTrace:  true,
		StackTraceHandler: logPanic,
	})
}
