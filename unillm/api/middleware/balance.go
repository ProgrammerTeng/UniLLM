package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BalanceChecker is a function that checks if a user has sufficient balance.
type BalanceChecker func(ctx context.Context, userID int64) (bool, error)

// BalanceCheck middleware rejects requests when user balance is insufficient.
func BalanceCheck(checker BalanceChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		if userID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": map[string]string{
					"message": "authentication required",
					"type":    "authentication_error",
				},
			})
			return
		}

		ok, err := checker(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": map[string]string{
					"message": "balance check failed",
					"type":    "server_error",
				},
			})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error": map[string]string{
					"message": "insufficient balance, please top up your account",
					"type":    "billing_error",
				},
			})
			return
		}

		c.Next()
	}
}
