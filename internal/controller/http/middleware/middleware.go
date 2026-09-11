package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"mockapi/internal/platform/requestid"
)

var ErrPanic = errors.New("panic recovered")

// security headers
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "no-referrer")
		return c.Next()
	}
}

// cors
func CORS(allowedOrigins string) fiber.Handler {
	origins := make(map[string]bool)
	for _, origin := range strings.Split(allowedOrigins, ",") {
		origins[strings.TrimSpace(origin)] = true
	}
	wildcard := allowedOrigins == "" || origins["*"]
	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")
		if !wildcard {
			c.Vary("Origin")
		}
		if origin != "" && (wildcard || origins[origin]) {
			allowed := origin
			if wildcard {
				allowed = "*"
			}
			c.Set("Access-Control-Allow-Origin", allowed)
			c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS,HEAD")
			c.Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Request-ID")
			c.Set("Access-Control-Expose-Headers", "X-Request-ID,Location")
			c.Set("Access-Control-Max-Age", "86400")
		}
		if c.Method() == fiber.MethodOptions {
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
}

// request id
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := requestid.Ensure(c.Get(requestid.Header))
		c.SetUserContext(requestid.WithContext(c.UserContext(), id))
		c.Set(requestid.Header, id)
		return c.Next()
	}
}

// access log
func AccessLog(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		if err != nil {
			err = c.App().ErrorHandler(c, err)
		}
		duration := time.Since(start)

		reqID := requestid.FromContext(c.UserContext())
		logger.Info("http request",
			"request_id", reqID,
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", duration.Milliseconds(),
		)
		return err
	}
}

// recover
func RecoverPanic(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				logger.Error("panic recovered",
					"request_id", requestid.FromContext(c.UserContext()),
					"panic", fmt.Sprint(r),
					"stack", stack,
				)
				err = ErrPanic
			}
		}()
		return c.Next()
	}
}

// timeout
func RequestTimeout(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()
		c.SetUserContext(ctx)
		err := c.Next()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return context.DeadlineExceeded
		}
		return err
	}
}
