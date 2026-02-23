// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: hi@xiexianbin.cn

package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Email        string `gorm:"type:varchar(255);uniqueIndex"`
	PasswordHash string `gorm:"type:varchar(255);nullable"`
	Avatar       string `gorm:"type:varchar(255);nullable"`
	// Relationship: One user can have multiple third-party accounts
	OauthAccounts []OauthAccount `gorm:"foreignKey:UserID"`
}

type OauthAccount struct {
	gorm.Model
	UserID         uint   `gorm:"index;not null"`
	Provider       string `gorm:"type:varchar(50);not null"`
	ProviderUserID string `gorm:"type:varchar(255);not null"`
	AccessToken    string `gorm:"type:text"`
	RefreshToken   string `gorm:"type:text"`
	ExpiresAt      *gorm.DeletedAt

	// Relationship: Belongs to one user
	User User `gorm:"foreignKey:UserID"`
}
