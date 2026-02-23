// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"go.xiexianbin.cn/authkit/types"
)

// GoogleProvider https://accounts.google.com/.well-known/openid-configuration
type GoogleProvider struct {
	Name   string
	config *oauth2.Config
}

func NewGoogleProvider(cfg *types.OauthConfig) types.Provider {
	return &GoogleProvider{
		Name: types.GOOGLE,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *GoogleProvider) GetAuthURL(ctx context.Context, state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *GoogleProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code, opts...)
}

func (p *GoogleProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	if googleUser.Email != "" && !googleUser.VerifiedEmail {
		return nil, fmt.Errorf("google email is not verified")
	}

	return &types.UserInfo{
		Provider:       types.GOOGLE,
		ProviderUserID: googleUser.ID,
		Email:          googleUser.Email,
		Name:           googleUser.Name,
		AvatarURL:      googleUser.Picture,
		RawData:        googleUser,
	}, nil
}
