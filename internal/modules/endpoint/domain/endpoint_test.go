package domain

import (
	"encoding/json"
	"testing"
)

func TestValidateMethod(t *testing.T) {
	cases := []struct {
		method  string
		want    string
		wantErr bool
	}{
		{"get", "GET", false},
		{"POST", "POST", false},
		{" delete ", "DELETE", false},
		{"", "", true},
		{"INVALID", "", true},
	}

	for _, tc := range cases {
		got, err := ValidateMethod(tc.method)
		if (err != nil) != tc.wantErr {
			t.Fatalf("ValidateMethod(%q) error = %v, wantErr = %v", tc.method, err, tc.wantErr)
		}
		if got != tc.want {
			t.Fatalf("ValidateMethod(%q) = %q, want %q", tc.method, got, tc.want)
		}
	}
}

func TestValidatePath(t *testing.T) {
	cases := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"users", "/users", false},
		{"/users/", "/users", false},
		{"/users/123", "/users/123", false},
		{"/", "", true},
		{"", "", true},
	}

	for _, tc := range cases {
		got, err := ValidatePath(tc.input)
		if (err != nil) != tc.wantErr {
			t.Fatalf("ValidatePath(%q) error = %v, wantErr = %v", tc.input, err, tc.wantErr)
		}
		if got != tc.want {
			t.Fatalf("ValidatePath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestValidateStatus(t *testing.T) {
	cases := []struct {
		status  int
		want    int
		wantErr bool
	}{
		{0, 200, false},
		{200, 200, false},
		{201, 201, false},
		{500, 500, false},
		{99, 0, true},
		{600, 0, true},
	}

	for _, tc := range cases {
		got, err := ValidateStatus(tc.status)
		if (err != nil) != tc.wantErr {
			t.Fatalf("ValidateStatus(%d) error = %v, wantErr = %v", tc.status, err, tc.wantErr)
		}
		if got != tc.want {
			t.Fatalf("ValidateStatus(%d) = %d, want %d", tc.status, got, tc.want)
		}
	}
}

func TestValidateResponse(t *testing.T) {
	cases := []struct {
		input   string
		wantErr bool
	}{
		{`{"ok": true}`, false},
		{`[1, 2, 3]`, false},
		{``, false},
		{`null`, false},
		{`invalid-json`, true},
	}

	for _, tc := range cases {
		_, err := ValidateResponse(json.RawMessage(tc.input))
		if (err != nil) != tc.wantErr {
			t.Fatalf("ValidateResponse(%q) error = %v, wantErr = %v", tc.input, err, tc.wantErr)
		}
	}
}
