package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"mockapi/internal/platform/requestid"
)

func TestAccessLog(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(AccessLog(slog.New(slog.NewTextHandler(io.Discard, nil))))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("unexpected code: %d", resp.StatusCode)
	}
}

func TestAccessLogUsesFinalErrorStatus(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{"code": "NOT_FOUND", "message": "route not found"},
			})
		},
	})
	app.Use(AccessLog(slog.New(slog.NewTextHandler(io.Discard, nil))))
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "route not found")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("unexpected code: %d", resp.StatusCode)
	}
}

func TestRecoverPanic(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{"code": "INTERNAL_ERROR", "message": "internal server error"},
			})
		},
	})
	app.Use(RecoverPanic(logger))
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("boom")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("unexpected code: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "INTERNAL_ERROR") {
		t.Fatalf("unexpected recovery response: %s", string(body))
	}
}

func TestRequestID(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(RequestID())
	app.Get("/id", func(c *fiber.Ctx) error {
		return c.SendString(requestid.FromContext(c.UserContext()))
	})

	req := httptest.NewRequest("GET", "/id", nil)
	req.Header.Set("X-Request-ID", "custom-req-id")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "custom-req-id" {
		t.Fatalf("got %s, want custom-req-id", string(body))
	}
	if resp.Header.Get("X-Request-ID") != "custom-req-id" {
		t.Fatalf("got header %s", resp.Header.Get("X-Request-ID"))
	}

	req2 := httptest.NewRequest("GET", "/id", nil)
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatal(err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	if len(body2) == 0 {
		t.Fatal("expected generated request id")
	}
}

func TestRequestTimeout(t *testing.T) {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if errors.Is(err, context.DeadlineExceeded) {
				return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{"error": "timeout"})
			}
			return c.SendStatus(fiber.StatusInternalServerError)
		},
	})
	app.Use(RequestTimeout(30 * time.Millisecond))
	app.Get("/timeout", func(c *fiber.Ctx) error {
		select {
		case <-time.After(80 * time.Millisecond):
		case <-c.UserContext().Done():
		}
		return nil
	})

	req := httptest.NewRequest("GET", "/timeout", nil)
	resp, err := app.Test(req, 200)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusGatewayTimeout {
		t.Fatalf("expected 504 Gateway Timeout, got %d", resp.StatusCode)
	}
}

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(SecurityHeaders())
	app.Get("/secure", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/secure", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("expected nosniff, got %q", resp.Header.Get("X-Content-Type-Options"))
	}
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("expected DENY, got %q", resp.Header.Get("X-Frame-Options"))
	}
	if resp.Header.Get("Referrer-Policy") != "no-referrer" {
		t.Fatalf("expected no-referrer, got %q", resp.Header.Get("Referrer-Policy"))
	}
}

func TestCORS(t *testing.T) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(CORS("https://example.com"))
	app.Get("/api", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	optionsReq := httptest.NewRequest(fiber.MethodOptions, "/api", nil)
	optionsReq.Header.Set("Origin", "https://example.com")
	resp, err := app.Test(optionsReq)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204 for OPTIONS, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("unexpected allow origin: %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
	if !strings.Contains(resp.Header.Get("Access-Control-Allow-Methods"), "PATCH") {
		t.Fatalf("expected PATCH in Access-Control-Allow-Methods: %q", resp.Header.Get("Access-Control-Allow-Methods"))
	}
	if resp.Header.Get("Access-Control-Max-Age") != "86400" {
		t.Fatalf("expected 86400 in Access-Control-Max-Age: %q", resp.Header.Get("Access-Control-Max-Age"))
	}

	getReq := httptest.NewRequest(fiber.MethodGet, "/api", nil)
	getReq.Header.Set("Origin", "https://example.com")
	getResp, err := app.Test(getReq)
	if err != nil {
		t.Fatal(err)
	}
	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", getResp.StatusCode)
	}
	if getResp.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("unexpected Allow-Origin: %s", getResp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSAllowlist(t *testing.T) {
	app := fiber.New()
	app.Use(CORS("https://a.example, https://b.example"))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })
	for _, origin := range []string{"https://a.example", "https://b.example", "https://evil.example", ""} {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", origin)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		want := origin
		if origin == "https://evil.example" {
			want = ""
		}
		if res.Header.Get("Access-Control-Allow-Origin") != want || !strings.Contains(res.Header.Get("Vary"), "Origin") {
			t.Fatalf("origin %q: %v", origin, res.Header)
		}
	}
}

func TestAccessLogUsesSanitizedRequestID(t *testing.T) {
	var logs bytes.Buffer
	app := fiber.New()
	app.Use(RequestID(), AccessLog(slog.New(slog.NewJSONHandler(&logs, nil))))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(200) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(requestid.Header, strings.Repeat("x", 100))
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var record struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(logs.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if !requestid.Valid(record.RequestID) || record.RequestID != res.Header.Get(requestid.Header) {
		t.Fatalf("log ID mismatch: %s", logs.String())
	}
}
