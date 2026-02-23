package api

import (
	"net/http"
	"strings"

	"example/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth middleware validates the JWT token
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Attempt to read from Authorization header first
		authHeader := c.GetHeader("Authorization")
		tokenString := ""

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// If no header, attempt to read from cookie or generic query param?
		if tokenString == "" {
			tokenStringVal, err := c.Cookie("jwt_token")
			if err == nil {
				tokenString = tokenStringVal
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
			c.Abort()
			return
		}

		userID, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid token"})
			c.Abort()
			return
		}

		// Store the user ID in the context for downstream handlers to use
		c.Set("userID", userID)
		c.Next()
	}
}
