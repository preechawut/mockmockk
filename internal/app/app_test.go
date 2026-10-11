package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mockapi/internal/config"
)

func TestAppHealthz(t *testing.T) {
	router := New(nil, config.Default())
	req := httptest.NewRequest("GET", "/healthz", nil)
	resp, err := router.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
