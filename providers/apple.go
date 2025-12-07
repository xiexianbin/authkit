// Copyright 2025 xiexianbin<me@xiexianbin.cn>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package providers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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

func (p *AppleProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
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
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("apple id_token not found in token")
	}

	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid id_token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode apple id_token payload: %w", err)
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal apple id_token claims: %w", err)
	}

	return &types.UserInfo{
		Provider:       "apple",
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		Name:           "", // Must be captured from the initial form post
	}, nil
}
