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
