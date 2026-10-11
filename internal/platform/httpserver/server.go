package httpserver

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func New(trustedProxies []string) *fiber.App {
	return fiber.New(fiber.Config{
		Immutable:          true,
		StrictRouting:      true,
		CaseSensitive:      true,
		AppName:            "mockmockk",
		ErrorHandler:       Error,
		BodyLimit:          64 * 1024,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       20 * time.Second,
		IdleTimeout:        60 * time.Second,
		TrustProxy:         true,
		TrustProxyConfig:   fiber.TrustProxyConfig{Proxies: trustedProxies},
		ProxyHeader:        fiber.HeaderXForwardedFor,
		EnableIPValidation: true,
	})
}
