package http_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	controllerhttp "mockapi/internal/controller/http"
	endpointhttp "mockapi/internal/modules/endpoint/controller/http"
	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/modules/endpoint/repository/memory"
	"mockapi/internal/modules/endpoint/service"
)

type unavailableRepository struct{ service.Repository }

func (unavailableRepository) FindMatch(context.Context, string, string) (domain.Endpoint, error) {
	return domain.Endpoint{}, errors.New("private database address and query details")
}

func TestMockStorageFailureIsSanitized500(t *testing.T) {
	repo := unavailableRepository{Repository: memory.NewRepository()}
	handler := endpointhttp.NewHandler(service.NewService(repo))
	var logs bytes.Buffer
	app := controllerhttp.NewRouter(slog.New(slog.NewTextHandler(&logs, nil)), nil, handler)
	response, err := app.Test(httptest.NewRequest("GET", "/mock/users", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 500 || !strings.Contains(string(body), "INTERNAL_ERROR") {
		t.Fatalf("got %d %s", response.StatusCode, body)
	}
	if !strings.Contains(logs.String(), "private database") {
		t.Fatalf("missing cause in log: %s", logs.String())
	}
	if strings.Contains(string(body), "private") {
		t.Fatalf("leaked storage failure: %s", body)
	}
}
