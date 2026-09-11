package config_test

import (
	"testing"
	"time"

	"mockapi/internal/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" || cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")
	cfg, err = config.Load()
	if err != nil || cfg.HTTPAddr != ":9090" || cfg.ShutdownTimeout != 3*time.Second {
		t.Fatalf("environment overrides failed: %+v, %v", cfg, err)
	}
	for _, value := range []string{"invalid", "0s", "-1s"} {
		t.Setenv("SHUTDOWN_TIMEOUT", value)
		if _, err := config.Load(); err == nil {
			t.Fatalf("accepted invalid timeout %q", value)
		}
	}
}

func TestRequestTimeoutValidation(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	for _, value := range []string{"bad", "0s", "-1s", "15s", "30s"} {
		t.Setenv("REQUEST_TIMEOUT", value)
		if _, err := config.Load(); err == nil {
			t.Fatalf("accepted %s", value)
		}
	}
	t.Setenv("REQUEST_TIMEOUT", "2s")
	cfg, err := config.Load()
	if err != nil || cfg.RequestTimeout != 2*time.Second {
		t.Fatalf("unexpected config: %+v %v", cfg, err)
	}
}

func TestPoolAndHTTPTimeoutConfig(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("DB_MAX_CONNS", "50")
	t.Setenv("DB_MIN_CONNS", "10")
	t.Setenv("DB_MAX_CONN_LIFETIME", "2h")
	t.Setenv("DB_MAX_CONN_IDLE_TIME", "15m")
	t.Setenv("HTTP_WRITE_TIMEOUT", "20s")
	t.Setenv("HTTP_READ_TIMEOUT", "10s")
	t.Setenv("REQUEST_TIMEOUT", "5s")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBMaxConns != 50 || cfg.DBMinConns != 10 {
		t.Fatalf("unexpected DB conns: max=%d min=%d", cfg.DBMaxConns, cfg.DBMinConns)
	}
	if cfg.DBMaxConnLifetime != 2*time.Hour || cfg.DBMaxConnIdleTime != 15*time.Minute {
		t.Fatalf("unexpected DB durations: lifetime=%v idle=%v", cfg.DBMaxConnLifetime, cfg.DBMaxConnIdleTime)
	}
	if cfg.HTTPWriteTimeout != 20*time.Second || cfg.HTTPReadTimeout != 10*time.Second {
		t.Fatalf("unexpected HTTP timeouts: write=%v read=%v", cfg.HTTPWriteTimeout, cfg.HTTPReadTimeout)
	}
}

