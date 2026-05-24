package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminOnly restricts access to users with the "admin" role.
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "admin access required",
			})
			return
		}
		c.Next()
	}
}
