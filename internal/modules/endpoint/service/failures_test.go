package service_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/modules/endpoint/repository/memory"
	"mockapi/internal/modules/endpoint/service"
)

type failedLookup struct {
	service.Repository
	err    error
	writes int
}

func (r *failedLookup) FindMatch(context.Context, string, string) (domain.Endpoint, error) {
	return domain.Endpoint{}, r.err
}

func (r *failedLookup) Create(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	r.writes++
	return r.Repository.Create(ctx, m)
}

func (r *failedLookup) Update(ctx context.Context, m domain.Endpoint) (domain.Endpoint, error) {
	r.writes++
	return r.Repository.Update(ctx, m)
}

func TestStorageFailureDoesNotBecomeMissingRouteOrWrite(t *testing.T) {
	for _, failure := range []error{errors.New("database unavailable"), context.Canceled, context.DeadlineExceeded} {
		t.Run(failure.Error(), func(t *testing.T) {
			ctx := context.Background()
			repo := memory.NewRepository()
			original, err := service.NewService(repo).Create(ctx, "GET", "/users", 200, nil)
			if err != nil {
				t.Fatal(err)
			}
			broken := &failedLookup{Repository: repo, err: fmt.Errorf("query: %w", failure)}
			svc := service.NewService(broken)
			for name, call := range map[string]func() (domain.Endpoint, error){
				"match":  func() (domain.Endpoint, error) { return svc.Match(ctx, "GET", "/users") },
				"create": func() (domain.Endpoint, error) { return svc.Create(ctx, "POST", "/other", 200, nil) },
				"update": func() (domain.Endpoint, error) { return svc.Update(ctx, original.ID, "", "/other", 0, nil) },
			} {
				_, err := call()
				if !errors.Is(err, failure) || errors.Is(err, domain.ErrNoMatch) {
					t.Fatalf("%s lost storage failure: %v", name, err)
				}
			}
			if broken.writes != 0 {
				t.Fatalf("wrote %d times after failed lookup", broken.writes)
			}
		})
	}
}

type conflictingCreate struct {
	service.Repository
	cancel context.CancelFunc
	err    error
}

func (r conflictingCreate) Create(context.Context, domain.Endpoint) (domain.Endpoint, error) {
	if r.cancel != nil {
		r.cancel()
	}
	return domain.Endpoint{}, r.err
}

func TestDuplicateStopsOnFailureAndCancellation(t *testing.T) {
	repo := memory.NewRepository()
	src, err := service.NewService(repo).Create(context.Background(), "GET", "/users", 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("storage unavailable")
	svc := service.NewService(conflictingCreate{Repository: repo, err: failure})
	if _, err := svc.Duplicate(context.Background(), src.ID); !errors.Is(err, failure) {
		t.Fatalf("lost failure: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc = service.NewService(conflictingCreate{Repository: repo, cancel: cancel, err: domain.ErrConflict})
	if _, err := svc.Duplicate(ctx, src.ID); !errors.Is(err, context.Canceled) {
		t.Fatalf("retry did not cancel: %v", err)
	}
}

func TestConcurrentDuplicatesHaveDistinctRoutes(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewRepository()
	svc := service.NewService(repo)
	src, err := svc.Create(ctx, "GET", "/users", 200, nil)
	if err != nil {
		t.Fatal(err)
	}
	const workers = 16
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.Duplicate(ctx, src.ID)
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	items, err := svc.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != workers+1 {
		t.Fatalf("got %d endpoints", len(items))
	}
	seen := map[string]bool{}
	for _, item := range items {
		if seen[item.Path] {
			t.Fatalf("duplicate path: %s", item.Path)
		}
		seen[item.Path] = true
	}
}
