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

	"golang.org/x/oauth2"

	"go.xiexianbin.cn/authkit/types"
)

type WechatProvider struct {
	Name   string
	config *oauth2.Config
}

func NewWechatProvider(cfg *types.OauthConfig) types.Provider {
	return &WechatProvider{
		Name: types.WECHAT,
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

func (p *WechatProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	// Wechat requires #wechat_redirect at the end
	url := p.config.AuthCodeURL(state, opts...)
	return url + "#wechat_redirect"
}

func (p *WechatProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// Wechat uses appid/secret instead of client_id/client_secret
	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		p.config.ClientID,
		p.config.ClientSecret,
		code,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
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
		// ExpiresIn is int seconds
		// Expiry: time.Now().Add(time.Duration(tokenData.ExpiresIn) * time.Second),
	}
	// Store openid/unionid in Extra
	return token.WithExtra(map[string]interface{}{
		"openid":  tokenData.OpenID,
		"unionid": tokenData.UnionID,
	}), nil
}

func (p *WechatProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	openid, ok := token.Extra("openid").(string)
	if !ok {
		return nil, fmt.Errorf("openid not found in token")
	}

	userInfoURL := fmt.Sprintf("https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s", token.AccessToken, openid)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var wechatUser struct {
		OpenID     string `json:"openid"`
		Nickname   string `json:"nickname"`
		HeadImgURL string `json:"headimgurl"`
		UnionID    string `json:"unionid"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &wechatUser); err != nil {
		return nil, err
	}

	if wechatUser.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %s", wechatUser.ErrMsg)
	}

	providerUserID := wechatUser.UnionID
	if providerUserID == "" {
		providerUserID = wechatUser.OpenID
	}

	return &types.UserInfo{
		Provider:       "wechat",
		ProviderUserID: providerUserID,
		Name:           wechatUser.Nickname,
		AvatarURL:      wechatUser.HeadImgURL,
	}, nil
}
