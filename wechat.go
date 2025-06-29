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
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

// 微信的流程略有不同，特别是获取用户信息的步骤。我们将实现 PC 网站扫码登录。
//
// **关键点**:
// * `scope` 固定为 `snsapi_login`。
// * 获取 `access_token` 的同时会得到 `openid`。
// * `UnionID` 是打通微信生态（公众号、小程序、网站）的关键，应优先作为 `ProviderUserID`。如果应用未加入微信开放平台，可能无法获取 `UnionID`，此时只能用 `openid`。
// * 微信通常 **不提供** 用户 Email。

type WechatProvider struct {
	Name   string
	config *oauth2.Config
}

// NewWechatProvider 创建一个新的微信Provider实例
func NewWechatProvider(cfg *OauthConfig) Provider {
	return &WechatProvider{
		Name: WECHAT,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"snsapi_login"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://open.weixin.qq.com/connect/qrconnect",
				TokenURL: "https://api.weixin.qq.com/sns/oauth2/access_token",
			},
		},
	}
}

// GetAuthURL 微信需要手动拼接 response_type=code
func (p *WechatProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("%s&response_type=code&state=%s#wechat_redirect",
		p.config.AuthCodeURL(state), state)
}

func (p *WechatProvider) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	// 微信的 token URL 参数是 appid 和 secret，而不是 client_id 和 client_secret
	// x/oauth2 库在 Exchange 时会使用 ClientID 和 ClientSecret，所以我们需要手动构造请求
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		p.config.ClientID,
		p.config.ClientSecret,
		code,
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

	var tokenData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		OpenID       string `json:"openid"`
		Scope        string `json:"scope"`
		UnionID      string `json:"unionid"`
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &tokenData); err != nil {
		return nil, err
	}

	if tokenData.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %s", tokenData.ErrMsg)
	}

	token := &oauth2.Token{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
	}
	// 将 openid 和 unionid 临时存放在 Extra 中，以便 GetUserInfo 使用
	return token.WithExtra(map[string]interface{}{
		"openid":  tokenData.OpenID,
		"unionid": tokenData.UnionID,
	}), nil
}

func (p *WechatProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	openid := token.Extra("openid").(string)

	userInfoURL := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", token.AccessToken, openid)
	resp, err := http.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wechatUser struct {
		OpenID     string   `json:"openid"`
		Nickname   string   `json:"nickname"`
		Sex        int      `json:"sex"`
		Province   string   `json:"province"`
		City       string   `json:"city"`
		Country    string   `json:"country"`
		HeadImgURL string   `json:"headimgurl"`
		Privilege  []string `json:"privilege"`
		UnionID    string   `json:"unionid"`
		ErrCode    int      `json:"errcode"`
		ErrMsg     string   `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &wechatUser); err != nil {
		return nil, err
	}

	if wechatUser.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %s", wechatUser.ErrMsg)
	}

	// 优先使用 UnionID 作为唯一标识符
	providerUserID := wechatUser.UnionID
	if providerUserID == "" {
		providerUserID = wechatUser.OpenID
	}

	return &UserInfo{
		Provider:       "wechat",
		ProviderUserID: providerUserID,
		Name:           wechatUser.Nickname,
		AvatarURL:      wechatUser.HeadImgURL,
		Email:          "", // 微信不提供邮箱
	}, nil
}
