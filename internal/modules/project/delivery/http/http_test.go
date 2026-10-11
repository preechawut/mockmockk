package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"mockapi/internal/modules/project/domain"
	"mockapi/internal/modules/project/service"
	"mockapi/internal/platform/httpserver"
	"mockapi/internal/shared/apperror"
)

type mockProjectRepo struct {
	projects map[string]domain.Project
}

func (m *mockProjectRepo) Create(ctx context.Context, p domain.Project) (domain.Project, error) {
	m.projects[p.ID] = p
	return p, nil
}

func (m *mockProjectRepo) List(ctx context.Context) ([]domain.Project, error) {
	res := make([]domain.Project, 0, len(m.projects))
	for _, p := range m.projects {
		res = append(res, p)
	}
	return res, nil
}

func (m *mockProjectRepo) Get(ctx context.Context, id string) (domain.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return domain.Project{}, apperror.New(apperror.NotFound, "project not found")
	}
	return p, nil
}

func (m *mockProjectRepo) GetBySlug(ctx context.Context, slug string) (domain.Project, error) {
	for _, p := range m.projects {
		if p.Slug == slug {
			return p, nil
		}
	}
	return domain.Project{}, apperror.New(apperror.NotFound, "project not found")
}

func (m *mockProjectRepo) Rename(ctx context.Context, id, name string) (domain.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return domain.Project{}, apperror.New(apperror.NotFound, "project not found")
	}
	p.Name = name
	m.projects[id] = p
	return p, nil
}

func (m *mockProjectRepo) Delete(ctx context.Context, id string) error {
	delete(m.projects, id)
	return nil
}

func TestProjectHandler(t *testing.T) {
	repo := &mockProjectRepo{projects: make(map[string]domain.Project)}
	svc := service.New(repo)
	handler := New(svc)

	app := httpserver.New(nil)
	app.Post("/projects", handler.Create)
	app.Get("/projects", handler.List)
	app.Get("/projects/:projectId", handler.Get)
	app.Patch("/projects/:projectId", handler.Rename)
	app.Delete("/projects/:projectId", handler.Delete)

	// Create
	body := `{"name":"Demo","slug":"demo-proj"}`
	req := httptest.NewRequest("POST", "/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var createResp struct {
		OK   bool            `json:"ok"`
		Data ProjectResponse `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	projID := createResp.Data.ID
	if projID == "" {
		t.Fatal("expected non-empty ID")
	}

	// Get
	req = httptest.NewRequest("GET", "/projects/"+projID, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Rename
	renameBody := `{"name":"Demo Updated"}`
	req = httptest.NewRequest("PATCH", "/projects/"+projID, bytes.NewBufferString(renameBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Delete
	req = httptest.NewRequest("DELETE", "/projects/"+projID, nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Invalid UUID
	req = httptest.NewRequest("GET", "/projects/invalid-uuid", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid UUID, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	_ = uuid.New()
}
