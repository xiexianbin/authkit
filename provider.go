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

import "golang.org/x/oauth2"

// UserInfo defines a standardized structure of user information
// obtained from any OAuth Provider
type UserInfo struct {
	Provider       string
	ProviderUserID string
	Email          string
	Name           string
	AvatarURL      string
	RawData        any
}

// Provider is a mandatory interface for all OAuth implementations
type Provider interface {
	GetAuthURL(state string) string
	ExchangeCodeForToken(code string) (*oauth2.Token, error)
	GetUserInfo(token *oauth2.Token) (*UserInfo, error)
}
