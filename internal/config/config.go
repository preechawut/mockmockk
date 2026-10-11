package config

import (
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TrustedProxies     []string
	HTTPAddr           string
	DatabaseURL        string
	CORSAllowedOrigins []string
	APIRateLimit       int
	LogFormat          string
	LogLevel           string
}

func Load() (Config, error) {
	cfg := Default()

	// trusted proxies
	for _, value := range strings.Split(os.Getenv("TRUSTED_PROXIES"), ",") {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		_, ipErr := netip.ParseAddr(value)
		_, cidrErr := netip.ParsePrefix(value)
		if ipErr != nil && cidrErr != nil {
			return Config{}, fmt.Errorf("TRUSTED_PROXIES must contain IP addresses or CIDRs")
		}

		cfg.TrustedProxies = append(cfg.TrustedProxies, value)
	}

	cfg.HTTPAddr = env("HTTP_ADDR", cfg.HTTPAddr)
	cfg.DatabaseURL = env("DATABASE_URL", cfg.DatabaseURL)
	cfg.LogFormat = strings.ToLower(env("LOG_FORMAT", cfg.LogFormat))
	cfg.LogLevel = strings.ToUpper(env("LOG_LEVEL", cfg.LogLevel))

	if cfg.LogFormat != "json" && cfg.LogFormat != "text" {
		return Config{}, fmt.Errorf("LOG_FORMAT must be json or text")
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return Config{}, fmt.Errorf("invalid LOG_LEVEL: %w", err)
	}

	rateLimit, err := positiveInt("API_RATE_LIMIT", cfg.APIRateLimit)
	if err != nil {
		return Config{}, err
	}
	cfg.APIRateLimit = rateLimit

	// cors origins
	if origins, ok := os.LookupEnv("CORS_ALLOWED_ORIGINS"); ok {
		cfg.CORSAllowedOrigins = nil
		for _, origin := range strings.Split(origins, ",") {
			origin = strings.TrimSpace(origin)
			if origin == "" {
				continue
			}

			if !isValidOrigin(origin) {
				return Config{}, fmt.Errorf("CORS_ALLOWED_ORIGINS must contain exact http(s) origins without paths or wildcards")
			}

			cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, origin)
		}
	}

	return cfg, nil
}

func isValidOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}

	return parsed.Host != "" &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.User == nil &&
		parsed.Path == "" &&
		parsed.RawQuery == "" && !parsed.ForceQuery &&
		parsed.Fragment == ""
}

func env(name, fallback string) string {
	value := os.Getenv(name)
	if value != "" {
		return value
	}

	return fallback
}

func Default() Config {
	return Config{
		HTTPAddr:    ":8080",
		DatabaseURL: "postgres://app:localpassword@localhost:5433/app?sslmode=disable",
		CORSAllowedOrigins: []string{
			"http://localhost:8080",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3001",
		},
		APIRateLimit: 300,
		LogFormat:    "json",
		LogLevel:     "INFO",
	}
}

func positiveInt(name string, fallback int) (int, error) {
	valStr := os.Getenv(name)
	if valStr == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(valStr)
	if err != nil || value <= 0 || value > 1000000 {
		return 0, fmt.Errorf("%s must be an integer between 1 and 1000000", name)
	}

	return value, nil
}
