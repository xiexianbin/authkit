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

package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string `gorm:"type:varchar(100);uniqueIndex;not null"`
	Email        string `gorm:"type:varchar(255);uniqueIndex"`
	PasswordHash string `gorm:"type:varchar(255);nullable"`
	Avatar       string `gorm:"type:varchar(255);nullable"`
	// 关系:一个用户可以有多个第三方账号
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

	// 关系: 属于一个用户
	User User `gorm:"foreignKey:UserID"`
}
