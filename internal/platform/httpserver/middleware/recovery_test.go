package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/platform/httpserver"
)

func TestRecoveryLogsCauseWithoutExposingIt(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	defer slog.SetDefault(previous)
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))

	router := httpserver.New(nil)
	router.Use(RequestID(), Logging())
	router.Use(httpserver.RequestTimeout(time.Second), Recovery())
	router.Get("/panic", func(c fiber.Ctx) error {
		panic("specific panic cause")
	})
	request := httptest.NewRequest("GET", "/panic", nil)
	request.Header.Set("X-Request-ID", "panic-test")
	response, err := router.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusInternalServerError || strings.Contains(string(body), "specific panic cause") {
		t.Fatalf("unexpected response: %d %s", response.StatusCode, body)
	}

	for _, expected := range []string{"specific panic cause", "stack", "recovery_test.go", "panic-test"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("log missing %q: %s", expected, output.String())
		}
	}
}
