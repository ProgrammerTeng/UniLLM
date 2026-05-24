package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	corebilling "github.com/unillm/unillm/core/billing"
	"github.com/unillm/unillm/infra/persistence"
)

type UsageHandler struct {
	usage   *persistence.UsageRepo
	users   *persistence.UserRepo
	billing *corebilling.Service
}

func NewUsageHandler(usage *persistence.UsageRepo, users *persistence.UserRepo, billing *corebilling.Service) *UsageHandler {
	return &UsageHandler{usage: usage, users: users, billing: billing}
}

func (h *UsageHandler) Summary(c *gin.Context) {
	userID := c.GetInt64("user_id")

	stats, err := h.usage.Summary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todayCost, _ := h.billing.GetDailyUsage(c.Request.Context(), userID)
	balance, _ := h.users.GetBalance(userID)

	c.JSON(http.StatusOK, gin.H{
		"total_requests": stats.TotalRequests,
		"total_tokens":   stats.TotalTokens,
		"total_cost":     stats.TotalCost,
		"avg_latency":    stats.AvgLatency,
		"success_rate":   stats.SuccessRate,
		"today_cost":     todayCost,
		"balance":        balance,
	})
}

func (h *UsageHandler) ByModel(c *gin.Context) {
	userID := c.GetInt64("user_id")
	results, err := h.usage.ByModel(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": results})
}

func (h *UsageHandler) Daily(c *gin.Context) {
	userID := c.GetInt64("user_id")
	results, err := h.usage.Daily(userID, 30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"daily": results})
}

func (h *UsageHandler) Recent(c *gin.Context) {
	userID := c.GetInt64("user_id")
	logs, err := h.usage.Recent(userID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
