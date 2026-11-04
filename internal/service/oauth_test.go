package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cubeface/cubeface/internal/common"
	"github.com/cubeface/cubeface/internal/config"
	"github.com/cubeface/cubeface/internal/model"
)

type fakeProvider struct {
	identity OAuthIdentity
	err      error
}

func (f *fakeProvider) ID() string          { return "fake" }
func (f *fakeProvider) DisplayName() string { return "Fake" }
func (f *fakeProvider) AuthCodeURL(state, nonce, verifier string) string {
	return "https://provider.example/authorize?state=" + state
}
func (f *fakeProvider) Exchange(context.Context, string, string, string) (*OAuthIdentity, error) {
	if f.err != nil {
		return nil, f.err
	}
	identity := f.identity
	return &identity, nil
}

func newOAuthTest(t *testing.T) (*OAuthService, *config.Config) {
	t.Helper()
	cfg := &config.Config{
		AppEnv:                   "test",
		SystemName:               "CubeFace",
		PublicOrigin:             "http://example.test",
		SessionSecret:            "oauth-test-session-secret-0123456789",
		SessionCookieName:        "cubeface_session",
		SessionTTL:               time.Hour,
		SessionTouchInterval:     time.Minute,
		RegisterEnabled:          true,
		SQLitePath:               filepath.Join(t.TempDir(), "oauth.db"),
		DBMaxOpenConns:           1,
		EmailVerificationEnabled: false,
		EmailDebugMode:           true,
		VerificationCodeTTL:      time.Minute,
		VerificationMaxAttempts:  5,
		VerificationResend:       time.Minute,
		VerificationIPHourly:     10,
		LoginRatePerMinute:       100,
		RegisterRatePerHour:      100,
	}
	if err := model.InitDB(cfg); err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() {
		_ = model.Close()
		model.DB = nil
	})

	sessions := NewSessionService(cfg)
	email := NewEmailService(cfg)
	verification := NewVerificationService(cfg, email)
	users := NewUserService(cfg, sessions, verification)
	return NewOAuthService(cfg, users, sessions), cfg
}

func (s *OAuthService) startAndComplete(t *testing.T, provider, redirect string) (string, *model.User, *SessionResult, *common.APIError) {
	t.Helper()
	_, cookie, apiErr := s.Start(provider, redirect)
	if apiErr != nil {
		return "", nil, nil, apiErr
	}
	return s.Complete(context.Background(), provider, "code", cookie.Value, "192.0.2.1", "test-agent")
}

func createLocalUser(t *testing.T, username, email string) *model.User {
	t.Helper()
	hash, err := common.HashPassword("Passw0rd")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	now := time.Now().UTC()
	user := &model.User{
		Username: username, DisplayName: username, PasswordHash: hash, Email: email,
		Role: model.RoleCommon, Status: model.StatusEnabled, AuthVersion: 1,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := model.DB.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func TestSanitizeRedirect(t *testing.T) {
	for raw, want := range map[string]string{
		"/models":          "/models",
		"/models?q=x#top":  "/models?q=x#top",
		"":                 "/dashboard",
		"   ":              "/dashboard",
		"https://evil.com": "/dashboard",
		"//evil.com":       "/dashboard",
		"/\\evil.com":      "/dashboard",
		"/a\tb":            "/dashboard",
		"javascript:x":     "/dashboard",
		"models":           "/dashboard",
	} {
		if got := SanitizeRedirect(raw); got != want {
			t.Errorf("SanitizeRedirect(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestOAuthStateValidation(t *testing.T) {
	svc, _ := newOAuthTest(t)
	svc.register(&fakeProvider{identity: OAuthIdentity{ProviderUserID: "state-user"}})

	_, cookie, apiErr := svc.Start("fake", "/models")
	if apiErr != nil {
		t.Fatalf("Start: %v", apiErr)
	}

	// A tampered signature is rejected.
	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		t.Fatalf("cookie value = %q", cookie.Value)
	}
	sig := []byte(parts[1])
	if sig[0] == 'a' {
		sig[0] = 'b'
	} else {
		sig[0] = 'a'
	}
	_, _, _, apiErr = svc.Complete(context.Background(), "fake", "code", parts[0]+"."+string(sig), "", "")
	if apiErr == nil || apiErr.Code != "OAUTH_STATE_INVALID" {
		t.Fatalf("tampered state = %+v", apiErr)
	}

	// A cookie issued for another provider cannot be used.
	other, err := svc.sealStateCookie(oauthStatePayload{
		Provider: "other", State: "s", Verifier: "v", Nonce: "n",
		Redirect: "/dashboard", Expires: time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	_, _, _, apiErr = svc.Complete(context.Background(), "fake", "code", other.Value, "", "")
	if apiErr == nil || apiErr.Code != "OAUTH_STATE_INVALID" {
		t.Fatalf("provider mismatch = %+v", apiErr)
	}

	// Expired states are rejected.
	expired, err := svc.sealStateCookie(oauthStatePayload{
		Provider: "fake", State: "s", Verifier: "v", Nonce: "n",
		Redirect: "/dashboard", Expires: time.Now().Add(-time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("seal: %v", err)
	}
	_, _, _, apiErr = svc.Complete(context.Background(), "fake", "code", expired.Value, "", "")
	if apiErr == nil || apiErr.Code != "OAUTH_STATE_INVALID" {
		t.Fatalf("expired state = %+v", apiErr)
	}

	// The sanitized redirect survives a valid round trip.
	_, cookie, apiErr = svc.Start("fake", "https://evil.example")
	if apiErr != nil {
		t.Fatalf("Start: %v", apiErr)
	}
	redirect, _, _, apiErr := svc.Complete(context.Background(), "fake", "code", cookie.Value, "", "")
	if apiErr != nil {
		t.Fatalf("Complete: %v", apiErr)
	}
	if redirect != "/dashboard" {
		t.Fatalf("redirect = %q, want /dashboard", redirect)
	}
}

func TestOAuthCreatesAndReusesAccount(t *testing.T) {
	svc, _ := newOAuthTest(t)
	svc.register(&fakeProvider{identity: OAuthIdentity{
		ProviderUserID: "github-1",
		Email:          "oauth@example.com",
		EmailVerified:  true,
		Username:       "OAuth User",
		DisplayName:    "OAuth User",
	}})

	redirect, user, result, apiErr := svc.startAndComplete(t, "fake", "/dashboard")
	if apiErr != nil {
		t.Fatalf("first sign-in: %v", apiErr)
	}
	if redirect != "/dashboard" || result == nil || result.Token == "" {
		t.Fatalf("unexpected first sign-in result: redirect=%q result=%+v", redirect, result)
	}
	if user.Username != "oauth-user" || user.DisplayName != "OAuth User" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if !user.EmailVerified() {
		t.Fatalf("provider-verified email should be marked verified")
	}

	_, again, _, apiErr := svc.startAndComplete(t, "fake", "/dashboard")
	if apiErr != nil {
		t.Fatalf("second sign-in: %v", apiErr)
	}
	if again.ID != user.ID {
		t.Fatalf("second sign-in created a new user: %d vs %d", again.ID, user.ID)
	}

	var users, bindings int64
	if err := model.DB.Model(&model.User{}).Count(&users).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := model.DB.Model(&model.OAuthBinding{}).Count(&bindings).Error; err != nil {
		t.Fatalf("count bindings: %v", err)
	}
	if users != 1 || bindings != 1 {
		t.Fatalf("users=%d bindings=%d, want 1/1", users, bindings)
	}
}

func TestOAuthBindsVerifiedEmail(t *testing.T) {
	svc, _ := newOAuthTest(t)
	local := createLocalUser(t, "localuser", "bind@example.com")
	svc.register(&fakeProvider{identity: OAuthIdentity{
		ProviderUserID: "google-1",
		Email:          "BIND@example.com",
		EmailVerified:  true,
		Username:       "someone-else",
	}})

	_, user, _, apiErr := svc.startAndComplete(t, "fake", "/dashboard")
	if apiErr != nil {
		t.Fatalf("sign-in: %v", apiErr)
	}
	if user.ID != local.ID {
		t.Fatalf("expected binding to existing user %d, got %d", local.ID, user.ID)
	}
	var stored model.User
	if err := model.DB.First(&stored, local.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if !stored.EmailVerified() {
		t.Fatalf("bound email should be marked verified")
	}
	var binding model.OAuthBinding
	if err := model.DB.Where("provider = ? AND provider_user_id = ?", "fake", "google-1").First(&binding).Error; err != nil {
		t.Fatalf("binding: %v", err)
	}
	if binding.UserID != local.ID {
		t.Fatalf("binding user = %d, want %d", binding.UserID, local.ID)
	}
}

func TestOAuthUsernameConflictGetsSuffix(t *testing.T) {
	svc, _ := newOAuthTest(t)
	createLocalUser(t, "conflict", "taken@example.com")
	svc.register(&fakeProvider{identity: OAuthIdentity{
		ProviderUserID: "gh-conflict",
		Email:          "different@example.com",
		EmailVerified:  true,
		Username:       "conflict",
	}})

	_, user, _, apiErr := svc.startAndComplete(t, "fake", "/dashboard")
	if apiErr != nil {
		t.Fatalf("sign-in: %v", apiErr)
	}
	if user.Username == "conflict" || !strings.HasPrefix(user.Username, "conflict-") {
		t.Fatalf("username = %q, want conflict-<suffix>", user.Username)
	}
	if err := common.ValidateUsername(user.Username); err != nil {
		t.Fatalf("generated username is invalid: %v", err)
	}
}

func TestOAuthRejectsDisabledAccount(t *testing.T) {
	svc, _ := newOAuthTest(t)
	local := createLocalUser(t, "disabled", "disabled@example.com")
	if err := model.DB.Model(&model.User{}).Where("id = ?", local.ID).
		Update("status", model.StatusDisabled).Error; err != nil {
		t.Fatalf("disable user: %v", err)
	}
	if err := model.DB.Create(&model.OAuthBinding{
		Provider: "fake", ProviderUserID: "gh-disabled", UserID: local.ID, Email: "disabled@example.com",
	}).Error; err != nil {
		t.Fatalf("create binding: %v", err)
	}
	svc.register(&fakeProvider{identity: OAuthIdentity{
		ProviderUserID: "gh-disabled", Email: "disabled@example.com", EmailVerified: true,
	}})

	_, _, _, apiErr := svc.startAndComplete(t, "fake", "/dashboard")
	if apiErr == nil || apiErr.Code != "USER_DISABLED" {
		t.Fatalf("disabled account = %+v", apiErr)
	}
}

func TestOAuthProviderNotConfigured(t *testing.T) {
	svc, _ := newOAuthTest(t)
	_, _, apiErr := svc.Start("github", "/dashboard")
	if apiErr == nil || apiErr.Code != "OAUTH_PROVIDER_DISABLED" {
		t.Fatalf("Start = %+v", apiErr)
	}
	_, _, _, apiErr = svc.Complete(context.Background(), "github", "code", "", "", "")
	if apiErr == nil || apiErr.Code != "OAUTH_PROVIDER_DISABLED" {
		t.Fatalf("Complete = %+v", apiErr)
	}
}
