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

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// GithubProvider https://github.com/login/oauth/.well-known/openid-configuration
type GithubProvider struct {
	Name   string
	config *oauth2.Config
}

// NewGithubProvider 创建一个新的 GitHub Provider实例
func NewGithubProvider(cfg *OauthConfig) Provider {
	return &GithubProvider{
		Name: GITHUB,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"read:user", "user:email"}, // 确保获取 email
			Endpoint:     github.Endpoint,
		},
	}
}

func (p *GithubProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state)
}

func (p *GithubProvider) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return p.config.Exchange(context.Background(), code)
}

func (p *GithubProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(context.Background(), token)
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

	// 如果主 API 没有返回 email，可以尝试从 /user/emails 获取
	if githubUser.Email == "" {
		// ... 此处可以添加一个请求去获取 email 列表并找到主 email
	}

	if githubUser.Name == "" {
		githubUser.Name = githubUser.Login
	}

	return &UserInfo{
		Provider:       "github",
		ProviderUserID: fmt.Sprintf("%d", githubUser.ID),
		Email:          githubUser.Email,
		Name:           githubUser.Name,
		AvatarURL:      githubUser.AvatarURL,
	}, nil
}
