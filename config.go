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

// OauthConfig defines the universal configuration for an OAuth provider.
// envconfig is tag for [https://github.com/kelseyhightower/envconfig]
type OauthConfig struct {
	ClientID     string `envconfig:"CLIENT_ID"`
	ClientSecret string `envconfig:"CLIENT_SECRET"`
	RedirectURL  string `envconfig:"REDIRECT_URL"`

	// Extra oauth config
	//
	// - Apple-specific fields: `TeamID` `KeyID` and `AppPrivateKey`(The content of your .p8 private key file for Apple)
	//
	// - 支付宝可能需要额外字段： `AppPrivateKey`
	Extra map[string]any
}

// Config is the main application configuration.
type Config struct {
	Alipay    OauthConfig
	Apple     OauthConfig
	Facebook  OauthConfig
	Github    OauthConfig
	Google    OauthConfig
	Microsoft OauthConfig
	QQ        OauthConfig
	Twitter   OauthConfig
	Wechat    OauthConfig
}
