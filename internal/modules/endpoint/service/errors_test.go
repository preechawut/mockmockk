package service

import (
	"context"
	"errors"
	"mockapi/internal/modules/endpoint/domain"
	"testing"
)

type failingLookup struct {
	Repository
	failure error
}

func (r failingLookup) GetByID(context.Context, string) (domain.Endpoint, error) {
	return domain.Endpoint{}, r.failure
}

func TestDuplicateSourceFailure(t *testing.T) {
	cause := errors.New("private database failure")
	svc := NewService(failingLookup{failure: cause})
	_, err := svc.Duplicate(context.Background(), "id")
	if !errors.Is(err, cause) {
		t.Fatalf("expected cause, got: %v", err)
	}
	svc = NewService(failingLookup{failure: domain.ErrNotFound})
	_, err = svc.Duplicate(context.Background(), "id")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("not found reclassified: %v", err)
	}
}
