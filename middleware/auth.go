package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gitlab.com/nodiviti/user-service/config"
	"gitlab.com/nodiviti/user-service/utils"
)

// AuthMiddleware validates JWT token with auth service
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	authClient := utils.NewAuthClient(cfg)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// Validate token with auth service
		authResp, err := authClient.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if !authResp.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Token validation failed",
			})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_id", authResp.User.ID)
		c.Set("username", authResp.User.Username)
		c.Set("email", authResp.User.Email)
		c.Set("role", authResp.User.Role)

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
			})
			c.Abort()
			return
		}

		userRole := role.(string)

		// Check if user role is in allowed roles
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

// AdminOnly middleware - only allows admin users
func AdminOnly() gin.HandlerFunc {
	return RoleMiddleware("admin")
}

// TeacherOnly middleware - allows admin and teacher users
func TeacherOnly() gin.HandlerFunc {
	return RoleMiddleware("admin", "teacher")
}

// StudentAccess middleware - allows all authenticated users
func StudentAccess() gin.HandlerFunc {
	return RoleMiddleware("admin", "teacher", "student")
}
