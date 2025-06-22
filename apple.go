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

package authkit

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
)

// It requires generating a JWT `client_secret` on the fly. User info is
// contained within the `id_token` returned during the token exchange.
// **CRITICAL NOTE**: Apple only provides user information (`name`, `email`)
// on the **very first** authorization. Your application MUST capture and
// save it then. Subsequent logins will not include this data.

type AppleProvider struct {
	Name        string
	config      *OauthConfig
	oauthConfig *oauth2.Config
}

func NewAppleProvider(cfg *OauthConfig) Provider {
	return &AppleProvider{
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

// GetAuthURL for Apple requires response_mode=form_post
func (p *AppleProvider) GetAuthURL(state string) string {
	return p.oauthConfig.AuthCodeURL(state, oauth2.SetAuthURLParam("response_mode", "form_post"))
}

// generateAppleClientSecret creates a JWT to be used as the client_secret
func (p *AppleProvider) generateAppleClientSecret() (string, error) {
	privateKeyBytes := []byte(p.config.Extra["AppPrivateKey"].(string))
	privateKey, err := jwt.ParseECPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse Apple private key: %w", err)
	}

	claims := &jwt.RegisteredClaims{
		Issuer:    p.config.Extra["TeamID"].(string),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Audience:  jwt.ClaimStrings{"https://appleid.apple.com"},
		Subject:   p.config.ClientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = p.config.Extra["KeyID"].(string)

	return token.SignedString(privateKey)
}

func (p *AppleProvider) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	clientSecret, err := p.generateAppleClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate Apple client secret: %w", err)
	}

	// Use the standard library to exchange, but provide the custom client secret
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, http.DefaultClient)
	token, err := p.oauthConfig.Exchange(ctx, code, oauth2.SetAuthURLParam("client_secret", clientSecret))

	if err != nil {
		// Try to parse Apple's specific error format
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

func (p *AppleProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("apple id_token not found in token")
	}

	// The user info is inside the ID Token JWT. We just need to decode it.
	// We don't need to verify the signature here because it came directly from Apple's server.
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid id_token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode apple id_token payload: %w", err)
	}

	var claims struct {
		Sub   string `json:"sub"` // The unique user ID
		Email string `json:"email"`
		// Name and other details might not be present on subsequent logins
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal apple id_token claims: %w", err)
	}

	// Note: You would also get the user's name from the initial POST request to your callback,
	// which is not part of this simplified interface.
	return &UserInfo{
		Provider:       "apple",
		ProviderUserID: claims.Sub,
		Email:          claims.Email,
		Name:           "", // Must be captured from the initial form post
		AvatarURL:      "",
	}, nil
}
