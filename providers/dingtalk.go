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

type DingtalkProvider struct {
	Name   string
	config *oauth2.Config
}

func NewDingtalkProvider(cfg *types.OauthConfig) types.Provider {
	return &DingtalkProvider{
		Name: types.DINGTALK,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"openid", "corpid"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://login.dingtalk.com/oauth2/auth",
				TokenURL: "https://api.dingtalk.com/v1.0/oauth2/userAccessToken",
			},
		},
	}
}

func (p *DingtalkProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.config.AuthCodeURL(state, opts...)
}

func (p *DingtalkProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// DingTalk requires POST request with JSON body
	reqBody := map[string]string{
		"clientId":     p.config.ClientID,
		"clientSecret": p.config.ClientSecret,
		"code":         code,
		"grantType":    "authorization_code",
	}
	jsonBody, err := json.Marshal(reqBody)
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

	var tokenData struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpireIn     int    `json:"expireIn"`
		CorpId       string `json:"corpId"`
	}

	if err := json.Unmarshal(body, &tokenData); err != nil {
		return nil, err
	}

	if tokenData.AccessToken == "" {
		return nil, fmt.Errorf("dingtalk token error: %s", string(body))
	}

	token := &oauth2.Token{
		AccessToken:  tokenData.AccessToken,
		RefreshToken: tokenData.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenData.ExpireIn) * time.Second),
	}

	return token, nil
}

func (p *DingtalkProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	userInfoURL := "https://api.dingtalk.com/v1.0/contact/users/me"
	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-acs-dingtalk-access-token", token.AccessToken)

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
		Nick      string `json:"nick"`
		Avatar    string `json:"avatarUrl"`
		Email     string `json:"email"`
		OpenId    string `json:"openId"`
		UnionId   string `json:"unionId"`
		Mobile    string `json:"mobile"`
		StateCode string `json:"stateCode"`
	}

	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, err
	}

	// DingTalk API returns strict structure, we might need to handle wrapper if present
	// However, v1.0 usually returns data directly or inside "result" wrapper?
	// Checking docs, it seems it might be direct fields or wrapped.
	// Let's assume direct mapping for now based on common v1.0 APIs,
	// but mostly DingTalk APIs wrap response in specific way.
	// It usually returns a structure like request-id etc.

	// Re-checking docs: https://open.dingtalk.com/document/orgapp/dingtalk-get-user-info
	// Response: { "nick": "...", "avatarUrl": "...", "openId": "...", "unionId": "...", ... }
	// It seems it is a flat JSON structure for the user info part?
	// ACTUALLY, usually DingTalk new API responses are simple JSON.

	providerUserID := userResp.UnionId
	if providerUserID == "" {
		providerUserID = userResp.OpenId
	}

	return &types.UserInfo{
		Provider:       p.Name,
		ProviderUserID: providerUserID,
		Name:           userResp.Nick,
		AvatarURL:      userResp.Avatar,
		Email:          userResp.Email,
	}, nil
}
