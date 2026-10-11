package middleware

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"mockapi/internal/platform/httpserver"
)

func TestRateLimitConcurrentRequests(t *testing.T) {
	router := httpserver.New(nil)
	router.Use(RateLimit(10, time.Minute))
	router.Get("/api/test", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	var allowed atomic.Int32
	var blocked atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/api/test", nil)
			resp, err := router.Test(req)
			if err != nil {
				return
			}
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				allowed.Add(1)
			} else if resp.StatusCode == http.StatusTooManyRequests {
				blocked.Add(1)
			}
		}()
	}
	wg.Wait()

	if allowed.Load() != 10 {
		t.Fatalf("allowed %d, want 10", allowed.Load())
	}
	if blocked.Load() != 40 {
		t.Fatalf("blocked %d, want 40", blocked.Load())
	}
}

func TestRateLimitSkipsProbesAndPreflightAndIgnoresSpoofedIP(t *testing.T) {
	router := httpserver.New(nil)
	router.Use(RateLimit(1, time.Minute))
	router.Get("/api/test", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	router.Get("/healthz", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	router.Options("/api/test", func(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) })

	for i, path := range []string{"/healthz", "/api/test", "/api/test", "/api/test", "/healthz"} {
		method, want := "GET", http.StatusOK
		if i == 1 {
			method, want = "OPTIONS", http.StatusNoContent
		}
		if i == 3 {
			want = http.StatusTooManyRequests
		}
		request := httptest.NewRequest(method, path, nil)
		request.Header.Set("X-Forwarded-For", "192.0.2."+strconv.Itoa(i+1))
		response, err := router.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(response.Body)
		response.Body.Close()
		if response.StatusCode != want {
			t.Fatalf("request %d status=%d want=%d", i, response.StatusCode, want)
		}
		if want == http.StatusTooManyRequests && (response.Header.Get("Retry-After") == "" || !strings.Contains(string(body), `"code":"rate_limited"`)) {
			t.Fatal("missing retry header or envelope")
		}
	}
}

func TestTrustedProxyUsesForwardedClientQuota(t *testing.T) {
	router := httpserver.New([]string{"127.0.0.1"})
	router.Use(func(c fiber.Ctx) error {
		c.RequestCtx().SetRemoteAddr(&net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1234})
		return c.Next()
	})
	router.Use(RateLimit(1, time.Minute))
	router.Get("/api/test", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	for i, ip := range []string{"192.0.2.1", "192.0.2.2", "192.0.2.1"} {
		request := httptest.NewRequest("GET", "/api/test", nil)
		request.Header.Set("X-Forwarded-For", ip)
		response, err := router.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		want := http.StatusOK
		if i == 2 {
			want = http.StatusTooManyRequests
		}
		if response.StatusCode != want {
			t.Fatalf("proxy client %d got %d want %d", i, response.StatusCode, want)
		}
	}
}
