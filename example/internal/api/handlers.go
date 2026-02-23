// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.xiexianbin.cn/authkit"
	"go.xiexianbin.cn/authkit/providers"
	"go.xiexianbin.cn/authkit/types"
	"golang.org/x/oauth2"

	"example/internal/services"
	"example/utils"
)

// AppConfig defines the application's configuration
type AppConfig struct {
	Alipay    types.OauthConfig
	Apple     types.OauthConfig
	Dingtalk  types.OauthConfig
	Facebook  types.OauthConfig
	Feishu    types.OauthConfig
	Github    types.OauthConfig
	Google    types.OauthConfig
	Microsoft types.OauthConfig
	QQ        types.OauthConfig
	Twitter   types.OauthConfig
	Wechat    types.OauthConfig
}

func InitProviders() {
	var err error
	err = godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found or error loading it, falling back to environment variables")
	}

	config := AppConfig{}
	err = envconfig.Process("", &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Printf("%#v", config)

	if config.Alipay.ClientID != "" {
		authkit.RegisterProvider(types.ALIPAY, providers.NewAlipayProvider(&config.Alipay))
	}
	if config.Apple.ClientID != "" {
		authkit.RegisterProvider(types.APPLE, providers.NewAppleProvider(&config.Apple))
	}
	if config.Dingtalk.ClientID != "" {
		authkit.RegisterProvider(types.DINGTALK, providers.NewDingtalkProvider(&config.Dingtalk))
	}
	if config.Facebook.ClientID != "" {
		authkit.RegisterProvider(types.FACEBOOK, providers.NewFacebookProvider(&config.Facebook))
	}
	if config.Feishu.ClientID != "" {
		authkit.RegisterProvider(types.FEISHU, providers.NewFeishuProvider(&config.Feishu))
	}
	if config.Github.ClientID != "" {
		authkit.RegisterProvider(types.GITHUB, providers.NewGithubProvider(&config.Github))
	}
	if config.Google.ClientID != "" {
		authkit.RegisterProvider(types.GOOGLE, providers.NewGoogleProvider(&config.Google))
	}
	if config.Microsoft.ClientID != "" {
		authkit.RegisterProvider(types.MICROSOFT, providers.NewMicrosoftProvider(&config.Microsoft))
	}
	if config.QQ.ClientID != "" {
		authkit.RegisterProvider(types.QQ, providers.NewQQProvider(&config.QQ))
	}
	if config.Twitter.ClientID != "" {
		authkit.RegisterProvider(types.TWITTER, providers.NewTwitterProvider(&config.Twitter))
	}
	if config.Wechat.ClientID != "" {
		authkit.RegisterProvider(types.WECHAT, providers.NewWechatProvider(&config.Wechat))
	}
}

type AuthHandler struct {
	AuthService *services.AuthService
}

// generateCryptoRandomString generates a secure random string of length 32
func generateCryptoRandomString() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
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
		html += fmt.Sprintf("<a href=\"/api/v1/oauth/%s/login\">Login with %s</a><br/>", name, name)
	}

	html += `</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// HandleOauthLoginRedirect handles login/binding redirection
func (h *AuthHandler) HandleOauthLoginRedirect(c *gin.Context) {
	providerName := c.Param("provider")
	provider, err := authkit.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// CSRF protection: generate a random state and store it in a cookie
	state := generateCryptoRandomString()

	// The secure flag value is usually based on environment config, set to false for now in dev
	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	// PKCE (Proof Key for Code Exchange)
	codeVerifier := oauth2.GenerateVerifier()
	c.SetCookie("oauth_code_verifier", codeVerifier, 3600, "/", "", false, true)

	redirectURL := provider.GetAuthURL(c.Request.Context(), state, oauth2.S256ChallengeOption(codeVerifier))
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// HandleOauthBindRedirect handles bind redirection for logged-in users
func (h *AuthHandler) HandleOauthBindRedirect(c *gin.Context) {
	providerName := c.Param("provider")
	provider, err := authkit.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	state := generateCryptoRandomString()
	// Prefix state with "bind_" to identify the action in the callback
	state = "bind_" + state

	c.SetCookie("oauth_state", state, 3600, "/", "", false, true)

	codeVerifier := oauth2.GenerateVerifier()
	c.SetCookie("oauth_code_verifier", codeVerifier, 3600, "/", "", false, true)

	redirectURL := provider.GetAuthURL(c.Request.Context(), state, oauth2.S256ChallengeOption(codeVerifier))
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// HandleOauthCallback handles the callback
func (h *AuthHandler) HandleOauthCallback(c *gin.Context) {
	providerName := c.Param("provider")

	// CSRF validation
	stateCookie, err := c.Cookie("oauth_state")
	if err != nil || c.Query("state") != stateCookie {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid state token"})
		return
	}

	// Read PKCE verifier from Cookie
	codeVerifierCookie, err := c.Cookie("oauth_code_verifier")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing code_verifier mapping"})
		return
	}

	provider, err := authkit.GetProvider(providerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := c.Query("code")
	// Pass VerifierOption to ExchangeCodeForToken
	token, err := provider.ExchangeCodeForToken(c.Request.Context(), code, oauth2.VerifierOption(codeVerifierCookie))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token: " + err.Error()})
		return
	}

	userInfo, err := provider.GetUserInfo(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	state := c.Query("state")
	isBind := false
	if len(state) > 5 && state[:5] == "bind_" {
		isBind = true
	}

	if isBind {
		// Read JWT token
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" && len(authHeader) > 7 {
			tokenString = authHeader[7:]
		} else {
			tokenStringVal, err := c.Cookie("jwt_token")
			if err == nil {
				tokenString = tokenStringVal
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided for binding"})
			return
		}

		userID, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid token for binding"})
			return
		}

		err = h.AuthService.HandleOauthBind(userID, userInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bind account: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Binding successful",
		})
		return
	}

	user, err := h.AuthService.HandleOauthLoginOrRegister(userInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user: " + err.Error()})
		return
	}

	// Login successful, generate our own JWT
	jwtToken, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	// Optionally store it in cookie for web clients
	c.SetCookie("jwt_token", jwtToken, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   jwtToken,
		"user":    user,
	})
}
