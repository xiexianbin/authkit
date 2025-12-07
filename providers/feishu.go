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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"go.xiexianbin.cn/authkit/types"
)

type FeishuProvider struct {
	Name   string
	config *oauth2.Config
}

func NewFeishuProvider(cfg *types.OauthConfig) types.Provider {
	return &FeishuProvider{
		Name: types.FEISHU,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{}, // Feishu scopes are configured in the app console, usually not needed here or just empty
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://passport.feishu.cn/suite/passport/oauth/authorize",
				TokenURL: "https://passport.feishu.cn/suite/passport/oauth/token",
			},
		},
	}
}

func (p *FeishuProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *FeishuProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// Feishu might require app_access_token separately effectively?
	// Or standard OAuth flow works?
	// Docs say POST to https://passport.feishu.cn/suite/passport/oauth/token
	// form-data or json? usually x-www-form-urlencoded compatible or json.

	values := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     p.config.ClientID,
		"client_secret": p.config.ClientSecret,
		"code":          code,
		"redirect_uri":  p.config.RedirectURL,
	}

	jsonBody, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint.TokenURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("feishu token error: %s", string(body))
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return token, nil
}

func (p *FeishuProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	userInfoURL := "https://passport.feishu.cn/suite/passport/oauth/userinfo"
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userResp struct {
		Name         string `json:"name"`
		EnName       string `json:"en_name"`
		AvatarUrl    string `json:"avatar_url"`
		AvatarThumb  string `json:"avatar_thumb"`
		AvatarMiddle string `json:"avatar_middle"`
		AvatarBig    string `json:"avatar_big"`
		Email        string `json:"email"`
		UserId       string `json:"user_id"`
		Mobile       string `json:"mobile"`
		UnionId      string `json:"union_id"`
		OpenId       string `json:"open_id"`
	}

	// Docs say it returns a JSON object with user info.
	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, err
	}

	if userResp.OpenId == "" && userResp.UnionId == "" {
		// Possibly wrapped in "data" or error?
		// "code": 0, "msg": "success", "data": {...} ?
		// Wait, some older docs say one thing, new docs say another.
		// Trying to handle common wrapper if initial unmarshal fails to find ID.
		var wrapper struct {
			Code int             `json:"code"`
			Msg  string          `json:"msg"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(body, &wrapper); err == nil && wrapper.Code != 0 {
			return nil, fmt.Errorf("feishu error: %s", wrapper.Msg)
		}
	}

	// Prioritize UnionID
	providerUserID := userResp.UnionId
	if providerUserID == "" {
		providerUserID = userResp.OpenId
	}

	return &types.UserInfo{
		Provider:       p.Name,
		ProviderUserID: providerUserID,
		Name:           userResp.Name,
		AvatarURL:      userResp.AvatarUrl,
		Email:          userResp.Email,
	}, nil
}
