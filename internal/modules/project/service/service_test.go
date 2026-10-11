package service

import (
	"context"
	"errors"
	"testing"

	"mockapi/internal/modules/project/domain"
	"mockapi/internal/shared/apperror"
)

type fakeProjectRepo struct {
	created domain.Project
	renamed domain.Project
	list    []domain.Project
	err     error
}

func (f *fakeProjectRepo) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	f.created = p
	return p, f.err
}

func (f *fakeProjectRepo) List(ctx context.Context) ([]domain.Project, error) {
	return f.list, f.err
}

func (f *fakeProjectRepo) Get(ctx context.Context, id string) (domain.Project, error) {
	if f.err != nil {
		return domain.Project{}, f.err
	}
	for _, p := range f.list {
		if p.ID == id {
			return p, nil
		}
	}
	return domain.Project{}, apperror.New(apperror.NotFound, "not found")
}

func (f *fakeProjectRepo) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	if f.err != nil {
		return domain.Project{}, f.err
	}
	for _, p := range f.list {
		if p.Slug == slug {
			return p, nil
		}
	}
	return domain.Project{}, apperror.New(apperror.NotFound, "not found")
}

func (f *fakeProjectRepo) Rename(ctx context.Context, id, name string) (domain.Project, error) {
	p := domain.Project{ID: id, Name: name}
	f.renamed = p
	return p, f.err
}

func (f *fakeProjectRepo) Delete(ctx context.Context, id string) error {
	return f.err
}

func TestProjectService_Create(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := New(repo)

	// invalid input
	_, err := svc.Create(context.Background(), "", "invalid-slug")
	if err == nil {
		t.Fatal("expected error for empty name")
	}

	// valid input
	res, err := svc.Create(context.Background(), "My Project", "my-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "My Project" || res.Slug != "my-project" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProjectService_Rename(t *testing.T) {
	repo := &fakeProjectRepo{}
	svc := New(repo)

	_, err := svc.Rename(context.Background(), "id-1", "")
	if err == nil {
		t.Fatal("expected error for empty name")
	}

	res, err := svc.Rename(context.Background(), "id-1", "New Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Name != "New Name" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestProjectService_Get(t *testing.T) {
	p := domain.Project{ID: "p1", Name: "P1", Slug: "p-1"}
	repo := &fakeProjectRepo{list: []domain.Project{p}}
	svc := New(repo)

	got, err := svc.Get(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "p1" {
		t.Fatalf("got %+v want %+v", got, p)
	}

	_, err = svc.Get(context.Background(), "missing")
	if !apperror.Is(err, apperror.NotFound) {
		t.Fatalf("expected NotFound, got %v", err)
	}

	repo.err = errors.New("db error")
	_, err = svc.Get(context.Background(), "p1")
	if err == nil {
		t.Fatal("expected db error")
	}
}
