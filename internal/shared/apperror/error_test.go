package apperror

import (
	"errors"
	"testing"
)

func TestAppError(t *testing.T) {
	err := New(NotFound, "item not found")
	if err == nil {
		t.Fatal("expected non-nil error")
	}

	if err.Error() != "item not found" {
		t.Fatalf("expected message 'item not found', got %q", err.Error())
	}

	if !Is(err, NotFound) {
		t.Fatal("expected Is(err, NotFound) to be true")
	}

	if Is(err, Invalid) {
		t.Fatal("expected Is(err, Invalid) to be false")
	}

	stdErr := errors.New("standard error")
	if Is(stdErr, NotFound) {
		t.Fatal("expected Is(stdErr, NotFound) to be false")
	}
}
