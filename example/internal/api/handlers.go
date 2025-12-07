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

package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.xiexianbin.cn/authkit"
	"go.xiexianbin.cn/authkit/types"

	"example/internal/services"
	"example/utils"
)

func init() {
	var err error
	err = godotenv.Load("/Users/xiexianbin/workspace/code/github.com/xiexianbin/authkit/example/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := types.Config{}
	err = envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("%#v", config)
	authkit.InitFactory(&config)
}

type AuthHandler struct {
	AuthService *services.AuthService
}

func (h *AuthHandler) Home(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
	<title>AuthKit Example</title>
</head>
<body>
	<h1>Welcome to AuthKit Example</h1>`

	for _, name := range authkit.GetProviders() {
		html += fmt.Sprintf("<a href=\"http://127.0.0.1:8080/api/v1/oauth/%s/login\">Login with %s</a><br/>", name, name)
	}

	html += `</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// HandleOauthLoginRedirect 处理登录/绑定跳转
func (h *AuthHandler) HandleOauthLoginRedirect(c *gin.Context) {
	providerName := c.Param("provider")
	provider, err := authkit.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// CSRF保护: 生成 state 并存入 cookie/session
	state := "random_state_string" // 实际应使用随机生成器
	c.SetCookie("oauth_state", state, 3600, "/", "localhost", false, true)

	redirectURL := provider.GetAuthURL(state)
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// HandleOauthCallback 处理回调
func (h *AuthHandler) HandleOauthCallback(c *gin.Context) {
	providerName := c.Param("provider")

	// CSRF 校验
	// log.Printf("%#v", c.Request.Cookies())
	// state, _ := c.Cookie("oauth_state")
	// if c.Query("state") != state {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid state token"})
	// 	return
	// }

	provider, err := authkit.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := c.Query("code")
	token, err := provider.ExchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	userInfo, err := provider.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	// 检查是登录还是绑定流程
	// (可以通过 state 参数、不同的回调 URL 或者检查是否存在 JWT 来区分)
	// 这里简化为统一处理登录/注册

	user, err := h.AuthService.HandleOauthLoginOrRegister(userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user: " + err.Error()})
		return
	}

	// 登录成功，生成我们自己的 JWT
	jwtToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   jwtToken,
		"user":    user,
	})
}
