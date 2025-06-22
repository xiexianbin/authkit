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
	"encoding/json"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	Name   string
	config *oauth2.Config
}

func NewGoogleProvider(cfg *OauthConfig) Provider {
	return &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *GoogleProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *GoogleProvider) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return p.config.Exchange(context.Background(), code)
}

func (p *GoogleProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(context.Background(), token)
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
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &UserInfo{
		Provider:       "google",
		ProviderUserID: googleUser.ID,
		Email:          googleUser.Email,
		Name:           googleUser.Name,
		AvatarURL:      googleUser.Picture,
	}, nil
}
