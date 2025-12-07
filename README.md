# authkit

[![Go Report Card](https://goreportcard.com/badge/go.xiexianbin.cn/authkit)](https://goreportcard.com/report/go.xiexianbin.cn/authkit)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/license/apache-2-0)
[![Go.Dev reference](https://pkg.go.dev/badge/go.xiexianbin.cn/authkit?utm_source=godoc)](https://pkg.go.dev/go.xiexianbin.cn/authkit)

[中文文档](README_zh-CN.md)

A Golang implementation that supports multiple third-party OAuth (e.g., GitHub, Google) logins and obtains user information.

## Features

- **Simplifies OAuth2/OpenID Connect**: Provides a unified interface for various providers.
- **Extensible**: Easy to add new providers.
- **Standardized User Info**: normalized user information structure across providers.

## Installation

```bash
go get go.xiexianbin.cn/authkit
```

## Usage

Here is a simple example using Gin:

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.xiexianbin.cn/authkit"
	"go.xiexianbin.cn/authkit/types"
)

func init() {
	// Initialize with configuration based on environment variables or config file
	config := &types.Config{
		// ... load headers mapstructure or similar
	}
	authkit.InitFactory(config)
}

func main() {
	r := gin.Default()

	// Redirect to provider login page
	r.GET("/oauth/:provider/login", func(c *gin.Context) {
		providerName := c.Param("provider")
		provider, err := authkit.GetProvider(providerName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// In production, use a random state for CSRF protection
		state := "random_state_string"
		redirectURL := provider.GetAuthURL(state)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	})

	// Handle callback from provider
	r.GET("/oauth/:provider/callback", func(c *gin.Context) {
		providerName := c.Param("provider")
		provider, err := authkit.GetProvider(providerName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		code := c.Query("code")
		// Exchange code for token
		token, err := provider.ExchangeCodeForToken(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get user info
		userInfo, err := provider.GetUserInfo(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, userInfo)
	})

	r.Run(":8080")
}
```

## Supported Providers

- Alipay
- Apple ID
- DingTalk
- Facebook
- Feishu / Lark
- [Github](https://github.com/settings/developers)
- [Google](https://console.cloud.google.com/auth/clients/create)
- Microsoft Account
- [QQ](https://connect.qq.com/)
- Twitter (X)
- [WeChat](https://open.weixin.qq.com/) ([ref](https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/login.html))

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
