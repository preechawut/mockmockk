package http_test

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	controllerhttp "mockapi/internal/controller/http"
	"mockapi/internal/controller/http/middleware"
)

type routeRegistrarFunc func(router fiber.Router)

func (f routeRegistrarFunc) RegisterRoutes(router fiber.Router) {
	f(router)
}

func TestRouter(t *testing.T) {
	mockRoute := routeRegistrarFunc(func(router fiber.Router) {
		router.Get("/custom", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"mounted": true})
		})
	})

	router := controllerhttp.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), func(context.Context) error { return nil }, mockRoute)
	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{
		{"GET", "/healthz", 200, `"status":"ok"`},
		{"GET", "/api/v1/custom", 200, `"mounted":true`},
		{"GET", "/missing", 404, ""},
		{"POST", "/healthz", 405, ""},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			resp, err := router.Test(req, -1)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.status {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tc.status)
			}
			bodyBytes, _ := io.ReadAll(resp.Body)
			if !strings.Contains(string(bodyBytes), tc.body) {
				t.Fatalf("unexpected body: %s", string(bodyBytes))
			}
			if tc.status == 200 && !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
				t.Fatal("expected JSON content type")
			}
		})
	}
}

func TestRecoverPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          controllerhttp.HTTPErrorHandler(logger),
	})
	app.Use(middleware.RecoverPanic(logger))
	app.Get("/", func(c *fiber.Ctx) error {
		panic("private detail")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 500 || strings.Contains(string(bodyBytes), "private detail") {
		t.Fatalf("unsafe error response: %d %s", resp.StatusCode, string(bodyBytes))
	}
}

func TestReadinessFailure(t *testing.T) {
	router := controllerhttp.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), func(context.Context) error { return context.DeadlineExceeded })
	req := httptest.NewRequest("GET", "/readyz", nil)
	resp, err := router.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 503 {
		t.Fatalf("got %d", resp.StatusCode)
	}
}

func TestReadinessSuccessWithPoolStats(t *testing.T) {
	poolMetrics := map[string]int{"total_conns": 5, "idle_conns": 3}
	router := controllerhttp.NewRouterWithOptions(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		func(context.Context) error { return nil },
		controllerhttp.Options{
			Ready:     func() bool { return true },
			PoolStats: func() any { return poolMetrics },
		},
	)
	req := httptest.NewRequest("GET", "/readyz", nil)
	resp, err := router.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("got status %d", resp.StatusCode)
	}
}
