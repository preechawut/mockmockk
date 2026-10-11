package service

import (
	"context"
	"encoding/json"
	"testing"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/shared/apperror"
)

type fakeEndpointRepo struct {
	endpoints map[string]domain.Endpoint
}

func newFakeEndpointRepo() *fakeEndpointRepo {
	return &fakeEndpointRepo{endpoints: make(map[string]domain.Endpoint)}
}

func (f *fakeEndpointRepo) Create(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	f.endpoints[m.ID] = m
	return m, nil
}

func (f *fakeEndpointRepo) List(ctx context.Context) ([]domain.Endpoint, error) {
	res := make([]domain.Endpoint, 0, len(f.endpoints))
	for _, v := range f.endpoints {
		res = append(res, v)
	}
	return res, nil
}

func (f *fakeEndpointRepo) ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error) {
	res := make([]domain.Endpoint, 0)
	for _, v := range f.endpoints {
		if v.ProjectID != nil && *v.ProjectID == projectID {
			res = append(res, v)
		}
	}
	return res, nil
}

func (f *fakeEndpointRepo) GetByID(ctx context.Context, id string) (domain.Endpoint, error) {
	e, ok := f.endpoints[id]
	if !ok {
		return domain.Endpoint{}, apperror.New(apperror.NotFound, "not found")
	}
	return e, nil
}

func (f *fakeEndpointRepo) FindMatch(ctx context.Context, method, path string) (domain.Endpoint, error) {
	for _, v := range f.endpoints {
		if v.ProjectID == nil && v.Method == method && v.Path == path {
			return v, nil
		}
	}
	return domain.Endpoint{}, apperror.New(apperror.NotFound, "not found")
}

func (f *fakeEndpointRepo) FindMatchByProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error) {
	for _, v := range f.endpoints {
		if v.ProjectID != nil && *v.ProjectID == projectID && v.Method == method && v.Path == path {
			return v, nil
		}
	}
	return domain.Endpoint{}, apperror.New(apperror.NotFound, "not found")
}

func (f *fakeEndpointRepo) Update(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	f.endpoints[m.ID] = m
	return m, nil
}

func (f *fakeEndpointRepo) Delete(ctx context.Context, id string) error {
	delete(f.endpoints, id)
	return nil
}

func TestEndpointService_CreateAndMatch(t *testing.T) {
	repo := newFakeEndpointRepo()
	svc := New(repo)

	// Create global endpoint
	created, err := svc.Create(context.Background(), "get", "users", 200, json.RawMessage(`{"users":[]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.Method != "GET" || created.Path != "/users" {
		t.Fatalf("unexpected endpoint: %+v", created)
	}

	// Conflict creation
	_, err = svc.Create(context.Background(), "GET", "/users", 200, nil)
	if !apperror.Is(err, apperror.Conflict) {
		t.Fatalf("expected Conflict, got: %v", err)
	}

	// Match
	matched, err := svc.Match(context.Background(), "GET", "/users")
	if err != nil {
		t.Fatalf("match failed: %v", err)
	}
	if matched.ID != created.ID {
		t.Fatalf("matched wrong endpoint: %+v", matched)
	}

	// Duplicate
	dup, err := svc.Duplicate(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("duplicate failed: %v", err)
	}
	if dup.Path != "/users-copy" {
		t.Fatalf("unexpected duplicate path: %s", dup.Path)
	}
}
