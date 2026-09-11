package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mockapi/internal/modules/endpoint/domain"
	"mockapi/internal/modules/endpoint/repository/postgres"
	"mockapi/internal/modules/endpoint/service"
)

func TestPostgresRepositoryIntegration(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	schema := fmt.Sprintf("endpoint_test_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := admin.Exec(cleanup, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Errorf("cleanup: %v", err)
		}
	}()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = quoted + ", public"
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for _, file := range []string{"000001_create_mockmocks.up.sql", "000002_create_projects_resources.up.sql", "000003_project_endpoints.up.sql"} {
		migration, err := os.ReadFile("../../../../../migrations/" + file)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, string(migration)); err != nil {
			t.Fatal(err)
		}
	}
	repo := postgres.NewRepository(pool)
	svc := service.NewService(repo)
	original, err := svc.Create(ctx, "GET", "/users", 201, []byte(`{"ok":true}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(ctx, "GET", "/users", 200, nil); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict: %v", err)
	}
	matched, err := svc.Match(ctx, "get", "users/")
	if err != nil || matched.ID != original.ID || matched.Status != 201 {
		t.Fatalf("match: %+v %v", matched, err)
	}
	if _, err := svc.Update(ctx, original.ID, "", "/people", 202, nil); err != nil {
		t.Fatal(err)
	}
	const count = 8
	var wg sync.WaitGroup
	failures := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, err := svc.Duplicate(ctx, original.ID); failures <- err }()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	items, err := svc.List(ctx)
	if err != nil || len(items) != count+1 {
		t.Fatalf("list count %d: %v", len(items), err)
	}
	if err := svc.Delete(ctx, original.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GetByID(ctx, original.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleted item: %v", err)
	}
	if _, err := svc.Match(ctx, "GET", "/people"); !errors.Is(err, domain.ErrNoMatch) {
		t.Fatalf("deleted route: %v", err)
	}
	pool.Close()
	if _, err := svc.Match(ctx, "GET", "/anything"); err == nil || errors.Is(err, domain.ErrNoMatch) {
		t.Fatalf("closed pool became missing route: %v", err)
	}
}
