// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package providers

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"

	"go.xiexianbin.cn/authkit/types"
)

type AlipayProvider struct {
	Name   string
	config *oauth2.Config
}

func NewAlipayProvider(cfg *types.OauthConfig) types.Provider {
	return &AlipayProvider{
		Name: types.ALIPAY,
		config: &oauth2.Config{
			ClientID:    cfg.ClientID,
			RedirectURL: cfg.RedirectURL,
			Scopes:      []string{"auth_user"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm",
				TokenURL: "https://openapi.alipay.com/gateway.do",
			},
		},
	}
}

func (p *AlipayProvider) GetAuthURL(ctx context.Context, state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *AlipayProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// Simplified implementation as per original code
	fmt.Println("Warning: Alipay ExchangeCodeForToken is a dummy implementation.")
	return &oauth2.Token{AccessToken: code}, nil
}

func (p *AlipayProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	fmt.Println("Warning: Alipay GetUserInfo is a dummy implementation.")
	// Original code assumed "user_id" in extra, which comes from real exchange.
	// We'll return dummy data.
	return &types.UserInfo{
		Provider:       p.Name,
		ProviderUserID: "dummy_alipay_user_id",
		Name:           "Alipay User",
		RawData:        nil,
	}, nil
}
