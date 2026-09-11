package apperror

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestAppError(t *testing.T) {
	baseErr := errors.New("underlying sql failure")
	appErr := Wrap(baseErr, 404, "NOT_FOUND", "mockmock not found")

	if appErr.Status != 404 {
		t.Fatalf("expected status 404, got %d", appErr.Status)
	}
	if !errors.Is(appErr, baseErr) {
		t.Fatal("expected unwrap to match baseErr")
	}

	sentinel := NotFound("NOT_FOUND", "mockmock not found")
	wrappedSentinel := fmt.Errorf("context: %w", sentinel)

	if !errors.Is(wrappedSentinel, sentinel) {
		t.Fatal("expected wrapped sentinel to match sentinel")
	}

	var extracted *Error
	if !errors.As(wrappedSentinel, &extracted) {
		t.Fatal("expected errors.As to succeed")
	}
	if extracted.Code != "NOT_FOUND" || extracted.Message != "mockmock not found" {
		t.Fatalf("unexpected extracted error: %+v", extracted)
	}
}

func TestInternalAndUnwrap(t *testing.T) {
	cause := errors.New("private database connection")
	internal := Internal("failed to load resource", cause)

	if internal.Status != 500 || internal.Code != "INTERNAL_ERROR" {
		t.Fatalf("unexpected internal error fields: %+v", internal)
	}
	if !errors.Is(internal, cause) {
		t.Fatal("error chain lost")
	}
	if !strings.Contains(internal.Error(), cause.Error()) {
		t.Fatal("missing diagnostic cause in Error() string")
	}
}

func TestIsRequiresExactStatusAndCode(t *testing.T) {
	err := BadRequest("INVALID_INPUT", "invalid input")
	if !errors.Is(err, New(400, "INVALID_INPUT", "same identity")) {
		t.Fatal("expected matching status and code")
	}
	if errors.Is(err, New(500, "INVALID_INPUT", "different status")) {
		t.Fatal("unexpected match on different status")
	}
	if errors.Is(err, New(400, "OTHER_CODE", "different code")) {
		t.Fatal("unexpected match on different code")
	}
}
