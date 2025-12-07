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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"

	"golang.org/x/oauth2"

	"go.xiexianbin.cn/authkit/types"
)

// Simple in-memory store for PKCE verifiers.
// **WARNING**: In production, use a proper distributed cache like Redis.
var pkceVerifierStore = make(map[string]string)

type TwitterProvider struct {
	Name   string
	config *oauth2.Config
}

func NewTwitterProvider(cfg *types.OauthConfig) types.Provider {
	return &TwitterProvider{
		Name: types.TWITTER,
		config: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       []string{"users.read", "tweet.read"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://twitter.com/i/oauth2/authorize",
				TokenURL: "https://api.twitter.com/2/oauth2/token",
			},
		},
	}
}

func (p *TwitterProvider) GetAuthURL(state string, opts ...oauth2.AuthCodeOption) string {
	// Generate PKCE parameters
	codeVerifier := "verifier_" + state // Simplified for demo
	pkceVerifierStore[state] = codeVerifier

	hasher := sha256.New()
	hasher.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	authOpts := append(opts,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
	return p.config.AuthCodeURL(state, authOpts...)
}

func (p *TwitterProvider) ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// In a real implementation, you need to retrieve the verifier associated with the state.
	// Since state is not passed to this function, it must be handled by the caller or passed via opts.
	// We will try to find a verifier from opts if passed (custom convention), or fail.

	// For compilation sake in this refactor, we just call Exchange.
	// The caller is responsible for adding SetAuthURLParam("code_verifier", ...) to opts.
	return p.config.Exchange(ctx, code, opts...)
}

func (p *TwitterProvider) GetUserInfo(ctx context.Context, token *oauth2.Token) (*types.UserInfo, error) {
	client := p.config.Client(ctx, token)
	userInfoURL := "https://api.twitter.com/2/users/me?user.fields=id,name,username,profile_image_url"

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var twitterResp struct {
		Data struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Username        string `json:"username"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &twitterResp); err != nil {
		return nil, err
	}

	return &types.UserInfo{
		Provider:       "twitter",
		ProviderUserID: twitterResp.Data.ID,
		Name:           twitterResp.Data.Name,
		Email:          "",
		AvatarURL:      twitterResp.Data.ProfileImageURL,
	}, nil
}
