package config

import "testing"

func TestCORSOriginValidation(t *testing.T) {
	for _, origin := range []string{
		"https://user:pass@example.com", "https://example.com/path",
		"https://example.com?", "https://example.com?q=1",
		"https://example.com#fragment", "ftp://example.com", "https://",
	} {
		t.Run(origin, func(t *testing.T) {
			t.Setenv("CORS_ALLOWED_ORIGINS", origin)
			if _, err := Load(); err == nil {
				t.Fatalf("accepted invalid origin %q", origin)
			}
		})
	}
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,http://localhost:3000")
	cfg, err := Load()
	if err != nil || len(cfg.CORSAllowedOrigins) != 2 {
		t.Fatalf("valid origins: %+v %v", cfg, err)
	}
	t.Setenv("CORS_ALLOWED_ORIGINS", "")
	cfg, err = Load()
	if err != nil || len(cfg.CORSAllowedOrigins) != 0 {
		t.Fatal("empty origin list must remain empty")
	}
}

func TestTrustedProxiesValidation(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "127.0.0.1,::1,10.0.0.0/24")
	cfg, err := Load()
	if err != nil || len(cfg.TrustedProxies) != 3 {
		t.Fatalf("valid proxies: %+v %v", cfg.TrustedProxies, err)
	}
	t.Setenv("TRUSTED_PROXIES", "*")
	if _, err := Load(); err == nil {
		t.Fatal("accepted invalid proxy")
	}
}
