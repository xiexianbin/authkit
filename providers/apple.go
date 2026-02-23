// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"go.xiexianbin.cn/authkit/types"
)

type AppleProvider struct {
	Name        string
	config      *types.OauthConfig
	oauthConfig *oauth2.Config
}

func NewAppleProvider(cfg *types.OauthConfig) types.Provider {
	return &AppleProvider{
		Name:   types.APPLE,
		config: cfg,
		oauthConfig: &oauth2.Config{
			ClientID:    cfg.ClientID,
			RedirectURL: cfg.RedirectURL,
			Scopes:      []string{"name", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://appleid.apple.com/auth/authorize",
				TokenURL: "https://appleid.apple.com/auth/token",
			},
		},
	}
}

func (p *AppleProvider) GetAuthURL(ctx context.Context, state string, opts ...oauth2.AuthCodeOption) string {
	authOpts := append(opts, oauth2.SetAuthURLParam("response_mode", "form_post"))
	return p.oauthConfig.AuthCodeURL(state, authOpts...)
}

func (p *AppleProvider) generateAppleClientSecret() (string, error) {
	keyContent, ok := p.config.Extra["AppPrivateKey"].(string)
	if !ok {
		return "", fmt.Errorf("AppPrivateKey not found in extra config")
	}
	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(keyContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse Apple private key: %w", err)
	}

	teamID, _ := p.config.Extra["TeamID"].(string)
	keyID, _ := p.config.Extra["KeyID"].(string)

	claims := &jwt.RegisteredClaims{
		Issuer:    teamID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Audience:  jwt.ClaimStrings{"https://appleid.apple.com"},
		Subject:   p.config.ClientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID

	return token.SignedString(privateKey)
}

func (p *AppleProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	clientSecret, err := p.generateAppleClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Apple client secret: %w", err)
	}

	// Add client_secret to options
	exchangeOpts := append(opts, oauth2.SetAuthURLParam("client_secret", clientSecret))

	// Ensure HTTP client is in context if not already
	if ctx.Value(oauth2.HTTPClient) == nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, http.DefaultClient)
	}

	token, err := p.oauthConfig.Exchange(ctx, code, exchangeOpts...)

	if err != nil {
		if e, ok := err.(*oauth2.RetrieveError); ok {
			var appleErr struct {
				Error string `json:"error"`
			}
			if json.Unmarshal(e.Body, &appleErr) == nil {
				return nil, fmt.Errorf("apple auth error: %s", appleErr.Error)
			}
		}
		return nil, err
	}

	return token, nil
}

func (p *AppleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	idTokenStr, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("apple id_token not found in token")
	}

	provider, err := oidc.NewProvider(ctx, "https://appleid.apple.com")
	if err != nil {
		return nil, fmt.Errorf("failed to get apple oidc provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: p.config.ClientID,
	})

	idToken, err := verifier.Verify(ctx, idTokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to verify apple id_token: %w", err)
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal apple id_token claims: %w", err)
	}

	var rawData map[string]interface{}
	if err := idToken.Claims(&rawData); err != nil {
		rawData = make(map[string]interface{})
	}

	return &types.UserInfo{
		Provider:       types.APPLE,
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		Name:           "", // Must be captured from the initial form post
		RawData:        rawData,
	}, nil
}
