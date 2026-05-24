package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	infrajwt "github.com/unillm/unillm/infra/jwt"
)

// JWTAuth validates JWT tokens from Authorization header for dashboard APIs.
func JWTAuth(issuer *infrajwt.Issuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := issuer.Parse(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// APIKeyAuth validates API keys (sk-xxx) for proxy endpoints.
func APIKeyAuth(resolver func(keyHash string) (userID int64, keyID int64, ok bool)) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": map[string]string{
					"message": "missing api key",
					"type":    "authentication_error",
				},
			})
			return
		}

		apiKey := strings.TrimPrefix(header, "Bearer ")
		hash := infrajwt.HashAPIKey(apiKey)
		userID, keyID, ok := resolver(hash)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": map[string]string{
					"message": "invalid api key",
					"type":    "authentication_error",
				},
			})
			return
		}

		c.Set("user_id", userID)
		c.Set("api_key_id", keyID)
		c.Next()
	}
}
