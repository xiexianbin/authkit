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
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"

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
			Endpoint:     microsoft.LiveConnectEndpoint,
		},
	}
}

func (p *MicrosoftProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
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
		Provider:       "microsoft",
		ProviderUserID: msUser.ID,
		Name:           msUser.DisplayName,
		Email:          msUser.Mail,
		AvatarURL:      "", // Microsoft Graph requires a separate call for the photo
	}, nil
}
