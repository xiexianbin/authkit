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

	"go.xiexianbin.cn/authkit/providers"
	"go.xiexianbin.cn/authkit/types"
)

var registry = make(map[string]types.Provider)

// InitFactory init all OAuth Provider
func InitFactory(cfg *types.Config) {
	if cfg.Alipay.ClientID != "" {
		registry[types.ALIPAY] = providers.NewAlipayProvider(&cfg.Alipay)
	}
	if cfg.Apple.ClientID != "" {
		registry[types.APPLE] = providers.NewAppleProvider(&cfg.Apple)
	}
	if cfg.Dingtalk.ClientID != "" {
		registry[types.DINGTALK] = providers.NewDingtalkProvider(&cfg.Dingtalk)
	}
	if cfg.Facebook.ClientID != "" {
		registry[types.FACEBOOK] = providers.NewFacebookProvider(&cfg.Facebook)
	}
	if cfg.Feishu.ClientID != "" {
		registry[types.FEISHU] = providers.NewFeishuProvider(&cfg.Feishu)
	}
	if cfg.Github.ClientID != "" {
		registry[types.GITHUB] = providers.NewGithubProvider(&cfg.Github)
	}
	if cfg.Google.ClientID != "" {
		registry[types.GOOGLE] = providers.NewGoogleProvider(&cfg.Google)
	}
	if cfg.Microsoft.ClientID != "" {
		registry[types.MICROSOFT] = providers.NewMicrosoftProvider(&cfg.Microsoft)
	}
	if cfg.QQ.ClientID != "" {
		registry[types.QQ] = providers.NewQQProvider(&cfg.QQ)
	}
	if cfg.Twitter.ClientID != "" {
		registry[types.TWITTER] = providers.NewTwitterProvider(&cfg.Twitter)
	}
	if cfg.Wechat.ClientID != "" {
		registry[types.WECHAT] = providers.NewWechatProvider(&cfg.Wechat)
	}
}

// GetProvider get an OAuth Provider instance by name
func GetProvider(name string) (types.Provider, error) {
	provider, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", name)
	}
	return provider, nil
}

// GetProviders returns a list of registered provider names
func GetProviders() []string {
	var names []string
	for k := range registry {
		names = append(names, k)
	}
	return names
}
