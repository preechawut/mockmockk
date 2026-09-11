package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"mockapi/internal/modules/endpoint/repository/memory"
	"mockapi/internal/modules/endpoint/service"
)

func TestServiceCRUD(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	svc := service.NewService(repo)

	created, err := svc.Create(ctx, "get", "users", 200, json.RawMessage(`{"users":[]}`))
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}
	if created.Method != "GET" || created.Path != "/users" || created.Status != 200 {
		t.Fatalf("unexpected created mockmock: %+v", created)
	}

	if _, err := svc.Create(ctx, "GET", "/users", 200, nil); err == nil {
		t.Fatal("expected conflict on duplicate method and path")
	}

	fetched, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if fetched.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, fetched.ID)
	}

	matched, err := svc.Match(ctx, "GET", "/users")
	if err != nil {
		t.Fatalf("unexpected match error: %v", err)
	}
	if matched.ID != created.ID {
		t.Fatalf("expected matched ID %s, got %s", created.ID, matched.ID)
	}

	if _, err := svc.Match(ctx, "POST", "/users"); err == nil {
		t.Fatal("expected match not found for POST /users")
	}

	updated, err := svc.Update(ctx, created.ID, "POST", "/users", 201, json.RawMessage(`{"created":true}`))
	if err != nil {
		t.Fatalf("unexpected update error: %v", err)
	}
	if updated.Method != "POST" || updated.Status != 201 {
		t.Fatalf("unexpected updated fields: %+v", updated)
	}

	dup, err := svc.Duplicate(ctx, created.ID)
	if err != nil {
		t.Fatalf("unexpected duplicate error: %v", err)
	}
	if dup.Path != "/users-copy" || dup.Method != "POST" {
		t.Fatalf("unexpected duplicate path: %s", dup.Path)
	}

	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 items, got %d", len(list))
	}

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if _, err := svc.GetByID(ctx, created.ID); err == nil {
		t.Fatal("expected not found after deletion")
	}
}
