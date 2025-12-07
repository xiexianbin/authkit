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

func (p *AlipayProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
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
	}, nil
}
