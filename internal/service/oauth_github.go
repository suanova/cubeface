package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"

	"github.com/cubeface/cubeface/internal/config"
)

const (
	githubAuthorizeURL = "https://github.com/login/oauth/authorize"
	githubTokenURL     = "https://github.com/login/oauth/access_token"
	githubAPIBase      = "https://api.github.com"
)

type githubProvider struct {
	cfg   *config.Config
	oauth *oauth2.Config
}

func newGitHubProvider(cfg *config.Config) *githubProvider {
	return &githubProvider{
		cfg: cfg,
		oauth: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.CallbackURL("github"),
			Scopes:       []string{"read:user", "user:email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  githubAuthorizeURL,
				TokenURL: githubTokenURL,
			},
		},
	}
}

func (p *githubProvider) ID() string          { return "github" }
func (p *githubProvider) DisplayName() string { return "GitHub" }

func (p *githubProvider) AuthCodeURL(state, _ string, _ string) string {
	return p.oauth.AuthCodeURL(state)
}

type githubUser struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type githubEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (p *githubProvider) Exchange(ctx context.Context, code, _, _ string) (*OAuthIdentity, error) {
	token, err := p.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	client := p.oauth.Client(ctx, token)

	var user githubUser
	if err := getJSON(ctx, client, githubAPIBase+"/user", &user); err != nil {
		return nil, fmt.Errorf("fetch github user: %w", err)
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("github user response has no id")
	}

	identity := &OAuthIdentity{
		ProviderUserID: strconv.FormatInt(user.ID, 10),
		Username:       user.Login,
		DisplayName:    user.Name,
	}

	// Prefer a primary verified address; fall back to any verified address.
	var emails []githubEmail
	if err := getJSON(ctx, client, githubAPIBase+"/user/emails", &emails); err == nil {
		var fallback string
		for _, e := range emails {
			if !e.Verified {
				continue
			}
			if e.Primary {
				identity.Email = e.Email
				identity.EmailVerified = true
				break
			}
			if fallback == "" {
				fallback = e.Email
			}
		}
		if identity.Email == "" && fallback != "" {
			identity.Email = fallback
			identity.EmailVerified = true
		}
	}
	if identity.Email == "" && user.Email != "" {
		// The public profile email is used for display, but GitHub does not
		// expose its verification state here, so it cannot be trusted for
		// automatic account linking.
		identity.Email = user.Email
		identity.EmailVerified = false
	}
	return identity, nil
}

func getJSON(ctx context.Context, client *http.Client, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "CubeFace")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("github api %s returned %d", url, resp.StatusCode)
	}
	return json.Unmarshal(body, out)
}
