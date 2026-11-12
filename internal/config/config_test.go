package config

import (
	"strings"
	"testing"
)

var envKeys = []string{
	"APP_ENV", "PORT", "BASE_PATH", "PUBLIC_ORIGIN", "SYSTEM_NAME", "TRUSTED_PROXIES",
	"SESSION_SECRET", "SESSION_COOKIE_NAME", "SESSION_COOKIE_SECURE",
	"SESSION_TTL_DAYS", "SESSION_TOUCH_INTERVAL_SECONDS",
	"REGISTER_ENABLED", "INITIAL_ADMIN_USERNAME", "INITIAL_ADMIN_PASSWORD", "INITIAL_ADMIN_EMAIL",
	"SQL_DSN", "SQLITE_PATH", "DB_MAX_OPEN_CONNS",
	"EMAIL_VERIFICATION_ENABLED", "EMAIL_DEBUG_MODE", "ALLOW_EMAIL_DEBUG_IN_PRODUCTION",
	"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM", "SMTP_TLS_MODE",
	"SMTP_TIMEOUT_SECONDS",
	"VERIFICATION_CODE_TTL_SECONDS", "VERIFICATION_MAX_ATTEMPTS",
	"VERIFICATION_RESEND_COOLDOWN_SECONDS", "VERIFICATION_IP_HOURLY_LIMIT",
	"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
	"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
	"OIDC_ISSUER_URL", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET", "OIDC_DISPLAY_NAME", "OIDC_SCOPES",
	"OAUTH_REDIRECT_BASE",
	"LOGIN_RATE_LIMIT_PER_MINUTE", "REGISTER_RATE_LIMIT_PER_HOUR",
	"MODEL_LIBRARY_ENABLED", "MODEL_LIBRARY_API_URL",
	"MODEL_LIBRARY_CACHE_TTL_SECONDS", "MODEL_LIBRARY_REQUEST_TIMEOUT_SECONDS",
	"DATASET_LIBRARY_ENABLED", "DATASET_LIBRARY_API_URL",
	"DATASET_LIBRARY_CACHE_TTL_SECONDS", "DATASET_LIBRARY_REQUEST_TIMEOUT_SECONDS",
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range envKeys {
		t.Setenv(key, "")
	}
}

func TestLoadDevelopmentDefaults(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AppEnv != "development" || cfg.SessionSecret != DefaultDevSessionSecret {
		t.Fatalf("unexpected development config: %+v", cfg)
	}
	if !cfg.EmailVerificationEnabled || !cfg.EmailDebugMode {
		t.Fatalf("development defaults should enable verification and debug mail")
	}
	if cfg.PublicOrigin != "http://localhost:5173" {
		t.Fatalf("public origin = %q", cfg.PublicOrigin)
	}
	if cfg.OAuthRedirectBase != cfg.PublicOrigin {
		t.Fatalf("redirect base = %q", cfg.OAuthRedirectBase)
	}
	if got := cfg.CallbackURL("github"); got != "http://localhost:5173/api/oauth/github/callback" {
		t.Fatalf("callback = %q", got)
	}
	if got := cfg.FrontendPath("/dashboard"); got != "/dashboard" {
		t.Fatalf("frontend path = %q", got)
	}
}

func TestDatasetLibraryConfig(t *testing.T) {
	clearEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.DatasetLibraryEnabled || cfg.DatasetLibraryAPIURL != "" {
		t.Fatalf("dataset defaults = %+v", cfg)
	}
	if cfg.DatasetLibraryCacheTTL.Seconds() != 30 || cfg.DatasetLibraryTimeout.Seconds() != 5 {
		t.Fatalf("dataset durations = %s/%s", cfg.DatasetLibraryCacheTTL, cfg.DatasetLibraryTimeout)
	}

	t.Setenv("DATASET_LIBRARY_ENABLED", "false")
	t.Setenv("DATASET_LIBRARY_API_URL", "http://127.0.0.1:8001/")
	t.Setenv("DATASET_LIBRARY_CACHE_TTL_SECONDS", "45")
	t.Setenv("DATASET_LIBRARY_REQUEST_TIMEOUT_SECONDS", "8")

	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DatasetLibraryEnabled {
		t.Fatal("DATASET_LIBRARY_ENABLED=false was ignored")
	}
	if cfg.DatasetLibraryAPIURL != "http://127.0.0.1:8001" {
		t.Fatalf("dataset API URL = %q", cfg.DatasetLibraryAPIURL)
	}
	if cfg.DatasetLibraryCacheTTL.Seconds() != 45 || cfg.DatasetLibraryTimeout.Seconds() != 8 {
		t.Fatalf("dataset durations = %s/%s", cfg.DatasetLibraryCacheTTL, cfg.DatasetLibraryTimeout)
	}
}

func TestLoadBasePath(t *testing.T) {
	clearEnv(t)
	t.Setenv("BASE_PATH", "/medrouter/")
	t.Setenv("PUBLIC_ORIGIN", "https://example.com/")
	t.Setenv("OAUTH_REDIRECT_BASE", "https://auth.example.com/prefix/")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BasePath != "/medrouter" {
		t.Fatalf("base path = %q", cfg.BasePath)
	}
	if cfg.PublicOrigin != "https://example.com" {
		t.Fatalf("public origin = %q", cfg.PublicOrigin)
	}
	if got := cfg.FrontendPath("sign-in"); got != "/medrouter/sign-in" {
		t.Fatalf("frontend path = %q", got)
	}
	if got := cfg.FrontendPath(""); got != "/medrouter/" {
		t.Fatalf("frontend root = %q", got)
	}
	if got := cfg.CallbackURL("google"); got != "https://auth.example.com/prefix/api/oauth/google/callback" {
		t.Fatalf("callback = %q", got)
	}
}

func TestProductionValidation(t *testing.T) {
	const secret = "0123456789abcdef0123456789abcdef"

	cases := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "missing secret",
			env:     map[string]string{"APP_ENV": "production"},
			wantErr: "SESSION_SECRET is required",
		},
		{
			name:    "short secret",
			env:     map[string]string{"APP_ENV": "production", "SESSION_SECRET": "short"},
			wantErr: "at least 32 characters",
		},
		{
			name: "debug email in production",
			env: map[string]string{
				"APP_ENV": "production", "SESSION_SECRET": secret,
				"EMAIL_DEBUG_MODE": "true",
			},
			wantErr: "EMAIL_DEBUG_MODE must be false",
		},
		{
			name: "missing smtp host",
			env: map[string]string{
				"APP_ENV": "production", "SESSION_SECRET": secret,
				"EMAIL_DEBUG_MODE": "false", "EMAIL_VERIFICATION_ENABLED": "true",
			},
			wantErr: "SMTP_HOST is required",
		},
		{
			name: "valid with smtp",
			env: map[string]string{
				"APP_ENV": "production", "SESSION_SECRET": secret,
				"EMAIL_DEBUG_MODE": "false", "EMAIL_VERIFICATION_ENABLED": "true",
				"SMTP_HOST": "smtp.example.com",
			},
		},
		{
			name: "debug explicitly allowed",
			env: map[string]string{
				"APP_ENV": "production", "SESSION_SECRET": secret,
				"EMAIL_DEBUG_MODE": "true", "ALLOW_EMAIL_DEBUG_IN_PRODUCTION": "true",
			},
		},
		{
			name: "verification disabled needs no smtp",
			env: map[string]string{
				"APP_ENV": "production", "SESSION_SECRET": secret,
				"EMAIL_DEBUG_MODE": "false", "EMAIL_VERIFICATION_ENABLED": "false",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clearEnv(t)
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			cfg, err := Load()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if !cfg.IsProduction() {
					t.Fatalf("IsProduction() = false")
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Load error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}
