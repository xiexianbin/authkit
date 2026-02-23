// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package api

import (
	"example/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(router *gin.Engine, db *gorm.DB) {
	// Dependency injection
	authService := services.NewAuthService(db)
	authHandler := &AuthHandler{AuthService: authService}

	router.GET("/", authHandler.Home)

	v1 := router.Group("/api/v1")
	{
		oauthGroup := v1.Group("/oauth/:provider")
		{
			// Corresponds to Step 1 of the login/registration process
			oauthGroup.GET("/login", authHandler.HandleOauthLoginRedirect)

			// Corresponds to Step 1 of the binding process (requires JWT auth middleware)
			oauthGroup.GET("/bind", JWTAuth(), authHandler.HandleOauthBindRedirect)

			// Unified callback handling
			oauthGroup.GET("/callback", authHandler.HandleOauthCallback)
		}
	}
}
