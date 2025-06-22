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
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

type FacebookProvider struct {
	Name   string
	config *oauth2.Config
}

func NewFacebookProvider(cfg *OauthConfig) Provider {
	return &FacebookProvider{
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"public_profile", "email"},
			Endpoint:     facebook.Endpoint,
		},
	}
}

func (p *FacebookProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *FacebookProvider) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return p.config.Exchange(context.Background(), code)
}

func (p *FacebookProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// Facebook requires specifying fields
	fields := "id,name,email,picture.type(large)"
	userInfoURL := fmt.Sprintf("https://graph.facebook.com/me?fields=%s&access_token=%s", fields, url.QueryEscape(token.AccessToken))

	resp, err := http.Get(userInfoURL)
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

	return &UserInfo{
		Provider:       "facebook",
		ProviderUserID: fbUser.ID,
		Email:          fbUser.Email,
		Name:           fbUser.Name,
		AvatarURL:      fbUser.Picture.Data.URL,
	}, nil
}
