package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestRequestTimeoutCancelsContext(t *testing.T) {
	router := New(nil)
	router.Use(ResponseHeaders(), RequestTimeout(10*time.Millisecond))
	finished := make(chan struct{})
	router.Get("/", func(c fiber.Ctx) error {
		defer close(finished)
		<-c.Context().Done()
		return c.Context().Err()
	})
	response, err := router.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", response.StatusCode)
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("context not canceled")
	}
}
