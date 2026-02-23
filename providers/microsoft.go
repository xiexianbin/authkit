// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package providers

import (
	"context"
	"encoding/json"
	"io"

	"golang.org/x/oauth2"

	"go.xiexianbin.cn/authkit/types"
)

type MicrosoftProvider struct {
	Name   string
	config *oauth2.Config
}

func NewMicrosoftProvider(cfg *types.OauthConfig) types.Provider {
	return &MicrosoftProvider{
		Name: types.MICROSOFT,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"User.Read", "openid", "profile", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
				TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			},
		},
	}
}

func (p *MicrosoftProvider) GetAuthURL(ctx context.Context, state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *MicrosoftProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code, opts...)
}

func (p *MicrosoftProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	client := p.config.Client(ctx, token)
	userInfoURL := "https://graph.microsoft.com/v1.0/me"

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msUser struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Mail        string `json:"mail"`
	}

	if err := json.Unmarshal(body, &msUser); err != nil {
		return nil, err
	}

	return &types.UserInfo{
		Provider:       types.MICROSOFT,
		ProviderUserID: msUser.ID,
		Name:           msUser.DisplayName,
		Email:          msUser.Mail,
		AvatarURL:      "", // Microsoft Graph requires a separate call for the photo
		RawData:        msUser,
	}, nil
}
