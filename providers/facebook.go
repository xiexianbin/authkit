// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"

	"go.xiexianbin.cn/authkit/types"
)

type FacebookProvider struct {
	Name   string
	config *oauth2.Config
}

func NewFacebookProvider(cfg *types.OauthConfig) types.Provider {
	return &FacebookProvider{
		Name: types.FACEBOOK,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"public_profile", "email"},
			Endpoint:     facebook.Endpoint,
		},
	}
}

func (p *FacebookProvider) GetAuthURL(ctx context.Context, state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *FacebookProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code, opts...)
}

func (p *FacebookProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	client := p.config.Client(ctx, token)
	// Facebook requires specifying fields
	fields := "id,name,email,picture.type(large)"
	userInfoURL := fmt.Sprintf("https://graph.facebook.com/me?fields=%s&access_token=%s", fields, url.QueryEscape(token.AccessToken))

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fbUser struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.Unmarshal(body, &fbUser); err != nil {
		return nil, err
	}

	return &types.UserInfo{
		Provider:       types.FACEBOOK,
		ProviderUserID: fbUser.ID,
		Email:          fbUser.Email,
		Name:           fbUser.Name,
		AvatarURL:      fbUser.Picture.Data.URL,
		RawData:        fbUser,
	}, nil
}
