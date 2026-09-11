package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Option func(*pgxpool.Config)

func WithMaxConns(n int32) Option {
	return func(c *pgxpool.Config) {
		if n > 0 {
			c.MaxConns = n
		}
	}
}

func WithMinConns(n int32) Option {
	return func(c *pgxpool.Config) {
		if n >= 0 {
			c.MinConns = n
		}
	}
}

func WithMaxConnLifetime(d time.Duration) Option {
	return func(c *pgxpool.Config) {
		if d > 0 {
			c.MaxConnLifetime = d
		}
	}
}

func WithMaxConnIdleTime(d time.Duration) Option {
	return func(c *pgxpool.Config) {
		if d > 0 {
			c.MaxConnIdleTime = d
		}
	}
}

func New(ctx context.Context, databaseURL string, opts ...Option) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, errors.New("database URL is required")
	}
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}
	for _, opt := range opts {
		opt(cfg)
	}
	startupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	pool, err := pgxpool.NewWithConfig(startupCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := pool.Ping(startupCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database startup ping failed: %w", err)
	}
	return pool, nil
}

type PoolStats struct {
	AcquiredConns int32 `json:"acquired_conns"`
	IdleConns     int32 `json:"idle_conns"`
	TotalConns    int32 `json:"total_conns"`
	MaxConns      int32 `json:"max_conns"`
}

func Stats(pool *pgxpool.Pool) PoolStats {
	if pool == nil {
		return PoolStats{}
	}
	s := pool.Stat()
	return PoolStats{
		AcquiredConns: s.AcquiredConns(),
		IdleConns:     s.IdleConns(),
		TotalConns:    s.TotalConns(),
		MaxConns:      s.MaxConns(),
	}
}
