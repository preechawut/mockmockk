package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	controllerhttp "mockapi/internal/controller/http"
	endpointhttp "mockapi/internal/modules/endpoint/controller/http"
	"mockapi/internal/modules/endpoint/repository/memory"
	"mockapi/internal/modules/endpoint/service"
)

func setupTestApp() *fiber.App {
	repo := memory.NewRepository()
	svc := service.NewService(repo)
	handler := endpointhttp.NewHandler(svc)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return controllerhttp.NewRouter(logger, nil, handler)
}

func TestEndpointHTTPFlow(t *testing.T) {
	app := setupTestApp()

	createBody := `{"method":"GET","path":"/users","status":200,"response":{"users":["alice","bob"]}}`
	req := httptest.NewRequest("POST", "/api/v1/mockmocks", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("create failed with status %d", resp.StatusCode)
	}
	var created endpointhttp.EndpointResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Path != "/users" {
		t.Fatalf("unexpected created response: %+v", created)
	}

	mockReq := httptest.NewRequest("GET", "/mock/users", nil)
	mockResp, err := app.Test(mockReq)
	if err != nil {
		t.Fatal(err)
	}
	if mockResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 from /mock/users, got %d", mockResp.StatusCode)
	}
	mockBody, _ := io.ReadAll(mockResp.Body)
	if !strings.Contains(string(mockBody), "alice") {
		t.Fatalf("unexpected mock response body: %s", string(mockBody))
	}

	getReq := httptest.NewRequest("GET", "/api/v1/mockmocks/"+created.ID, nil)
	getResp, err := app.Test(getReq)
	if err != nil {
		t.Fatal(err)
	}
	if getResp.StatusCode != fiber.StatusOK {
		t.Fatalf("get failed with status %d", getResp.StatusCode)
	}

	dupReq := httptest.NewRequest("POST", "/api/v1/mockmocks/"+created.ID+"/duplicate", nil)
	dupResp, err := app.Test(dupReq)
	if err != nil {
		t.Fatal(err)
	}
	if dupResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("duplicate failed with status %d", dupResp.StatusCode)
	}

	listReq := httptest.NewRequest("GET", "/api/v1/mockmocks", nil)
	listResp, err := app.Test(listReq)
	if err != nil {
		t.Fatal(err)
	}
	var items []endpointhttp.EndpointResponse
	listBytes, _ := io.ReadAll(listResp.Body)
	if err := json.Unmarshal(listBytes, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	delReq := httptest.NewRequest("DELETE", "/api/v1/mockmocks/"+created.ID, nil)
	delResp, err := app.Test(delReq)
	if err != nil {
		t.Fatal(err)
	}
	if delResp.StatusCode != fiber.StatusOK {
		t.Fatalf("delete failed with status %d", delResp.StatusCode)
	}

	mockNotFoundReq := httptest.NewRequest("GET", "/mock/users", nil)
	mockNotFoundResp, err := app.Test(mockNotFoundReq)
	if err != nil {
		t.Fatal(err)
	}
	if mockNotFoundResp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("expected 404 for deleted mock, got %d", mockNotFoundResp.StatusCode)
	}
}

func TestEndpointPatchUpdate(t *testing.T) {
	app := setupTestApp()

	createBody := `{"method":"GET","path":"/patch-target","status":200,"response":{"status":"initial"}}`
	req := httptest.NewRequest("POST", "/api/v1/mockmocks", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	var created endpointhttp.EndpointResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &created); err != nil {
		t.Fatal(err)
	}

	patchBody := `{"status":202}`
	patchReq := httptest.NewRequest("PATCH", "/api/v1/mockmocks/"+created.ID, bytes.NewBufferString(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, err := app.Test(patchReq)
	if err != nil {
		t.Fatal(err)
	}
	if patchResp.StatusCode != fiber.StatusOK {
		t.Fatalf("PATCH failed with status %d", patchResp.StatusCode)
	}
	var patched endpointhttp.EndpointResponse
	patchRespBytes, _ := io.ReadAll(patchResp.Body)
	if err := json.Unmarshal(patchRespBytes, &patched); err != nil {
		t.Fatal(err)
	}
	if patched.Status != 202 || patched.Path != "/patch-target" || patched.Method != "GET" {
		t.Fatalf("unexpected patched endpoint: %+v", patched)
	}
}

func TestEndpointUnknownFieldsRejected(t *testing.T) {
	app := setupTestApp()

	createBody := `{"method":"GET","path":"/unknown-field","status":200,"extra_field":"invalid"}`
	req := httptest.NewRequest("POST", "/api/v1/mockmocks", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400 for unknown fields, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "INVALID_JSON") {
		t.Fatalf("expected INVALID_JSON error code, got %s", string(body))
	}
}

type stubProjectLookup map[string]string

func (s stubProjectLookup) Resolve(_ context.Context, ref string, _ bool) (string, error) {
	if id, ok := s[ref]; ok {
		return id, nil
	}
	return ref, nil
}

func TestProjectScopedEndpointFlow(t *testing.T) {
	repo := memory.NewRepository()
	svc := service.NewService(repo)
	lookup := stubProjectLookup{"ecommerce": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"}
	handler := endpointhttp.NewHandler(svc, lookup)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	app := controllerhttp.NewRouter(logger, nil, handler)

	pid := "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d"
	createBody := `{"method":"POST","path":"/auth/login","status":200,"response":{"token":"jwt-123","user":{"id":1,"name":"Gain"}}}`
	req := httptest.NewRequest("POST", "/api/v1/projects/"+pid+"/endpoints", bytes.NewBufferString(createBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	listReq := httptest.NewRequest("GET", "/api/v1/projects/"+pid+"/endpoints", nil)
	listResp, err := app.Test(listReq)
	if err != nil {
		t.Fatal(err)
	}
	if listResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", listResp.StatusCode)
	}
	var items []endpointhttp.EndpointResponse
	_ = json.NewDecoder(listResp.Body).Decode(&items)
	if len(items) != 1 || items[0].Path != "/auth/login" {
		t.Fatalf("unexpected list: %+v", items)
	}

	mockReq := httptest.NewRequest("POST", "/mock/ecommerce/auth/login", nil)
	mockResp, err := app.Test(mockReq)
	if err != nil {
		t.Fatal(err)
	}
	mockBody, _ := io.ReadAll(mockResp.Body)
	if mockResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d: %s", mockResp.StatusCode, string(mockBody))
	}
	var resBody map[string]any
	if err := json.Unmarshal(mockBody, &resBody); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if resBody["token"] != "jwt-123" {
		t.Fatalf("unexpected mock response: %+v", resBody)
	}
}
