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

package database

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"example/internal/models"
)

var DB *gorm.DB

func init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("example.db"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %s", err))
	}

	// 迁移 schema
	DB.AutoMigrate(&models.User{}, &models.OauthAccount{})
}
