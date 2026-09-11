package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// config
type Config struct {
	ServiceName        string
	RequestTimeout     time.Duration
	HTTPWriteTimeout   time.Duration
	HTTPReadTimeout    time.Duration
	LogLevel           string
	LogFormat          string
	DatabaseURL        string
	DBMaxConns         int32
	DBMinConns         int32
	DBMaxConnLifetime  time.Duration
	DBMaxConnIdleTime  time.Duration
	Environment        string
	HTTPAddr           string
	ShutdownTimeout    time.Duration
	CORSAllowedOrigins string
}

func Load() (Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("DB_URL")
	}
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	timeout, err := time.ParseDuration(env("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil || timeout <= 0 {
		return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT must be a positive duration, e.g. 10s")
	}

	writeTimeout, err := time.ParseDuration(env("HTTP_WRITE_TIMEOUT", "15s"))
	if err != nil || writeTimeout <= 0 {
		return Config{}, fmt.Errorf("HTTP_WRITE_TIMEOUT must be a positive duration, e.g. 15s")
	}

	readTimeout, err := time.ParseDuration(env("HTTP_READ_TIMEOUT", "15s"))
	if err != nil || readTimeout <= 0 {
		return Config{}, fmt.Errorf("HTTP_READ_TIMEOUT must be a positive duration, e.g. 15s")
	}

	requestTimeout, err := time.ParseDuration(env("REQUEST_TIMEOUT", "10s"))
	if err != nil || requestTimeout <= 0 || requestTimeout >= writeTimeout {
		return Config{}, fmt.Errorf("REQUEST_TIMEOUT must be positive and less than HTTP write timeout (%v)", writeTimeout)
	}

	maxConns, err := parsePositiveInt32(env("DB_MAX_CONNS", "25"))
	if err != nil {
		return Config{}, fmt.Errorf("DB_MAX_CONNS must be a positive integer: %w", err)
	}

	minConns, err := parseNonNegativeInt32(env("DB_MIN_CONNS", "5"))
	if err != nil || minConns > maxConns {
		return Config{}, fmt.Errorf("DB_MIN_CONNS must be non-negative and <= DB_MAX_CONNS")
	}

	maxConnLifetime, err := time.ParseDuration(env("DB_MAX_CONN_LIFETIME", "1h"))
	if err != nil || maxConnLifetime <= 0 {
		return Config{}, fmt.Errorf("DB_MAX_CONN_LIFETIME must be a positive duration, e.g. 1h")
	}

	maxConnIdleTime, err := time.ParseDuration(env("DB_MAX_CONN_IDLE_TIME", "30m"))
	if err != nil || maxConnIdleTime <= 0 {
		return Config{}, fmt.Errorf("DB_MAX_CONN_IDLE_TIME must be a positive duration, e.g. 30m")
	}

	httpAddr := env("HTTP_ADDR", "")
	if httpAddr == "" {
		if port := os.Getenv("PORT"); port != "" {
			httpAddr = ":" + port
		} else {
			httpAddr = ":8080"
		}
	}

	return Config{
		ServiceName:        env("SERVICE_NAME", "mockmock-service"),
		RequestTimeout:     requestTimeout,
		HTTPWriteTimeout:   writeTimeout,
		HTTPReadTimeout:    readTimeout,
		LogLevel:           env("LOG_LEVEL", "info"),
		LogFormat:          env("LOG_FORMAT", "json"),
		DatabaseURL:        dbURL,
		DBMaxConns:         maxConns,
		DBMinConns:         minConns,
		DBMaxConnLifetime:  maxConnLifetime,
		DBMaxConnIdleTime:  maxConnIdleTime,
		Environment:        env("APP_ENV", "development"),
		HTTPAddr:           httpAddr,
		ShutdownTimeout:    timeout,
		CORSAllowedOrigins: env("CORS_ALLOWED_ORIGINS", "*"),
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func parsePositiveInt32(val string) (int32, error) {
	n, err := strconv.ParseInt(val, 10, 32)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("expected positive int32, got %q", val)
	}
	return int32(n), nil
}

func parseNonNegativeInt32(val string) (int32, error) {
	n, err := strconv.ParseInt(val, 10, 32)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("expected non-negative int32, got %q", val)
	}
	return int32(n), nil
}
