package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	// DefaultDevSessionSecret is only used when APP_ENV is not production.
	// It keeps sessions valid across restarts during development.
	DefaultDevSessionSecret = "medrouter-dev-only-session-secret-change-me"
	Version                 = "0.1.0"
)

type Config struct {
	AppEnv         string
	Port           string
	BasePath       string
	PublicOrigin   string
	SystemName     string
	Version        string
	TrustedProxies []string

	SessionSecret        string
	SessionCookieName    string
	SessionCookieSecure  bool
	SessionTTL           time.Duration
	SessionTouchInterval time.Duration

	RegisterEnabled   bool
	InitialAdminUser  string
	InitialAdminPass  string
	InitialAdminEmail string

	SQLDSN         string
	SQLitePath     string
	DBMaxOpenConns int

	EmailVerificationEnabled bool
	EmailDebugMode           bool
	AllowEmailDebugInProd    bool
	SMTPHost                 string
	SMTPPort                 int
	SMTPUsername             string
	SMTPPassword             string
	SMTPFrom                 string
	SMTPTLSMode              string
	SMTPTimeout              time.Duration

	VerificationCodeTTL     time.Duration
	VerificationMaxAttempts int
	VerificationResend      time.Duration
	VerificationIPHourly    int

	GitHubClientID     string
	GitHubClientSecret string

	GoogleClientID     string
	GoogleClientSecret string

	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCDisplayName  string
	OIDCScopes       string

	OAuthRedirectBase string

	LoginRatePerMinute  int
	RegisterRatePerHour int

	ModelLibraryEnabled  bool
	ModelLibraryAPIURL   string
	ModelLibraryCacheTTL time.Duration
	ModelLibraryTimeout  time.Duration

	DatasetLibraryEnabled  bool
	DatasetLibraryAPIURL   string
	DatasetLibraryCacheTTL time.Duration
	DatasetLibraryTimeout  time.Duration
}

// Load reads configuration from the environment. A .env file in the working
// directory is loaded when present (development convenience).
func Load() (*Config, error) {
	_ = godotenv.Load()

	appEnv := strings.ToLower(getenv("APP_ENV", "development"))

	cfg := &Config{
		AppEnv:         appEnv,
		Port:           getenv("PORT", "3000"),
		BasePath:       normalizeBasePath(os.Getenv("BASE_PATH")),
		PublicOrigin:   strings.TrimRight(getenv("PUBLIC_ORIGIN", "http://localhost:5173"), "/"),
		SystemName:     getenv("SYSTEM_NAME", "MedRouter"),
		Version:        Version,
		TrustedProxies: splitCSV(os.Getenv("TRUSTED_PROXIES")),

		SessionSecret:        os.Getenv("SESSION_SECRET"),
		SessionCookieName:    getenv("SESSION_COOKIE_NAME", "medrouter_session"),
		SessionCookieSecure:  boolEnv("SESSION_COOKIE_SECURE", appEnv == "production"),
		SessionTTL:           daysEnv("SESSION_TTL_DAYS", 30),
		SessionTouchInterval: secondsEnv("SESSION_TOUCH_INTERVAL_SECONDS", 3600),

		RegisterEnabled:   boolEnv("REGISTER_ENABLED", true),
		InitialAdminUser:  os.Getenv("INITIAL_ADMIN_USERNAME"),
		InitialAdminPass:  os.Getenv("INITIAL_ADMIN_PASSWORD"),
		InitialAdminEmail: os.Getenv("INITIAL_ADMIN_EMAIL"),

		SQLDSN:         strings.TrimSpace(os.Getenv("SQL_DSN")),
		SQLitePath:     getenv("SQLITE_PATH", "data/medrouter.db"),
		DBMaxOpenConns: intEnv("DB_MAX_OPEN_CONNS", 10),

		EmailVerificationEnabled: boolEnv("EMAIL_VERIFICATION_ENABLED", true),
		EmailDebugMode:           boolEnv("EMAIL_DEBUG_MODE", appEnv != "production"),
		AllowEmailDebugInProd:    boolEnv("ALLOW_EMAIL_DEBUG_IN_PRODUCTION", false),
		SMTPHost:                 os.Getenv("SMTP_HOST"),
		SMTPPort:                 intEnv("SMTP_PORT", 465),
		SMTPUsername:             os.Getenv("SMTP_USERNAME"),
		SMTPPassword:             os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:                 getenv("SMTP_FROM", "MedRouter <no-reply@medrouter.local>"),
		SMTPTLSMode:              strings.ToLower(getenv("SMTP_TLS_MODE", "implicit")),
		SMTPTimeout:              secondsEnv("SMTP_TIMEOUT_SECONDS", 10),

		VerificationCodeTTL:     secondsEnv("VERIFICATION_CODE_TTL_SECONDS", 600),
		VerificationMaxAttempts: intEnv("VERIFICATION_MAX_ATTEMPTS", 5),
		VerificationResend:      secondsEnv("VERIFICATION_RESEND_COOLDOWN_SECONDS", 60),
		VerificationIPHourly:    intEnv("VERIFICATION_IP_HOURLY_LIMIT", 10),

		GitHubClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),

		OIDCIssuerURL:    strings.TrimRight(os.Getenv("OIDC_ISSUER_URL"), "/"),
		OIDCClientID:     os.Getenv("OIDC_CLIENT_ID"),
		OIDCClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		OIDCDisplayName:  getenv("OIDC_DISPLAY_NAME", "SSO"),
		OIDCScopes:       getenv("OIDC_SCOPES", "openid email profile"),

		OAuthRedirectBase: strings.TrimRight(os.Getenv("OAUTH_REDIRECT_BASE"), "/"),

		LoginRatePerMinute:  intEnv("LOGIN_RATE_LIMIT_PER_MINUTE", 10),
		RegisterRatePerHour: intEnv("REGISTER_RATE_LIMIT_PER_HOUR", 5),

		ModelLibraryEnabled:  boolEnv("MODEL_LIBRARY_ENABLED", true),
		ModelLibraryAPIURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("MODEL_LIBRARY_API_URL")), "/"),
		ModelLibraryCacheTTL: secondsEnv("MODEL_LIBRARY_CACHE_TTL_SECONDS", 30),
		ModelLibraryTimeout:  secondsEnv("MODEL_LIBRARY_REQUEST_TIMEOUT_SECONDS", 5),

		DatasetLibraryEnabled:  boolEnv("DATASET_LIBRARY_ENABLED", true),
		DatasetLibraryAPIURL:   strings.TrimRight(strings.TrimSpace(os.Getenv("DATASET_LIBRARY_API_URL")), "/"),
		DatasetLibraryCacheTTL: secondsEnv("DATASET_LIBRARY_CACHE_TTL_SECONDS", 30),
		DatasetLibraryTimeout:  secondsEnv("DATASET_LIBRARY_REQUEST_TIMEOUT_SECONDS", 5),
	}

	if cfg.SessionSecret == "" {
		if appEnv == "production" {
			return nil, fmt.Errorf("SESSION_SECRET is required in production")
		}
		cfg.SessionSecret = DefaultDevSessionSecret
	}
	if appEnv == "production" && len(cfg.SessionSecret) < 32 {
		return nil, fmt.Errorf("SESSION_SECRET must be at least 32 characters in production")
	}
	if appEnv == "production" && cfg.EmailDebugMode && !cfg.AllowEmailDebugInProd {
		return nil, fmt.Errorf("EMAIL_DEBUG_MODE must be false in production (set ALLOW_EMAIL_DEBUG_IN_PRODUCTION=true to override)")
	}
	if appEnv == "production" && cfg.EmailVerificationEnabled && !cfg.EmailDebugMode && cfg.SMTPHost == "" {
		return nil, fmt.Errorf("SMTP_HOST is required when EMAIL_VERIFICATION_ENABLED=true in production")
	}
	if cfg.OAuthRedirectBase == "" {
		cfg.OAuthRedirectBase = cfg.PublicOrigin + cfg.BasePath
	}

	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.AppEnv == "production" }

// CallbackURL returns the OAuth callback registered with the provider.
func (c *Config) CallbackURL(provider string) string {
	return fmt.Sprintf("%s/api/oauth/%s/callback", c.OAuthRedirectBase, provider)
}

// FrontendPath prefixes an absolute frontend path with BASE_PATH.
func (c *Config) FrontendPath(path string) string {
	if path == "" {
		return c.BasePath + "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.BasePath + path
}

func normalizeBasePath(raw string) string {
	raw = strings.Trim(strings.TrimSpace(raw), "/")
	if raw == "" {
		return ""
	}
	return "/" + raw
}

func splitCSV(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func boolEnv(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return v
}

func intEnv(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func secondsEnv(key string, fallback int) time.Duration {
	return time.Duration(intEnv(key, fallback)) * time.Second
}

func daysEnv(key string, fallback int) time.Duration {
	return time.Duration(intEnv(key, fallback)) * 24 * time.Hour
}
