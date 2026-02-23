// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package types

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
	// - Alipay might require extra field: `AppPrivateKey`
	Extra map[string]any
}
