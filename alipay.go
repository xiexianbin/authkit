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
	"fmt"
	"net/url"

	"golang.org/x/oauth2"
)

// 支付宝的 OAuth 流程比较特殊，它严重依赖其 SDK 进行签名。为了简化并保持接口一致，我们这里只演示流程，并**强烈建议在生产环境中使用支付宝官方 SDK** 来处理签名和 API 调用。这里的实现是一个示意性的简化版。
//
// **关键点**:
// * `scope` 为 `auth_user`。
// * 获取用户信息需要 `alipay.user.info.share` 接口。
// * 实际 API 调用需要复杂的签名过程。

type AlipayProvider struct {
	Name   string
	config *oauth2.Config
	// 实际项目中需要 alipay.Client
}

// NewAlipayProvider 创建一个新的支付宝Provider实例
func NewAlipayProvider(cfg *OauthConfig) Provider {
	return &AlipayProvider{
		Name: ALIPAY,
		config: &oauth2.Config{
			ClientID: cfg.ClientID,
			// ClientSecret 在支付宝 OAuth 中不直接使用
			RedirectURL: cfg.RedirectURL,
			Scopes:      []string{"auth_user"},
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm",
				// TokenURL 的调用很特殊，标准库的 Exchange 无法直接使用
				TokenURL: "https://openapi.alipay.com/gateway.do",
			},
		},
	}
}

func (p *AlipayProvider) GetAuthURL(state string) string {
	// 支付宝的授权 URL 参数是 app_id
	return fmt.Sprintf(
		"%s?app_id=%s&scope=%s&redirect_uri=%s&state=%s",
		p.config.Endpoint.AuthURL,
		p.config.ClientID,
		p.config.Scopes[0],
		url.QueryEscape(p.config.RedirectURL),
		state,
	)
}

// ExchangeCodeForToken **简化版，仅作示意**。
// 生产环境必须使用支付宝 SDK 进行签名请求。
func (p *AlipayProvider) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	// 伪代码:
	// 1. 创建 alipay.Client 实例 (使用 AppID, PrivateKey, AlipayPublicKey)
	// 2. 构造 alipay.SystemOauthTokenRequest
	// 3. 设置 GrantType="authorization_code", Code=code
	// 4. 调用 client.Execute(request)
	// 5. 从响应中解析 access_token, user_id, refresh_token

	// 此处返回一个模拟的 token，实际无法工作
	fmt.Println("警告: 支付宝 ExchangeCodeForToken 是一个简化实现，生产环境必须使用官方SDK。")
	return &oauth2.Token{
		AccessToken: "DUMMY_ALIPAY_ACCESS_TOKEN", // 假设从 SDK 获取
	}, nil
}

// GetUserInfo **简化版，仅作示意**。
func (p *AlipayProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	// 伪代码:
	// 1. 创建 alipay.Client 实例
	// 2. 构造 alipay.UserInfoShareRequest
	// 3. 调用 client.Execute(request, WithAuthToken(token.AccessToken))
	// 4. 从响应中解析 user_id, nick_name, avatar 等

	fmt.Println("警告: 支付宝 GetUserInfo 是一个简化实现，生产环境必须使用官方SDK。")
	userID := token.Extra("user_id").(string)

	return &UserInfo{
		Provider:       p.Name,
		ProviderUserID: userID,
		Name:           "支付宝用户", // 假设从 SDK 获取
		AvatarURL:      "",      // 假设从 SDK 获取
		Email:          "",      // 支付宝不提供邮箱
	}, nil
}
