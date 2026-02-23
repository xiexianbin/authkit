package main

import (
	"example/internal/api"
	"example/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	api.InitProviders() // Initialize OAuth providers based on environment config

	gin.SetMode(gin.DebugMode)
	engine := gin.Default()
	api.SetupRoutes(engine, database.DB)
	engine.Run("127.0.0.1:8080")
}
