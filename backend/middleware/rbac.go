package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RoleRequired(requiredRole string) gin.HandlerFunc {

	return func(c *gin.Context) {

		roleRaw, exists := c.Get("role")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}

		role := roleRaw.(string)

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}