package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestNewValidation(t *testing.T) {
	ctx := context.Background()

	if _, err := New(ctx, ""); err == nil {
		t.Fatal("expected error for empty database URL")
	}

	if _, err := New(ctx, "://invalid-url"); err == nil {
		t.Fatal("expected error for malformed database URL")
	}

	_, err := New(ctx, "postgres://user:pass@127.0.0.1:54321/db?sslmode=disable")
	if err == nil {
		t.Fatal("expected error for unreachable database")
	}
	if !strings.Contains(err.Error(), "database startup ping failed") {
		t.Fatalf("expected ping failure diagnostic, got: %v", err)
	}
}

func TestOptions(t *testing.T) {
	cfg, err := pgxpool.ParseConfig("postgres://localhost:5432/app")
	if err != nil {
		t.Fatal(err)
	}

	opts := []Option{
		WithMaxConns(50),
		WithMinConns(10),
		WithMaxConnLifetime(2 * time.Hour),
		WithMaxConnIdleTime(15 * time.Minute),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.MaxConns != 50 {
		t.Fatalf("expected MaxConns=50, got %d", cfg.MaxConns)
	}
	if cfg.MinConns != 10 {
		t.Fatalf("expected MinConns=10, got %d", cfg.MinConns)
	}
	if cfg.MaxConnLifetime != 2*time.Hour {
		t.Fatalf("expected MaxConnLifetime=2h, got %v", cfg.MaxConnLifetime)
	}
	if cfg.MaxConnIdleTime != 15*time.Minute {
		t.Fatalf("expected MaxConnIdleTime=15m, got %v", cfg.MaxConnIdleTime)
	}
}

func TestStatsNil(t *testing.T) {
	s := Stats(nil)
	if s.AcquiredConns != 0 || s.IdleConns != 0 || s.TotalConns != 0 || s.MaxConns != 0 {
		t.Fatalf("expected zero stats for nil pool, got: %+v", s)
	}
}
