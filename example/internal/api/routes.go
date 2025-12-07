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
