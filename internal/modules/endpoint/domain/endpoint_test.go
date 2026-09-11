package domain_test

import (
	"encoding/json"
	"testing"

	"mockapi/internal/modules/endpoint/domain"
)

func TestNormalizePath(t *testing.T) {
	for _, tc := range []struct {
		input, want string
	}{
		{"/users", "/users"},
		{"users", "/users"},
		{"/users/", "/users"},
		{"  /users/123/  ", "/users/123"},
		{"//users//profile///", "/users/profile"},
		{"///", "/"},
		{"/", "/"},
		{"", "/"},
	} {
		if got := domain.NormalizePath(tc.input); got != tc.want {
			t.Fatalf("NormalizePath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestValidation(t *testing.T) {
	m, err := domain.ValidateMethod("get")
	if err != nil || m != "GET" {
		t.Fatalf("unexpected method: %v, %s", err, m)
	}
	mPatch, err := domain.ValidateMethod("patch")
	if err != nil || mPatch != "PATCH" {
		t.Fatalf("unexpected method: %v, %s", err, mPatch)
	}
	if _, err := domain.ValidateMethod(""); err == nil {
		t.Fatal("expected error on empty method")
	}
	if _, err := domain.ValidateMethod("INVALID_METHOD"); err == nil {
		t.Fatal("expected error on unsupported method")
	}

	p, err := domain.ValidatePath("users")
	if err != nil || p != "/users" {
		t.Fatalf("unexpected path: %v, %s", err, p)
	}
	if _, err := domain.ValidatePath("/"); err == nil {
		t.Fatal("expected error on root path")
	}

	s, err := domain.ValidateStatus(0)
	if err != nil || s != 200 {
		t.Fatalf("unexpected status default: %v, %d", err, s)
	}
	if _, err := domain.ValidateStatus(99); err == nil {
		t.Fatal("expected error on status 99")
	}

	r, err := domain.ValidateResponse(nil)
	if err != nil || string(r) != "null" {
		t.Fatalf("unexpected response default: %v, %s", err, string(r))
	}
	if _, err := domain.ValidateResponse(json.RawMessage("invalid-json")); err == nil {
		t.Fatal("expected error on invalid json response")
	}
}
