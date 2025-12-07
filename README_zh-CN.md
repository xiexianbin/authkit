# authkit

[![Go Report Card](https://goreportcard.com/badge/go.xiexianbin.cn/authkit)](https://goreportcard.com/report/go.xiexianbin.cn/authkit)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/license/apache-2-0)
[![Go.Dev reference](https://pkg.go.dev/badge/go.xiexianbin.cn/authkit?utm_source=godoc)](https://pkg.go.dev/go.xiexianbin.cn/authkit)

[English Documentation](README.md)

一个支持多种第三方 OAuth（如 GitHub, Google）登录和获取用户信息的 golang 实现。

## 特性

- **简化 OAuth2/OpenID Connect**: 为各种提供商提供统一的接口。
- **可扩展**: 易于添加新的提供商。
- **标准化用户信息**: 跨提供商规范化用户信息结构。

## 安装

```bash
go get go.xiexianbin.cn/authkit
```

## 使用方法

这是一个使用 Gin 的简单示例：

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.xiexianbin.cn/authkit"
	"go.xiexianbin.cn/authkit/types"
)

func init() {
	// 使用环境变量或配置文件初始化配置
	config := &types.Config{
		// ... 加载 headers mapstructure 等配置
	}
	authkit.InitFactory(config)
}

func main() {
	r := gin.Default()

	// 重定向到提供商登录页面
	r.GET("/oauth/:provider/login", func(c *gin.Context) {
		providerName := c.Param("provider")
		provider, err := authkit.GetProvider(providerName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 在生产环境中，应使用随机 state 进行 CSRF 保护
		state := "random_state_string"
		redirectURL := provider.GetAuthURL(state)
		c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	})

	// 处理提供商回调
	r.GET("/oauth/:provider/callback", func(c *gin.Context) {
		providerName := c.Param("provider")
		provider, err := authkit.GetProvider(providerName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		code := c.Query("code")
		// 用 code 换取 token
		token, err := provider.ExchangeCodeForToken(c.Request.Context(), code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		//获取用户信息
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

## 支持的提供商

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
- [WeChat](https://open.weixin.qq.com/) ([参考](https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/login.html))

## 许可证

本项目采用 Apache 2.0 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。
