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
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"golang.org/x/oauth2"
)

// Twitter's OAuth 2.0 uses PKCE for security. This makes the flow more
// complex as we need to generate and later verify a `code_verifier`.
// **A server-side session or cache (like Redis) is required to temporarily
// store the `code_verifier`**, associating it with the `state` parameter.
// this code the verifier but won't implement the session logic. In a real
// app, you would store `codeVerifier` in a session keyed by `state` in
// `GetAuthURL`, and retrieve it in `ExchangeCodeForToken`.

// Simple in-memory store for PKCE verifiers.
// **WARNING**: In production, use a proper distributed cache like Redis.
var pkceVerifierStore = make(map[string]string)

type TwitterProvider struct {
	Name   string
	config *oauth2.Config
}

func NewTwitterProvider(cfg *OauthConfig) Provider {
	return &TwitterProvider{
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

func (p *TwitterProvider) GetAuthURL(state string) string {
	// Generate PKCE parameters
	// In a real app, you should use a more robust random string generator
	codeVerifier := "a_very_random_and_long_string_for_pkce"
	pkceVerifierStore[state] = codeVerifier // Store verifier associated with state

	hasher := sha256.New()
	hasher.Write([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	return p.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

func (p *TwitterProvider) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	// We need the state to retrieve the code_verifier, but the interface doesn't pass it.
	// This highlights a limitation. A real implementation would need access to the request context
	// to get the state. We'll assume we can retrieve it for this example.
	// In your handler: state, _ := c.Cookie("oauth_state")
	state := "random_state_string" // Placeholder - MUST be retrieved from the actual request
	codeVerifier := pkceVerifierStore[state]
	if codeVerifier == "" {
		return nil, fmt.Errorf("code verifier not found for state, session might have expired")
	}
	delete(pkceVerifierStore, state) // Clean up

	return p.config.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
}

func (p *TwitterProvider) GetUserInfo(token *oauth2.Token) (*UserInfo, error) {
	client := p.config.Client(context.Background(), token)
	// Twitter API v2 requires specifying fields
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

	return &UserInfo{
		Provider:       "twitter",
		ProviderUserID: twitterResp.Data.ID,
		Name:           twitterResp.Data.Name,
		// Twitter OAuth 2.0 does not provide email by default
		Email:     "",
		AvatarURL: twitterResp.Data.ProfileImageURL,
	}, nil
}
