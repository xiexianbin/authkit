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

package services

import (
	"go.xiexianbin.cn/authkit/types"
	"gorm.io/gorm"

	"example/internal/models"
)

type AuthService struct {
	DB *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{DB: db}
}

// HandleOauthLoginOrRegister 处理 OAuth 回调后的核心逻辑
func (s *AuthService) HandleOauthLoginOrRegister(userInfo *types.UserInfo) (*models.User, error) {
	var oauthAccount models.OauthAccount

	// 1. 检查第三方账号是否已存在
	err := s.DB.Preload("User").Where("provider = ? AND provider_user_id = ?", userInfo.Provider, userInfo.ProviderUserID).First(&oauthAccount).Error
	if err == nil {
		// 已存在，直接返回关联的本地用户
		return &oauthAccount.User, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err // 其他数据库错误
	}

	// 事务处理，确保数据一致性
	tx := s.DB.Begin()

	// 2. 第三方账号不存在，检查 email 是否已注册
	var user models.User
	err = tx.Where("email = ?", userInfo.Email).First(&user).Error

	if err == nil {
		// Email 已存在，为该用户绑定新的 Oauth 账号
		// (在这里创建 oauthAccount 记录并关联到 user.ID)
	} else if err == gorm.ErrRecordNotFound {
		// 3. 全新用户，创建本地用户和 Oauth 账号
		user = models.User{
			Username: userInfo.Name, // Or generate a unique one
			Email:    userInfo.Email,
			Avatar:   userInfo.AvatarURL,
		}
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		tx.Rollback()
		return nil, err // 其他数据库错误
	}

	// 为 user 创建新的 oauthAccount
	newOauthAccount := models.OauthAccount{
		UserID:         user.ID,
		Provider:       userInfo.Provider,
		ProviderUserID: userInfo.ProviderUserID,
		// 可选：存储 token
	}

	if err := tx.Create(&newOauthAccount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 返回新创建或找到的用户信息
	user.OauthAccounts = append(user.OauthAccounts, newOauthAccount)
	return &user, nil
}
