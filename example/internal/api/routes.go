package api

import (
	"example/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// 依赖注入
	authService := services.NewAuthService(db)
	authHandler := &AuthHandler{AuthService: authService}

	router.GET("/", authHandler.Home)

	v1 := router.Group("/api/v1")
	{
		oauthGroup := v1.Group("/oauth/:provider")
		{
			// 对应登录/注册流程的第1步
			oauthGroup.GET("/login", authHandler.HandleOauthLoginRedirect)

			// 对应绑定流程的第1步 (需要JWT认证中间件)
			// oauthGroup.GET("/bind", middleware.JWTAuth(), authHandler.HandleOauthLoginRedirect)

			// 统一的回调处理
			oauthGroup.GET("/callback", authHandler.HandleOauthCallback)
		}
	}
}
