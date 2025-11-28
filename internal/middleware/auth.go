package middleware

import (
	"net/http"
	"strings"

	"taxifleet/backend/internal/permissions"
	"taxifleet/backend/internal/repository"
	"taxifleet/backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Auth(authService *service.AuthService, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is empty"})
			c.Abort()
			return
		}

		user, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		logger.Infof("User authenticated: %s %s (%s)", user.FirstName, user.LastName, permissions.GetRoleName(user.Permission))

		// Store user in context
		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Set("tenantID", user.TenantID)
		c.Set("permission", user.Permission)

		c.Next()
	}
}

func RequirePermission(requiredPermissions ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userObj, ok := user.(*repository.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user object"})
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		hasPermission := false
		for _, perm := range requiredPermissions {
			if permissions.HasPermission(userObj.Permission, perm) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole is kept for backward compatibility, converts role names to permissions
func RequireRole(roles ...string) gin.HandlerFunc {
	perms := make([]int, len(roles))
	for i, role := range roles {
		perms[i] = permissions.GetPermissionForRole(role)
	}
	return RequirePermission(perms...)
}
