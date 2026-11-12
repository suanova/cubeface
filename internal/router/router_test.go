package router_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/suanova/medrouter/internal/common"
	"github.com/suanova/medrouter/internal/config"
	"github.com/suanova/medrouter/internal/controller"
	"github.com/suanova/medrouter/internal/model"
	"github.com/suanova/medrouter/internal/router"
	"github.com/suanova/medrouter/internal/service"
)

type envelope struct {
	Success bool            `json:"success"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type userDTO struct {
	ID            int64      `json:"id"`
	Username      string     `json:"username"`
	DisplayName   string     `json:"display_name"`
	Email         string     `json:"email"`
	Role          int        `json:"role"`
	Status        int        `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     time.Time  `json:"created_at"`
	LastLoginAt   *time.Time `json:"last_login_at"`
}

type testApp struct {
	t      *testing.T
	server *httptest.Server
	client *http.Client
	cfg    *config.Config
}

func baseConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		AppEnv:               "test",
		SystemName:           "MedRouter",
		Version:              "test",
		PublicOrigin:         "http://example.test",
		SessionSecret:        "test-session-secret-with-enough-entropy",
		SessionCookieName:    "medrouter_session",
		SessionTTL:           24 * time.Hour,
		SessionTouchInterval: time.Hour,

		RegisterEnabled: true,
		SQLitePath:      filepath.Join(t.TempDir(), "medrouter.db"),
		DBMaxOpenConns:  1,

		EmailVerificationEnabled: false,
		EmailDebugMode:           true,
		VerificationCodeTTL:      10 * time.Minute,
		VerificationMaxAttempts:  5,
		VerificationResend:       time.Minute,
		VerificationIPHourly:     100,

		LoginRatePerMinute:  1000,
		RegisterRatePerHour: 1000,
	}
}

func newTestApp(t *testing.T, mutate func(*config.Config)) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	cfg := baseConfig(t)
	if mutate != nil {
		mutate(cfg)
	}
	if err := model.InitDB(cfg); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		if err := model.Close(); err != nil {
			t.Errorf("close db: %v", err)
		}
		model.DB = nil
	})

	sessions := service.NewSessionService(cfg)
	email := service.NewEmailService(cfg)
	verification := service.NewVerificationService(cfg, email)
	users := service.NewUserService(cfg, sessions, verification)
	oauth := service.NewOAuthService(cfg, users, sessions)

	engine := router.New(router.Deps{
		Config:       cfg,
		Users:        controller.NewUserController(cfg, users, sessions),
		Verification: controller.NewVerificationController(verification),
		OAuth:        controller.NewOAuthController(cfg, oauth, sessions),
		Status:       controller.NewStatusController(cfg, oauth),
		Sessions:     sessions,
	})
	server := httptest.NewServer(engine)
	t.Cleanup(server.Close)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	return &testApp{
		t:      t,
		server: server,
		cfg:    cfg,
		client: &http.Client{
			Jar: jar,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (a *testApp) apiPath(path string) string {
	if a.cfg.BasePath != "" && strings.HasPrefix(path, "/api/") {
		return a.cfg.BasePath + path
	}
	return path
}

func (a *testApp) request(method, path string, body any, mutate ...func(*http.Request)) (*http.Response, envelope) {
	a.t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			a.t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, a.server.URL+a.apiPath(path), reader)
	if err != nil {
		a.t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if method != http.MethodGet && method != http.MethodHead {
		req.Header.Set("Origin", a.cfg.PublicOrigin)
	}
	for _, fn := range mutate {
		fn(req)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		a.t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		a.t.Fatalf("read response: %v", err)
	}
	var env envelope
	if len(raw) > 0 && strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(raw, &env); err != nil {
			a.t.Fatalf("decode envelope: %v (%s)", err, raw)
		}
	}
	return resp, env
}

func (a *testApp) register(username, email, password, code string) (*http.Response, envelope) {
	a.t.Helper()
	return a.request(http.MethodPost, "/api/user/register", map[string]any{
		"username":          username,
		"email":             email,
		"password":          password,
		"verification_code": code,
	})
}

func (a *testApp) login(identifier, password string) (*http.Response, envelope) {
	a.t.Helper()
	return a.request(http.MethodPost, "/api/user/login", map[string]any{
		"username": identifier,
		"password": password,
	})
}

func (a *testApp) self() (*http.Response, envelope) {
	a.t.Helper()
	return a.request(http.MethodGet, "/api/user/self", nil)
}

func (a *testApp) mustSuccess(env envelope, label string) {
	a.t.Helper()
	if !env.Success {
		a.t.Fatalf("%s: expected success, got code=%q message=%q", label, env.Code, env.Message)
	}
}

func (a *testApp) mustCode(resp *http.Response, env envelope, status int, code, label string) {
	a.t.Helper()
	if resp.StatusCode != status || env.Success || env.Code != code {
		a.t.Fatalf("%s: got status=%d success=%v code=%q message=%q; want status=%d code=%q",
			label, resp.StatusCode, env.Success, env.Code, env.Message, status, code)
	}
}

func decodeUser(t *testing.T, env envelope) userDTO {
	t.Helper()
	var payload struct {
		User userDTO `json:"user"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("decode user: %v (%s)", err, env.Data)
	}
	return payload.User
}

func insertVerificationCode(t *testing.T, cfg *config.Config, email, code string, mutate func(*model.VerificationCode)) *model.VerificationCode {
	t.Helper()
	now := time.Now().UTC()
	record := &model.VerificationCode{
		Email:     common.NormalizeEmail(email),
		Purpose:   model.PurposeRegister,
		CodeHash:  common.VerificationCodeHash(cfg.SessionSecret, common.NormalizeEmail(email), model.PurposeRegister, code),
		IP:        "192.0.2.10",
		ExpiresAt: now.Add(cfg.VerificationCodeTTL),
		CreatedAt: now,
	}
	if mutate != nil {
		mutate(record)
	}
	if err := model.DB.Create(record).Error; err != nil {
		t.Fatalf("insert verification code: %v", err)
	}
	return record
}

func TestRegisterLoginSelfLogout(t *testing.T) {
	app := newTestApp(t, nil)

	resp, env := app.register("Alice", "Alice@Example.com", "Passw0rd", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register status: %d", resp.StatusCode)
	}
	app.mustSuccess(env, "register")
	user := decodeUser(t, env)
	if user.Username != "alice" || user.DisplayName != "Alice" {
		t.Fatalf("unexpected username/display name: %+v", user)
	}
	if user.Email != "alice@example.com" || user.EmailVerified {
		t.Fatalf("unexpected email state: %+v", user)
	}

	resp, env = app.login("alice", "Passw0rd")
	app.mustSuccess(env, "login by username")
	if decodeUser(t, env).ID != user.ID {
		t.Fatalf("login returned a different user")
	}

	_, env = app.self()
	app.mustSuccess(env, "self after login")
	if decodeUser(t, env).Username != "alice" {
		t.Fatalf("self returned wrong user")
	}

	_, env = app.request(http.MethodPost, "/api/user/auth/logout", nil)
	app.mustSuccess(env, "logout")

	resp, env = app.self()
	app.mustCode(resp, env, http.StatusUnauthorized, "UNAUTHENTICATED", "self after logout")

	// Email may be used in place of the username.
	_, env = app.login("ALICE@EXAMPLE.COM", "Passw0rd")
	app.mustSuccess(env, "login by email")

	// Logging out twice is idempotent.
	_, env = app.request(http.MethodPost, "/api/user/auth/logout", nil)
	app.mustSuccess(env, "logout twice")
}

func TestRegistrationValidation(t *testing.T) {
	app := newTestApp(t, nil)
	app.mustSuccess(app.registerOrFail("Bob", "bob@example.com", "Passw0rd"), "register bob")

	resp, env := app.register("Bob", "other@example.com", "Passw0rd", "")
	app.mustCode(resp, env, http.StatusOK, "USERNAME_TAKEN", "duplicate username")

	resp, env = app.register("bobby", "BOB@example.com", "Passw0rd", "")
	app.mustCode(resp, env, http.StatusOK, "EMAIL_TAKEN", "duplicate email")

	resp, env = app.register("weakling", "weak@example.com", "weak", "")
	app.mustCode(resp, env, http.StatusOK, "INVALID_PASSWORD", "weak password")

	resp, env = app.register("bad name!", "bad@example.com", "Passw0rd", "")
	app.mustCode(resp, env, http.StatusOK, "INVALID_USERNAME", "invalid username")

	resp, env = app.register("badmail", "not-an-email", "Passw0rd", "")
	app.mustCode(resp, env, http.StatusOK, "INVALID_EMAIL", "malformed email")

	// Registration disabled.
	disabled := newTestApp(t, func(c *config.Config) { c.RegisterEnabled = false })
	resp, env = disabled.register("Carol", "carol@example.com", "Passw0rd", "")
	app.mustCode(resp, env, http.StatusOK, "REGISTER_DISABLED", "registration disabled")
}

func (a *testApp) registerOrFail(username, email, password string) envelope {
	a.t.Helper()
	_, env := a.register(username, email, password, "")
	a.mustSuccess(env, "register "+username)
	return env
}

func TestEmailVerificationRegistration(t *testing.T) {
	app := newTestApp(t, func(c *config.Config) { c.EmailVerificationEnabled = true })
	const email = "carol@example.com"
	record := insertVerificationCode(t, app.cfg, email, "123456", nil)

	// Wrong code is rejected and the attempt is persisted.
	resp, env := app.register("Carol", email, "Passw0rd", "000000")
	app.mustCode(resp, env, http.StatusOK, "VERIFICATION_INVALID", "wrong code")
	var stored model.VerificationCode
	if err := model.DB.First(&stored, record.ID).Error; err != nil {
		t.Fatalf("reload code: %v", err)
	}
	if stored.Attempts != 1 {
		t.Fatalf("attempts = %d, want 1", stored.Attempts)
	}

	// Correct code registers the user and consumes the code.
	_, env = app.register("Carol", email, "Passw0rd", "123456")
	app.mustSuccess(env, "register with code")
	user := decodeUser(t, env)
	if !user.EmailVerified {
		t.Fatalf("registered user should have a verified email")
	}
	if err := model.DB.First(&stored, record.ID).Error; err != nil {
		t.Fatalf("reload consumed code: %v", err)
	}
	if stored.ConsumedAt == nil {
		t.Fatalf("verification code was not consumed")
	}

	// Missing code.
	resp, env = app.register("Dave", "dave@example.com", "Passw0rd", "")
	app.mustCode(resp, env, http.StatusOK, "VERIFICATION_REQUIRED", "missing code")

	// Expired code.
	insertVerificationCode(t, app.cfg, "dave@example.com", "111111", func(r *model.VerificationCode) {
		r.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	})
	resp, env = app.register("Dave", "dave@example.com", "Passw0rd", "111111")
	app.mustCode(resp, env, http.StatusOK, "VERIFICATION_EXPIRED", "expired code")

	// Too many attempts.
	insertVerificationCode(t, app.cfg, "erin@example.com", "222222", func(r *model.VerificationCode) {
		r.Attempts = app.cfg.VerificationMaxAttempts
	})
	resp, env = app.register("Erin", "erin@example.com", "Passw0rd", "222222")
	app.mustCode(resp, env, http.StatusOK, "VERIFICATION_TOO_MANY_ATTEMPTS", "attempt limit")
}

func TestVerificationSendCooldown(t *testing.T) {
	app := newTestApp(t, func(c *config.Config) {
		c.EmailVerificationEnabled = true
		c.VerificationIPHourly = 100
	})

	_, env := app.request(http.MethodPost, "/api/verification/send", map[string]any{"email": "first@example.com"})
	app.mustSuccess(env, "send first code")

	var count int64
	if err := model.DB.Model(&model.VerificationCode{}).Where("email = ?", "first@example.com").Count(&count).Error; err != nil {
		t.Fatalf("count codes: %v", err)
	}
	if count != 1 {
		t.Fatalf("stored codes = %d, want 1", count)
	}

	// The same email hits the resend cooldown.
	resp, env := app.request(http.MethodPost, "/api/verification/send", map[string]any{"email": "first@example.com"})
	app.mustCode(resp, env, http.StatusTooManyRequests, "RATE_LIMITED", "resend cooldown")

	// Invalid email never reaches the mailer.
	resp, env = app.request(http.MethodPost, "/api/verification/send", map[string]any{"email": "not-an-email"})
	app.mustCode(resp, env, http.StatusOK, "INVALID_EMAIL", "invalid email")

	// Registered emails cannot request a registration code.
	now := time.Now().UTC()
	hash, err := common.HashPassword("Passw0rd")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := model.DB.Create(&model.User{
		Username: "taken", DisplayName: "Taken", PasswordHash: hash, Email: "taken@example.com",
		Role: model.RoleCommon, Status: model.StatusEnabled, AuthVersion: 1,
		CreatedAt: now, UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	resp, env = app.request(http.MethodPost, "/api/verification/send", map[string]any{"email": "taken@example.com"})
	app.mustCode(resp, env, http.StatusOK, "EMAIL_TAKEN", "send to registered email")
}

func TestVerificationSendIPQuota(t *testing.T) {
	app := newTestApp(t, func(c *config.Config) {
		c.EmailVerificationEnabled = true
		c.VerificationIPHourly = 1
	})

	_, env := app.request(http.MethodPost, "/api/verification/send", map[string]any{"email": "first@example.com"})
	app.mustSuccess(env, "send first code")

	// The per-IP hourly quota rejects a second recipient.
	resp, env := app.request(http.MethodPost, "/api/verification/send", map[string]any{"email": "second@example.com"})
	app.mustCode(resp, env, http.StatusTooManyRequests, "RATE_LIMITED", "ip quota")
}

func TestSessionInvalidation(t *testing.T) {
	app := newTestApp(t, nil)
	app.mustSuccess(app.registerOrFail("Dana", "dana@example.com", "Passw0rd"), "register")

	_, env := app.login("dana", "Passw0rd")
	app.mustSuccess(env, "login")
	userID := decodeUser(t, env).ID
	app.mustSuccess(app.mustSelf(), "self before invalidation")

	// Bumping auth_version invalidates every existing session.
	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).
		Update("auth_version", gorm.Expr("auth_version + 1")).Error; err != nil {
		t.Fatalf("bump auth version: %v", err)
	}
	resp, env := app.self()
	app.mustCode(resp, env, http.StatusUnauthorized, "SESSION_EXPIRED", "self after auth version bump")

	// A fresh login works and carries the new auth version.
	_, env = app.login("dana", "Passw0rd")
	app.mustSuccess(env, "login after bump")

	// Disabling the account invalidates access and login.
	if err := model.DB.Model(&model.User{}).Where("id = ?", userID).
		Update("status", model.StatusDisabled).Error; err != nil {
		t.Fatalf("disable user: %v", err)
	}
	resp, env = app.self()
	app.mustCode(resp, env, http.StatusUnauthorized, "USER_DISABLED", "self when disabled")
	resp, env = app.login("dana", "Passw0rd")
	app.mustCode(resp, env, http.StatusOK, "USER_DISABLED", "login when disabled")
}

func (a *testApp) mustSelf() envelope {
	a.t.Helper()
	_, env := a.self()
	a.mustSuccess(env, "self")
	return env
}

func TestSessionExpiryAndSlidingTouch(t *testing.T) {
	app := newTestApp(t, nil)
	app.mustSuccess(app.registerOrFail("Eve", "eve@example.com", "Passw0rd"), "register")
	_, env := app.login("eve", "Passw0rd")
	app.mustSuccess(env, "login")
	userID := decodeUser(t, env).ID

	var session model.UserSession
	if err := model.DB.Where("user_id = ?", userID).First(&session).Error; err != nil {
		t.Fatalf("load session: %v", err)
	}

	// An elapsed last-seen interval triggers a sliding expiry.
	oldExpiry := session.ExpiresAt
	if err := model.DB.Model(&model.UserSession{}).Where("id = ?", session.ID).
		Update("last_seen_at", time.Now().UTC().Add(-2*time.Hour)).Error; err != nil {
		t.Fatalf("age session: %v", err)
	}
	app.mustSuccess(app.mustSelf(), "self with aged session")
	if err := model.DB.First(&session, "id = ?", session.ID).Error; err != nil {
		t.Fatalf("reload session: %v", err)
	}
	if !session.ExpiresAt.After(oldExpiry) {
		t.Fatalf("expires_at was not extended: old=%v new=%v", oldExpiry, session.ExpiresAt)
	}

	// An expired session is rejected.
	if err := model.DB.Model(&model.UserSession{}).Where("id = ?", session.ID).
		Update("expires_at", time.Now().UTC().Add(-time.Minute)).Error; err != nil {
		t.Fatalf("expire session: %v", err)
	}
	resp, env := app.self()
	app.mustCode(resp, env, http.StatusUnauthorized, "SESSION_EXPIRED", "self with expired session")
}

func TestOriginAndContentTypeChecks(t *testing.T) {
	app := newTestApp(t, nil)
	app.mustSuccess(app.registerOrFail("Frank", "frank@example.com", "Passw0rd"), "register")

	// Requests without Origin/Referer are allowed for non-browser clients.
	_, env := app.request(http.MethodPost, "/api/user/login",
		map[string]any{"username": "frank", "password": "Passw0rd"},
		func(r *http.Request) { r.Header.Del("Origin") })
	app.mustSuccess(env, "login without origin")

	resp, env := app.request(http.MethodPost, "/api/user/login",
		map[string]any{"username": "frank", "password": "Passw0rd"},
		func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") })
	app.mustCode(resp, env, http.StatusForbidden, "CROSS_ORIGIN_REQUEST", "evil origin")

	resp, env = app.request(http.MethodPost, "/api/user/login",
		map[string]any{"username": "frank", "password": "Passw0rd"},
		func(r *http.Request) {
			r.Header.Del("Origin")
			r.Header.Set("Referer", "https://evil.example/page")
		})
	app.mustCode(resp, env, http.StatusForbidden, "CROSS_ORIGIN_REQUEST", "evil referer")

	_, env = app.request(http.MethodPost, "/api/user/login",
		map[string]any{"username": "frank", "password": "Passw0rd"},
		func(r *http.Request) {
			r.Header.Del("Origin")
			r.Header.Set("Referer", app.cfg.PublicOrigin+"/sign-in")
		})
	app.mustSuccess(env, "same-origin referer")

	resp, env = app.request(http.MethodPost, "/api/user/login", map[string]any{"username": "frank"},
		func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") })
	app.mustCode(resp, env, http.StatusUnsupportedMediaType, "INVALID_REQUEST", "non-json content type")

	// Safe methods are never origin-checked.
	resp, _ = app.request(http.MethodGet, "/api/status", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status endpoint: %d", resp.StatusCode)
	}
}

func TestInvalidCredentialsAreIndistinguishable(t *testing.T) {
	app := newTestApp(t, nil)
	app.mustSuccess(app.registerOrFail("Grace", "grace@example.com", "Passw0rd"), "register")

	resp, env := app.login("grace", "WrongPass1")
	app.mustCode(resp, env, http.StatusOK, "INVALID_CREDENTIALS", "wrong password")

	resp, env = app.login("nobody", "WrongPass1")
	app.mustCode(resp, env, http.StatusOK, "INVALID_CREDENTIALS", "unknown user")
}

func TestOAuthDisabledProviderRedirects(t *testing.T) {
	app := newTestApp(t, nil)

	resp, _ := app.request(http.MethodGet, "/api/oauth/github/start", nil)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("start status = %d, want 302", resp.StatusCode)
	}
	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, "/sign-in?oauth_error=") {
		t.Fatalf("start redirect = %q, want /sign-in?oauth_error=...", location)
	}

	resp, _ = app.request(http.MethodGet, "/api/oauth/github/callback?code=x&state=y", nil)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("callback status = %d, want 302", resp.StatusCode)
	}
	if location := resp.Header.Get("Location"); !strings.Contains(location, "oauth_error=") {
		t.Fatalf("callback redirect = %q, want oauth_error", location)
	}
}

func TestBasePathAPIAndOAuthRedirect(t *testing.T) {
	app := newTestApp(t, func(c *config.Config) {
		c.BasePath = "/medrouter"
		c.PublicOrigin = "http://example.test/medrouter"
	})

	resp, env := app.request(http.MethodGet, "/api/status", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status status = %d", resp.StatusCode)
	}
	app.mustSuccess(env, "status under base path")

	app.mustSuccess(app.registerOrFail("Helen", "helen@example.com", "Passw0rd"), "register")
	_, env = app.login("helen", "Passw0rd")
	app.mustSuccess(env, "login under base path")

	resp, _ = app.request(http.MethodGet, "/api/oauth/github/start", nil)
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("oauth start status = %d", resp.StatusCode)
	}
	if location := resp.Header.Get("Location"); !strings.HasPrefix(location, "/medrouter/sign-in?oauth_error=") {
		t.Fatalf("base path redirect = %q", location)
	}
}
