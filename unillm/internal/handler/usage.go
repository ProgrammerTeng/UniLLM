package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unillm/unillm/internal/model"
	"github.com/unillm/unillm/internal/service"
	"gorm.io/gorm"
)

type UsageHandler struct {
	db      *gorm.DB
	billing *service.BillingService
}

func NewUsageHandler(db *gorm.DB, billing *service.BillingService) *UsageHandler {
	return &UsageHandler{db: db, billing: billing}
}

// Summary returns aggregate usage stats for the authenticated user.
func (h *UsageHandler) Summary(c *gin.Context) {
	userID := c.GetInt64("user_id")

	// Get totals from PG
	var stats struct {
		TotalRequests    int64   `json:"total_requests"`
		TotalTokens      int64   `json:"total_tokens"`
		TotalCost        float64 `json:"total_cost"`
		AvgLatency       float64 `json:"avg_latency"`
		SuccessRate      float64 `json:"success_rate"`
	}

	h.db.Model(&model.UsageLog{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(latency), 0) as avg_latency
		`).Scan(&stats)

	var okCount int64
	h.db.Model(&model.UsageLog{}).Where("user_id = ? AND status = 'ok'", userID).Count(&okCount)
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(okCount) / float64(stats.TotalRequests) * 100
	}

	// Get today's cost from Redis (more real-time)
	todayCost, _ := h.billing.GetDailyUsage(c.Request.Context(), userID)

	// Get user balance
	var user model.User
	h.db.Select("balance").First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"total_requests": stats.TotalRequests,
		"total_tokens":   stats.TotalTokens,
		"total_cost":     stats.TotalCost,
		"avg_latency":    stats.AvgLatency,
		"success_rate":   stats.SuccessRate,
		"today_cost":     todayCost,
		"balance":        user.Balance,
	})
}

// ByModel returns per-model usage breakdown.
func (h *UsageHandler) ByModel(c *gin.Context) {
	userID := c.GetInt64("user_id")

	type ModelStats struct {
		ModelName    string  `json:"model_name"`
		Requests     int64   `json:"requests"`
		TotalTokens  int64   `json:"total_tokens"`
		TotalCost    float64 `json:"total_cost"`
		AvgLatency   float64 `json:"avg_latency"`
	}

	var results []ModelStats
	h.db.Model(&model.UsageLog{}).
		Where("user_id = ?", userID).
		Select(`
			model_name,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(latency), 0) as avg_latency
		`).
		Group("model_name").
		Order("total_cost DESC").
		Scan(&results)

	c.JSON(http.StatusOK, gin.H{"models": results})
}

// Daily returns daily usage for the last N days.
func (h *UsageHandler) Daily(c *gin.Context) {
	userID := c.GetInt64("user_id")
	days := 30

	type DailyStats struct {
		Date     string  `json:"date"`
		Requests int64   `json:"requests"`
		Tokens   int64   `json:"tokens"`
		Cost     float64 `json:"cost"`
	}

	var results []DailyStats
	since := time.Now().AddDate(0, 0, -days)
	h.db.Model(&model.UsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Select(`
			TO_CHAR(created_at, 'YYYY-MM-DD') as date,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as tokens,
			COALESCE(SUM(cost), 0) as cost
		`).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date").
		Scan(&results)

	c.JSON(http.StatusOK, gin.H{"daily": results})
}

// Recent returns the most recent usage logs with pagination.
func (h *UsageHandler) Recent(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var logs []model.UsageLog
	h.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&logs)

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
