package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cubeface/cubeface/internal/common"
	"github.com/cubeface/cubeface/internal/config"
	"github.com/cubeface/cubeface/internal/model"
)

const oauthStateCookieName = "cubeface_oauth_state"
const oauthStateTTL = 10 * time.Minute

// OAuthIdentity is the normalized result of a provider callback.
type OAuthIdentity struct {
	Provider       string
	ProviderUserID string
	Email          string
	EmailVerified  bool
	Username       string
	DisplayName    string
}

type OAuthProvider interface {
	ID() string
	DisplayName() string
	AuthCodeURL(state, nonce, verifier string) string
	Exchange(ctx context.Context, code, verifier, nonce string) (*OAuthIdentity, error)
}

type ProviderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OAuthService struct {
	cfg       *config.Config
	providers map[string]OAuthProvider
	order     []string
	users     *UserService
	sessions  *SessionService
}

type oauthStatePayload struct {
	Provider string `json:"provider"`
	State    string `json:"state"`
	Verifier string `json:"verifier"`
	Nonce    string `json:"nonce"`
	Redirect string `json:"redirect"`
	Expires  int64  `json:"exp"`
}

func NewOAuthService(cfg *config.Config, users *UserService, sessions *SessionService) *OAuthService {
	s := &OAuthService{
		cfg:       cfg,
		providers: make(map[string]OAuthProvider),
		users:     users,
		sessions:  sessions,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if cfg.GitHubClientID != "" && cfg.GitHubClientSecret != "" {
		s.register(newGitHubProvider(cfg))
	}
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		p, err := newOIDCProvider(ctx, "google", "Google", "https://accounts.google.com",
			cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.CallbackURL("google"),
			[]string{"openid", "email", "profile"})
		if err != nil {
			log.Printf("WARN oauth provider google disabled: %v", err)
		} else {
			s.register(p)
		}
	}
	if cfg.OIDCIssuerURL != "" && cfg.OIDCClientID != "" {
		scopes := strings.Fields(cfg.OIDCScopes)
		if len(scopes) == 0 {
			scopes = []string{"openid", "email", "profile"}
		}
		p, err := newOIDCProvider(ctx, "oidc", cfg.OIDCDisplayName, cfg.OIDCIssuerURL,
			cfg.OIDCClientID, cfg.OIDCClientSecret, cfg.CallbackURL("oidc"), scopes)
		if err != nil {
			log.Printf("WARN oauth provider oidc disabled: %v", err)
		} else {
			s.register(p)
		}
	}
	return s
}

func (s *OAuthService) register(p OAuthProvider) {
	s.providers[p.ID()] = p
	s.order = append(s.order, p.ID())
}

func (s *OAuthService) Providers() []ProviderInfo {
	out := make([]ProviderInfo, 0, len(s.order))
	for _, id := range s.order {
		p := s.providers[id]
		out = append(out, ProviderInfo{ID: p.ID(), Name: p.DisplayName()})
	}
	return out
}

func (s *OAuthService) Enabled(provider string) bool {
	_, ok := s.providers[provider]
	return ok
}

// StateCookieName is the name of the signed OAuth state cookie.
func (s *OAuthService) StateCookieName() string { return oauthStateCookieName }

// Start builds the provider authorization URL and the signed state cookie.
func (s *OAuthService) Start(provider, redirect string) (string, *http.Cookie, *common.APIError) {
	p, ok := s.providers[provider]
	if !ok {
		return "", nil, common.NewAPIError("OAUTH_PROVIDER_DISABLED", "This sign-in provider is not configured")
	}

	state, err := common.RandomToken(32)
	if err != nil {
		return "", nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start sign-in")
	}
	verifier, err := common.RandomToken(32)
	if err != nil {
		return "", nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start sign-in")
	}
	nonce, err := common.RandomToken(16)
	if err != nil {
		return "", nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start sign-in")
	}

	payload := oauthStatePayload{
		Provider: provider,
		State:    state,
		Verifier: verifier,
		Nonce:    nonce,
		Redirect: SanitizeRedirect(redirect),
		Expires:  time.Now().Add(oauthStateTTL).Unix(),
	}
	cookie, err := s.sealStateCookie(payload)
	if err != nil {
		return "", nil, common.NewAPIErrorStatus(http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to start sign-in")
	}
	return p.AuthCodeURL(state, nonce, verifier), cookie, nil
}

// Complete validates the callback and returns a signed-in user.
func (s *OAuthService) Complete(ctx context.Context, provider, code, rawStateCookie, ip, userAgent string) (string, *model.User, *SessionResult, *common.APIError) {
	p, ok := s.providers[provider]
	if !ok {
		return "", nil, nil, common.NewAPIError("OAUTH_PROVIDER_DISABLED", "This sign-in provider is not configured")
	}
	payload, apiErr := s.openStateCookie(rawStateCookie, provider)
	if apiErr != nil {
		return "", nil, nil, apiErr
	}
	if strings.TrimSpace(code) == "" {
		return "", nil, nil, common.NewAPIError("OAUTH_STATE_INVALID", "The sign-in request was rejected by the provider")
	}

	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	identity, err := p.Exchange(ctx, code, payload.Verifier, payload.Nonce)
	if err != nil {
		log.Printf("WARN oauth exchange failed provider=%s: %v", provider, err)
		return "", nil, nil, common.NewAPIError("OAUTH_EXCHANGE_FAILED", "Failed to complete sign-in with the provider")
	}
	identity.Provider = provider
	if identity.ProviderUserID == "" {
		return "", nil, nil, common.NewAPIError("OAUTH_EXCHANGE_FAILED", "The provider did not return a user identifier")
	}

	user, apiErr := s.users.FindOrCreateByOAuth(*identity)
	if apiErr != nil {
		return "", nil, nil, apiErr
	}

	method := provider
	if method == "google" || method == "oidc" {
		method = model.LoginMethodOIDC
	}
	result, apiErr := s.sessions.Create(user, method, ip, userAgent)
	if apiErr != nil {
		return "", nil, nil, apiErr
	}
	now := time.Now().UTC()
	_ = model.DB.Model(&model.User{}).Where("id = ?", user.ID).Update("last_login_at", now).Error
	user.LastLoginAt = &now

	return payload.Redirect, user, result, nil
}

func (s *OAuthService) ClearStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     s.cfg.BasePath + "/api/oauth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
}

func (s *OAuthService) sealStateCookie(payload oauthStatePayload) (*http.Cookie, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	signature := common.HMACHex(s.cfg.SessionSecret, "oauth-state:"+encoded)
	return &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    encoded + "." + signature,
		Path:     s.cfg.BasePath + "/api/oauth",
		MaxAge:   int(oauthStateTTL.Seconds()),
		HttpOnly: true,
		Secure:   s.cfg.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func (s *OAuthService) openStateCookie(value, provider string) (*oauthStatePayload, *common.APIError) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, common.NewAPIError("OAUTH_STATE_INVALID", "Sign-in session expired, please try again")
	}
	expected := common.HMACHex(s.cfg.SessionSecret, "oauth-state:"+parts[0])
	if !common.ConstantTimeEqual(expected, parts[1]) {
		return nil, common.NewAPIError("OAUTH_STATE_INVALID", "Sign-in state could not be verified, please try again")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, common.NewAPIError("OAUTH_STATE_INVALID", "Sign-in state could not be verified, please try again")
	}
	var payload oauthStatePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, common.NewAPIError("OAUTH_STATE_INVALID", "Sign-in state could not be verified, please try again")
	}
	if payload.Provider != provider {
		return nil, common.NewAPIError("OAUTH_STATE_INVALID", "Sign-in state does not match the provider")
	}
	if time.Now().Unix() > payload.Expires {
		return nil, common.NewAPIError("OAUTH_STATE_INVALID", "Sign-in session expired, please try again")
	}
	payload.Redirect = SanitizeRedirect(payload.Redirect)
	return &payload, nil
}

// SanitizeRedirect only permits same-origin absolute paths.
func SanitizeRedirect(raw string) string {
	const fallback = "/dashboard"
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return fallback
	}
	if strings.ContainsAny(raw, "\\\r\n\t") {
		return fallback
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.Scheme != "" {
		return fallback
	}
	return raw
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
