// morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/backend/internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"filevault-backend/internal/services"
	"filevault-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		authHeader := c.GetHeader("Authorization")
		
		// Check for token in header first, then in query parameter for downloads
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else if c.Query("token") != "" {
			tokenString = c.Query("token")
		} else {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header or token required")
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token")
			c.Abort()
			return
		}

		// Set user info in context from the token's claims
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("isAdmin", claims.IsAdmin)
		
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}