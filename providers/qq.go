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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"

	"go.xiexianbin.cn/authkit/types"
)

// QQ 的流程也有些非标准，尤其是在获取 `openid` 的步骤。
//
// **关键点**:
// * 获取 `access_token` 后，需要额外调用一个接口获取 `openid`。
// * `openid` 返回的格式是 `callback( {"client_id":"...","openid":"..."} );`，需要特殊处理。
// * 同样，`UnionID` 是打通腾讯生态的关键。

type QQProvider struct {
	Name   string
	config *oauth2.Config
}

func NewQQProvider(cfg *types.OauthConfig) types.Provider {
	return &QQProvider{
		Name: types.QQ,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"get_user_info"}, // QQ scope
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://graph.qq.com/oauth2.0/authorize",
				TokenURL: "https://graph.qq.com/oauth2.0/token",
			},
		},
	}
}

func (p *QQProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *QQProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// QQ 的 token 返回是 text/plain, 类似 access_token=...&expires_in=...
	// 标准库的 Exchange 会失败，需要手动请求
	// The standard oauth2.Config.Exchange method expects a JSON response.
	// QQ's token endpoint returns a URL-encoded string.
	// Therefore, we need to manually perform the exchange for QQ.
	tokenURL := fmt.Sprintf(
		"%s?grant_type=authorization_code&client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		p.config.Endpoint.TokenURL,
		p.config.ClientID,
		p.config.ClientSecret,
		code,
		url.QueryEscape(p.config.RedirectURL),
	)

	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	if values.Get("access_token") == "" {
		return nil, fmt.Errorf("qq get access_token failed: %s", string(body))
	}

	token := &oauth2.Token{
		AccessToken:  values.Get("access_token"),
		RefreshToken: values.Get("refresh_token"),
	}
	return token, nil
}

// getOpenID 是 QQ 特有的步骤
func (p *QQProvider) getOpenID(accessToken string) (string, string, error) {
	openidURL := fmt.Sprintf("https://graph.qq.com/oauth2.0/me?access_token=%s&unionid=1", accessToken)
	resp, err := http.Get(openidURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	// 返回格式: callback( {"client_id":"...","openid":"...","unionid":"..."} );
	s := strings.Trim(string(body), "callback( );\n\r")

	var data struct {
		OpenID  string `json:"openid"`
		UnionID string `json:"unionid"`
		Error   int    `json:"error"`
		Msg     string `json:"error_description"`
	}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return "", "", err
	}
	if data.Error != 0 {
		return "", "", fmt.Errorf("qq get openid error: %s", data.Msg)
	}

	return data.OpenID, data.UnionID, nil
}

func (p *QQProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	openid, unionid, err := p.getOpenID(token.AccessToken)
	if err != nil {
		return nil, err
	}

	userInfoURL := fmt.Sprintf(
		"https://graph.qq.com/user/get_user_info?access_token=%s&oauth_consumer_key=%s&openid=%s",
		token.AccessToken,
		p.config.ClientID,
		openid,
	)

	resp, err := http.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var qqUser struct {
		Ret        int    `json:"ret"`
		Msg        string `json:"msg"`
		Nickname   string `json:"nickname"`
		Avatar     string `json:"figureurl_qq_2"` // 100x100
		AvatarFull string `json:"figureurl_qq_1"` // 40x40
	}

	if err := json.Unmarshal(body, &qqUser); err != nil {
		return nil, err
	}
	if qqUser.Ret != 0 {
		return nil, fmt.Errorf("qq get user info error: %s", qqUser.Msg)
	}

	// 优先使用 UnionID
	providerUserID := unionid
	if providerUserID == "" {
		providerUserID = openid
	}

	return &types.UserInfo{
		Provider:       p.Name,
		ProviderUserID: providerUserID,
		Name:           qqUser.Nickname,
		AvatarURL:      qqUser.Avatar,
		Email:          "", // QQ 不提供邮箱
	}, nil
}
