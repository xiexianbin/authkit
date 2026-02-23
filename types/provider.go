// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package types

import (
	"context"

	"golang.org/x/oauth2"
)

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
	GetAuthURL(ctx context.Context, state string, opts ...oauth2.AuthCodeOption) string
	ExchangeCodeForToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error)
}
