package memory_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/modules/endpoint/repository/memory"
)

func TestRepositoryOwnsStoredJSON(t *testing.T) {
	ctx := context.Background()
	r := memory.NewRepository()
	input := domain.Endpoint{ID: "one", Method: "GET", Path: "/a", Response: json.RawMessage(`{"v":1}`)}
	created, err := r.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	input.Response[5] = '2'
	created.Response[5] = '3'
	for _, read := range []func() (domain.Endpoint, error){
		func() (domain.Endpoint, error) { return r.GetByID(ctx, "one") },
		func() (domain.Endpoint, error) { return r.FindMatch(ctx, "GET", "/a") },
		func() (domain.Endpoint, error) {
			items, err := r.List(ctx)
			if err != nil {
				return domain.Endpoint{}, err
			}
			return items[0], nil
		},
	} {
		got, err := read()
		if err != nil {
			t.Fatal(err)
		}
		if string(got.Response) != `{"v":1}` {
			t.Fatalf("stored JSON mutated: %s", got.Response)
		}
		got.Response[5] = '4'
	}
	updated, err := r.Update(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	input.Response[5] = '5'
	updated.Response[5] = '6'
	got, err := r.GetByID(ctx, "one")
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Response) != `{"v":2}` {
		t.Fatalf("update aliases stored JSON: %s", got.Response)
	}
	input.Path = "/other"
	if _, err := r.Create(ctx, input); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("reused ID overwrote record: %v", err)
	}
}
