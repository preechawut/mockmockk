package app

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"mockapi/internal/config"
)

func TestMissingDatabaseIsNotReady(t *testing.T) {
	router := NewRouter(config.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil)
	res, err := router.Test(httptest.NewRequest("GET", "/readyz", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 503 {
		t.Fatalf("missing database reported status %d", res.StatusCode)
	}
}
