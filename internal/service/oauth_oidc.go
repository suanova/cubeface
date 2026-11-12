package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/suanova/medrouter/internal/common"
)

type oidcProvider struct {
	id       string
	name     string
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func newOIDCProvider(ctx context.Context, id, name, issuer, clientID, clientSecret, redirectURL string, scopes []string) (*oidcProvider, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("discover issuer %s: %w", issuer, err)
	}
	endpoint := provider.Endpoint()
	return &oidcProvider{
		id:       id,
		name:     name,
		verifier: provider.Verifier(&oidc.Config{ClientID: clientID}),
		oauth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     endpoint,
		},
	}, nil
}

func (p *oidcProvider) ID() string          { return p.id }
func (p *oidcProvider) DisplayName() string { return p.name }

func (p *oidcProvider) AuthCodeURL(state, nonce, verifier string) string {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	if nonce != "" {
		opts = append(opts, oauth2.SetAuthURLParam("nonce", nonce))
	}
	return p.oauth.AuthCodeURL(state, opts...)
}

type oidcClaims struct {
	Subject           string `json:"sub"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	Nonce             string `json:"nonce"`
}

func (p *oidcProvider) Exchange(ctx context.Context, code, verifier, nonce string) (*OAuthIdentity, error) {
	token, err := p.oauth.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", verifier))
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, fmt.Errorf("token response has no id_token")
	}
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("verify id_token: %w", err)
	}
	var claims oidcClaims
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("decode id_token claims: %w", err)
	}
	if nonce != "" && !common.ConstantTimeEqual(claims.Nonce, nonce) {
		return nil, fmt.Errorf("id_token nonce mismatch")
	}
	if claims.Subject == "" {
		return nil, fmt.Errorf("id_token has no subject")
	}
	email := strings.TrimSpace(claims.Email)
	if email == "" {
		var extra struct {
			Email string `json:"email"`
		}
		if err := idToken.Claims(&extra); err == nil {
			email = strings.TrimSpace(extra.Email)
		}
	}
	return &OAuthIdentity{
		ProviderUserID: claims.Subject,
		Email:          email,
		EmailVerified:  claims.EmailVerified,
		Username:       claims.PreferredUsername,
		DisplayName:    claims.Name,
	}, nil
}
