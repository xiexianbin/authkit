// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package services

import (
	"fmt"

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

// HandleOauthLoginOrRegister handles the core logic after OAuth callback
func (s *AuthService) HandleOauthLoginOrRegister(userInfo *types.UserInfo) (*models.User, error) {
	var oauthAccount models.OauthAccount

	// 1. Check if the third-party account already exists
	err := s.DB.Preload("User").Where("provider = ? AND provider_user_id = ?", userInfo.Provider, userInfo.ProviderUserID).First(&oauthAccount).Error
	if err == nil {
		// Exists, return the associated local user directly
		return &oauthAccount.User, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err // Other database errors
	}

	// Transaction processing to ensure data consistency
	tx := s.DB.Begin()

	// 2. Third-party account does not exist, check if email is registered
	var user models.User
	var emailErr error

	if userInfo.Email != "" {
		emailErr = tx.Where("email = ?", userInfo.Email).First(&user).Error
	} else {
		// If email is empty, we force the flow to create a brand new user
		emailErr = gorm.ErrRecordNotFound
	}

	if emailErr == nil {
		// Email exists, bind a new OAuth account for the user
		// (Create oauthAccount record here and associate it with user.ID)
	} else if emailErr == gorm.ErrRecordNotFound {
		// 3. Brand new user, create local user and OAuth account

		// Ensure we have a valid username fallback if empty
		username := userInfo.Name
		if username == "" {
			username = "oauthUser_" + userInfo.ProviderUserID[:min(8, len(userInfo.ProviderUserID))]
		}

		user = models.User{
			Username: username,
			Email:    userInfo.Email,
			Avatar:   userInfo.AvatarURL,
		}

		// Ensure username is unique, though practically a better approach is to append a random string if error
		// Simple fallback:
		err = tx.Create(&user).Error
		if err != nil {
			// If unique constraint fails, try appending the provider userId
			user.Username = username + "_" + userInfo.ProviderUserID[:min(4, len(userInfo.ProviderUserID))]
			if errRetry := tx.Create(&user).Error; errRetry != nil {
				tx.Rollback()
				return nil, errRetry
			}
		}
	} else {
		tx.Rollback()
		return nil, emailErr // Other database errors
	}

	// Create a new oauthAccount for the user
	newOauthAccount := models.OauthAccount{
		UserID:         user.ID,
		Provider:       userInfo.Provider,
		ProviderUserID: userInfo.ProviderUserID,
		// Optional: store token
	}

	if err := tx.Create(&newOauthAccount).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Return the newly created or found user info
	user.OauthAccounts = append(user.OauthAccounts, newOauthAccount)
	return &user, nil
}

// HandleOauthBind specifically handles binding a new OAuth info to an existing User
func (s *AuthService) HandleOauthBind(userID uint, userInfo *types.UserInfo) error {
	var oauthAccount models.OauthAccount

	// Check if this specific oauth provider record already exists globally
	err := s.DB.Where("provider = ? AND provider_user_id = ?", userInfo.Provider, userInfo.ProviderUserID).First(&oauthAccount).Error
	if err == nil {
		if oauthAccount.UserID == userID {
			// Already bound to themselves, valid no-op
			return nil
		}
		// Bound to someone else
		return gorm.ErrInvalidData
	}

	if err != gorm.ErrRecordNotFound {
		return err // Other database errors
	}

	// Important security check: Does the user already have this provider bound?
	// E.g., a user shouldn't bind 2 github accounts
	var existingBinding models.OauthAccount
	err = s.DB.Where("user_id = ? AND provider = ?", userID, userInfo.Provider).First(&existingBinding).Error
	if err == nil {
		// User already bounds an account from this provider
		return fmt.Errorf("user already has a %s account bound", userInfo.Provider)
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	newOauthAccount := models.OauthAccount{
		UserID:         userID,
		Provider:       userInfo.Provider,
		ProviderUserID: userInfo.ProviderUserID,
	}

	return s.DB.Create(&newOauthAccount).Error
}
