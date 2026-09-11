package service

import (
	"context"
	"errors"
	"mockapi/internal/modules/project/domain"
	"testing"
)

type failingRepository struct {
	Repository
	failure error
}

func (r failingRepository) Create(context.Context, domain.Project) (domain.Project, error) {
	return domain.Project{}, r.failure
}

func TestCreateErrorPropagation(t *testing.T) {
	cause := errors.New("private database failure")
	svc := NewService(failingRepository{failure: cause})
	_, err := svc.Create(context.Background(), "Demo", "demo")
	if !errors.Is(err, cause) {
		t.Fatalf("expected cause, got: %v", err)
	}
	svc = NewService(failingRepository{failure: domain.ErrConflict})
	_, err = svc.Create(context.Background(), "Demo", "demo")
	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got: %v", err)
	}
}
