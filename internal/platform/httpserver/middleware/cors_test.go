package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/platform/httpserver"
)

func TestCORS(t *testing.T) {
	router := httpserver.New(nil)
	router.Use(CORS([]string{"https://app.example.com"}))
	router.Get("/test", func(c fiber.Ctx) error { return c.SendString("ok") })
	for _, test := range []struct {
		method, origin string
		status         int
	}{
		{"GET", "https://app.example.com", http.StatusOK},
		{"OPTIONS", "https://app.example.com", http.StatusNoContent},
		{"GET", "https://evil.example.com", http.StatusForbidden},
		{"GET", "", http.StatusOK},
	} {
		request := httptest.NewRequest(test.method, "/test", nil)
		request.Header.Set("Origin", test.origin)
		request.Header.Set("Access-Control-Request-Method", "GET")
		request.Header.Set("Access-Control-Request-Headers", "X-Request-ID")
		response, err := router.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != test.status {
			t.Fatalf("%s %s status=%d body=%s", test.method, test.origin, response.StatusCode, body)
		}
		if test.origin == "https://app.example.com" && response.Header.Get("Access-Control-Allow-Origin") != test.origin {
			t.Fatal("missing allow origin")
		}
		if test.status == http.StatusForbidden && response.Header.Get("Access-Control-Allow-Origin") != "" {
			t.Fatal("disallowed origin was exposed")
		}
		if test.method == "OPTIONS" && !strings.Contains(strings.ToLower(response.Header.Get("Access-Control-Allow-Headers")), "x-request-id") {
			t.Fatalf("preflight headers=%v", response.Header)
		}
	}
}
