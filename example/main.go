package main

import (
	"example/internal/api"
	"example/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)
	engine := gin.Default()
	api.SetupRoutes(engine, database.DB)
	engine.Run("127.0.0.1:8080")
}
