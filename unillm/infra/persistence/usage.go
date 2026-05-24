package persistence

import (
	"time"

	"github.com/unillm/unillm/internal/model"
	"gorm.io/gorm"
)

type UsageRepo struct {
	db *gorm.DB
}

func NewUsageRepo(db *gorm.DB) *UsageRepo {
	return &UsageRepo{db: db}
}

type UsageSummary struct {
	TotalRequests int64
	TotalTokens   int64
	TotalCost     float64
	AvgLatency    float64
	SuccessRate   float64
}

type ModelUsageStats struct {
	ModelName   string
	Requests    int64
	TotalTokens int64
	TotalCost   float64
	AvgLatency  float64
}

type DailyUsageStats struct {
	Date     string
	Requests int64
	Tokens   int64
	Cost     float64
}

type PlatformStats struct {
	TotalUsers    int64
	TotalRequests int64
	TotalCost     float64
	TotalTokens   int64
	ActiveKeys    int64
}

func (r *UsageRepo) Summary(userID int64) (UsageSummary, error) {
	var stats UsageSummary
	err := r.db.Model(&model.UsageLog{}).
		Where("user_id = ?", userID).
		Select(`
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost), 0) as total_cost,
			COALESCE(AVG(latency), 0) as avg_latency
		`).Scan(&stats).Error
	if err != nil {
		return stats, err
	}

	var okCount int64
	r.db.Model(&model.UsageLog{}).Where("user_id = ? AND status = 'ok'", userID).Count(&okCount)
	if stats.TotalRequests > 0 {
		stats.SuccessRate = float64(okCount) / float64(stats.TotalRequests) * 100
	}
	return stats, nil
}

func (r *UsageRepo) ByModel(userID int64) ([]ModelUsageStats, error) {
	var results []ModelUsageStats
	err := r.db.Model(&model.UsageLog{}).
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
		Scan(&results).Error
	return results, err
}

func (r *UsageRepo) Daily(userID int64, days int) ([]DailyUsageStats, error) {
	var results []DailyUsageStats
	since := time.Now().AddDate(0, 0, -days)
	err := r.db.Model(&model.UsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Select(`
			TO_CHAR(created_at, 'YYYY-MM-DD') as date,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as tokens,
			COALESCE(SUM(cost), 0) as cost
		`).
		Group("TO_CHAR(created_at, 'YYYY-MM-DD')").
		Order("date").
		Scan(&results).Error
	return results, err
}

func (r *UsageRepo) Recent(userID int64, limit int) ([]model.UsageLog, error) {
	var logs []model.UsageLog
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *UsageRepo) PlatformStats() (PlatformStats, error) {
	var stats PlatformStats
	r.db.Model(&model.User{}).Count(&stats.TotalUsers)
	r.db.Model(&model.UsageLog{}).Count(&stats.TotalRequests)
	r.db.Model(&model.UsageLog{}).Select("COALESCE(SUM(cost), 0)").Scan(&stats.TotalCost)
	r.db.Model(&model.UsageLog{}).Select("COALESCE(SUM(total_tokens), 0)").Scan(&stats.TotalTokens)
	r.db.Model(&model.APIKey{}).Where("is_active = true").Count(&stats.ActiveKeys)
	return stats, nil
}
