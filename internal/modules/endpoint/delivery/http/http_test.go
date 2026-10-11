package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/modules/endpoint/service"
	projectdomain "mockapi/internal/modules/project/domain"
	"mockapi/internal/platform/httpserver"
	"mockapi/internal/shared/apperror"
)

type mockEndpointRepo struct {
	endpoints map[string]domain.Endpoint
}

func (m *mockEndpointRepo) Create(ctx context.Context, ep domain.Endpoint) (domain.Endpoint, error) {
	m.endpoints[ep.ID] = ep
	return ep, nil
}

func (m *mockEndpointRepo) List(ctx context.Context) ([]domain.Endpoint, error) {
	res := make([]domain.Endpoint, 0, len(m.endpoints))
	for _, ep := range m.endpoints {
		res = append(res, ep)
	}
	return res, nil
}

func (m *mockEndpointRepo) ListByProject(ctx context.Context, projectID string) ([]domain.Endpoint, error) {
	res := make([]domain.Endpoint, 0)
	for _, ep := range m.endpoints {
		if ep.ProjectID != nil && *ep.ProjectID == projectID {
			res = append(res, ep)
		}
	}
	return res, nil
}

func (m *mockEndpointRepo) GetByID(ctx context.Context, id string) (domain.Endpoint, error) {
	ep, ok := m.endpoints[id]
	if !ok {
		return domain.Endpoint{}, apperror.New(apperror.NotFound, "not found")
	}
	return ep, nil
}

func (m *mockEndpointRepo) FindMatch(ctx context.Context, method, path string) (domain.Endpoint, error) {
	for _, ep := range m.endpoints {
		if ep.ProjectID == nil && ep.Method == method && ep.Path == path {
			return ep, nil
		}
	}
	return domain.Endpoint{}, apperror.New(apperror.NotFound, "not found")
}

func (m *mockEndpointRepo) FindMatchByProject(ctx context.Context, projectID, method, path string) (domain.Endpoint, error) {
	for _, ep := range m.endpoints {
		if ep.ProjectID != nil && *ep.ProjectID == projectID && ep.Method == method && ep.Path == path {
			return ep, nil
		}
	}
	return domain.Endpoint{}, apperror.New(apperror.NotFound, "not found")
}

func (m *mockEndpointRepo) Update(ctx context.Context, ep domain.Endpoint) (domain.Endpoint, error) {
	m.endpoints[ep.ID] = ep
	return ep, nil
}

func (m *mockEndpointRepo) Delete(ctx context.Context, id string) error {
	delete(m.endpoints, id)
	return nil
}

type mockProjectLookup struct {
	projects map[string]projectdomain.Project
}

func (m *mockProjectLookup) GetBySlug(ctx context.Context, slug string) (projectdomain.Project, error) {
	for _, p := range m.projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return projectdomain.Project{}, apperror.New(apperror.NotFound, "project not found")
}

func TestEndpointHandlerAndMockServe(t *testing.T) {
	endpointRepo := &mockEndpointRepo{endpoints: make(map[string]domain.Endpoint)}
	endpointSvc := service.New(endpointRepo)
	projectLookup := &mockProjectLookup{
		projects: map[string]projectdomain.Project{
			"p1": {ID: "00000000-0000-0000-0000-000000000001", Slug: "my-demo", Name: "My Demo"},
		},
	}
	handler := New(endpointSvc, projectLookup)

	app := httpserver.New(nil)
	app.Post("/mockmocks", handler.Create)
	app.Get("/mockmocks", handler.List)
	app.All("/mock", handler.ServeMock)
	app.All("/mock/*", handler.ServeMock)

	// Create global endpoint
	createBody := `{"method":"GET","path":"/hello","status":200,"response":{"message":"world"}}`
	req := httptest.NewRequest("POST", "/mockmocks", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Call mock
	req = httptest.NewRequest("GET", "/mock/hello", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !bytes.Contains(bodyBytes, []byte(`"message":"world"`)) {
		t.Fatalf("unexpected mock body: %s", bodyBytes)
	}

	// Call non-existent mock -> 404
	req = httptest.NewRequest("GET", "/mock/not-found", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
