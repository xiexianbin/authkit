// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"go.xiexianbin.cn/authkit/types"
)

// GithubProvider https://github.com/login/oauth/.well-known/openid-configuration
type GithubProvider struct {
	Name   string
	config *oauth2.Config
}

// NewGithubProvider creates a new GitHub Provider instance
func NewGithubProvider(cfg *types.OauthConfig) types.Provider {
	return &GithubProvider{
		Name: types.GITHUB,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"read:user", "user:email"}, // ensure getting email
			Endpoint:     github.Endpoint,
		},
	}
}

func (p *GithubProvider) GetAuthURL(ctx context.Context, state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *GithubProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code, opts...)
}

func (p *GithubProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	client := p.config.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, err
	}

	// If the main API does not return an email, try fetching from /user/emails
	if githubUser.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer emailResp.Body.Close()
			emailBody, err := io.ReadAll(emailResp.Body)
			if err == nil {
				var emails []struct {
					Email    string `json:"email"`
					Primary  bool   `json:"primary"`
					Verified bool   `json:"verified"`
				}
				if json.Unmarshal(emailBody, &emails) == nil {
					for _, e := range emails {
						if e.Primary && e.Verified {
							githubUser.Email = e.Email
							break
						}
					}
				}
			}
		}
	}

	if githubUser.Name == "" {
		githubUser.Name = githubUser.Login
	}

	return &types.UserInfo{
		Provider:       types.GITHUB,
		ProviderUserID: fmt.Sprintf("%d", githubUser.ID),
		Email:          githubUser.Email,
		Name:           githubUser.Name,
		AvatarURL:      githubUser.AvatarURL,
		RawData:        githubUser,
	}, nil
}
