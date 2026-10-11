package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/platform/httpserver"
	"mockapi/internal/platform/logger"
)

func TestLoggingJSONFormat(t *testing.T) {
	router, output := newLoggingTestRouter(t, "json", "INFO")
	requestLoggingRoute(t, router, "/ok", http.StatusCreated)

	var entry struct {
		Message   string `json:"msg"`
		RequestID string `json:"request_id"`
		Method    string `json:"method"`
		Path      string `json:"path"`
		Status    int    `json:"status"`
		Duration  *int64 `json:"duration"`
	}
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Message != "http request" {
		t.Fatalf("message=%q", entry.Message)
	}
	if entry.RequestID != "access-test" {
		t.Fatalf("request_id=%q", entry.RequestID)
	}
	if entry.Method != http.MethodGet {
		t.Fatalf("method=%q", entry.Method)
	}
	if entry.Path != "/ok" {
		t.Fatalf("path=%q", entry.Path)
	}
	if entry.Status != http.StatusCreated {
		t.Fatalf("status=%d", entry.Status)
	}
	if entry.Duration == nil {
		t.Fatal("missing duration")
	}
}

func TestLoggingTextFormat(t *testing.T) {
	router, output := newLoggingTestRouter(t, "text", "INFO")
	requestLoggingRoute(t, router, "/ok", http.StatusCreated)

	for _, field := range []string{"request_id=access-test", "method=GET", "path=/ok", "status=201", "duration="} {
		if !strings.Contains(output.String(), field) {
			t.Fatalf("log missing %q: %s", field, output.String())
		}
	}
	if strings.Count(strings.TrimSpace(output.String()), "\n") != 0 {
		t.Fatalf("expected one log: %s", output.String())
	}
}

func TestLoggingRespectsLogLevel(t *testing.T) {
	router, output := newLoggingTestRouter(t, "json", "ERROR")
	requestLoggingRoute(t, router, "/ok", http.StatusCreated)
	if output.Len() != 0 {
		t.Fatalf("INFO log was not filtered: %s", output.String())
	}
}

func TestLoggingRecordsResponseStatus(t *testing.T) {
	cases := []struct {
		name   string
		path   string
		status int
	}{
		{name: "success", path: "/ok", status: http.StatusCreated},
		{name: "handler error", path: "/invalid", status: http.StatusBadRequest},
		{name: "missing route", path: "/missing", status: http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router, output := newLoggingTestRouter(t, "json", "INFO")
			requestLoggingRoute(t, router, tc.path, tc.status)

			var entry struct {
				Status int `json:"status"`
			}
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatal(err)
			}
			if entry.Status != tc.status {
				t.Fatalf("logged status=%d want=%d", entry.Status, tc.status)
			}
		})
	}
}

func newLoggingTestRouter(t *testing.T, format, level string) (*fiber.App, *bytes.Buffer) {
	t.Helper()
	output := new(bytes.Buffer)
	log, err := logger.New(output, format, level)
	if err != nil {
		t.Fatal(err)
	}
	previous := slog.Default()
	slog.SetDefault(log)
	t.Cleanup(func() { slog.SetDefault(previous) })

	router := httpserver.New(nil)
	router.Use(RequestID(), Logging())
	router.Get("/ok", func(c fiber.Ctx) error { return c.SendStatus(http.StatusCreated) })
	router.Get("/invalid", func(c fiber.Ctx) error { return fiber.ErrBadRequest })
	return router, output
}

func requestLoggingRoute(t *testing.T, router *fiber.App, path string, status int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-Request-ID", "access-test")
	resp, err := router.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != status {
		t.Fatalf("response status=%d want=%d", resp.StatusCode, status)
	}
}
