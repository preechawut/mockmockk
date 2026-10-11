package domain

import (
	"encoding/json"
	"path"
	"strings"
	"time"

	"mockapi/internal/shared/apperror"
)

type Endpoint struct {
	ID        string
	ProjectID *string
	Method    string
	Path      string
	Status    int
	Response  json.RawMessage
	CreatedAt time.Time
}

var validMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"PATCH":   true,
	"DELETE":  true,
	"HEAD":    true,
	"OPTIONS": true,
}

func NormalizePath(p string) string {
	p = strings.TrimSpace(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = path.Clean(p)
	if len(p) > 1 {
		p = strings.TrimRight(p, "/")
	}
	return p
}

func ValidateMethod(method string) (string, error) {
	m := strings.ToUpper(strings.TrimSpace(method))
	if m == "" {
		return "", apperror.New(apperror.Invalid, "method is required")
	}
	if !validMethods[m] {
		return "", apperror.New(apperror.Invalid, "unsupported HTTP method")
	}
	return m, nil
}

func ValidatePath(path string) (string, error) {
	p := NormalizePath(path)
	if p == "/" {
		return "", apperror.New(apperror.Invalid, "path is required")
	}
	return p, nil
}

func ValidateStatus(status int) (int, error) {
	if status == 0 {
		return 200, nil
	}
	if status < 100 || status > 599 {
		return 0, apperror.New(apperror.Invalid, "status must be between 100 and 599")
	}
	return status, nil
}

func ValidateResponse(resp json.RawMessage) (json.RawMessage, error) {
	if len(resp) == 0 || string(resp) == "" {
		return json.RawMessage("null"), nil
	}
	if !json.Valid(resp) {
		return nil, apperror.New(apperror.Invalid, "response must be valid JSON")
	}
	return resp, nil
}
