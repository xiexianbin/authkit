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
)

var (
	ALIPAY    = "alipay"
	APPLE     = "apple"
	FACEBOOK  = "facebook"
	GITHUB    = "github"
	GOOGLE    = "google"
	MICROSOFT = "microsoft"
	QQ        = "qq"
	TWITTER   = "twitter"
	WECHAT    = "wechat"
)

var providers = make(map[string]Provider)

// InitFactory init all OAuth Provider
func InitFactory(cfg *Config) {
	if cfg.Alipay.ClientID != "" {
		providers[ALIPAY] = NewAlipayProvider(&cfg.Alipay)
	}
	if cfg.Apple.ClientID != "" {
		providers[APPLE] = NewAppleProvider(&cfg.Apple)
	}
	if cfg.Facebook.ClientID != "" {
		providers[FACEBOOK] = NewFacebookProvider(&cfg.Facebook)
	}
	if cfg.Github.ClientID != "" {
		providers[GITHUB] = NewGithubProvider(&cfg.Github)
	}
	if cfg.Google.ClientID != "" {
		providers[GOOGLE] = NewGoogleProvider(&cfg.Google)
	}
	if cfg.Microsoft.ClientID != "" {
		providers[MICROSOFT] = NewMicrosoftProvider(&cfg.Microsoft)
	}
	if cfg.QQ.ClientID != "" {
		providers[QQ] = NewQQProvider(&cfg.QQ)
	}
	if cfg.Twitter.ClientID != "" {
		providers[TWITTER] = NewTwitterProvider(&cfg.Twitter)
	}
	if cfg.Wechat.ClientID != "" {
		providers[WECHAT] = NewWechatProvider(&cfg.Wechat)
	}
}

// GetProvider get an OAuth Provider instance by name
func GetProvider(name string) (Provider, error) {
	provider, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", name)
	}
	return provider, nil
}
