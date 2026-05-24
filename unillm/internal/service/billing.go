package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unillm/unillm/internal/model"
	"gorm.io/gorm"
)

// BillingService handles real-time usage tracking via Redis and periodic flush to PG.
type BillingService struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewBillingService(rdb *redis.Client, db *gorm.DB) *BillingService {
	return &BillingService{rdb: rdb, db: db}
}

// RecordUsage atomically increments usage counters in Redis and queues for PG flush.
func (s *BillingService) RecordUsage(ctx context.Context, log model.UsageLog) error {
	pipe := s.rdb.Pipeline()

	// Per-user daily cost counter
	dayKey := fmt.Sprintf("usage:%d:%s", log.UserID, time.Now().Format("2006-01-02"))
	pipe.IncrByFloat(ctx, dayKey, log.Cost)
	pipe.Expire(ctx, dayKey, 48*time.Hour)

	// Per-user total cost counter
	totalKey := fmt.Sprintf("balance_used:%d", log.UserID)
	pipe.IncrByFloat(ctx, totalKey, log.Cost)

	// Per-model request counter (for status page)
	modelKey := fmt.Sprintf("model_reqs:%s:%s", log.ModelName, time.Now().Format("2006-01-02-15"))
	pipe.Incr(ctx, modelKey)
	pipe.Expire(ctx, modelKey, 25*time.Hour)

	// Queue the full log entry for async PG flush (safe JSON encoding)
	logBytes, err := json.Marshal(map[string]interface{}{
		"user_id":           log.UserID,
		"api_key_id":        log.APIKeyID,
		"model_name":        log.ModelName,
		"provider_name":     log.ProviderName,
		"prompt_tokens":     log.PromptTokens,
		"completion_tokens": log.CompletionTokens,
		"total_tokens":      log.TotalTokens,
		"cost":              log.Cost,
		"latency":           log.Latency,
		"status":            log.Status,
		"http_status":       log.HTTPStatus,
		"is_stream":         log.IsStream,
	})
	if err != nil {
		return fmt.Errorf("marshal usage log: %w", err)
	}
	pipe.RPush(ctx, "usage_log_queue", string(logBytes))

	_, execErr := pipe.Exec(ctx)
	return execErr
}

// CheckBalance returns true if user has sufficient balance.
func (s *BillingService) CheckBalance(ctx context.Context, userID int64, estimatedCost float64) (bool, error) {
	// Get user balance from PG
	var user model.User
	if err := s.db.Select("balance").First(&user, userID).Error; err != nil {
		return false, err
	}

	// Get used amount from Redis
	totalKey := fmt.Sprintf("balance_used:%d", userID)
	usedStr, err := s.rdb.Get(ctx, totalKey).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}
	used := 0.0
	if usedStr != "" {
		used, _ = strconv.ParseFloat(usedStr, 64)
	}

	remaining := user.Balance - used
	// Require positive remaining balance (not just >= 0)
	if remaining <= 0 {
		return false, nil
	}
	return remaining >= estimatedCost, nil
}

// GetDailyUsage returns today's cost for a user.
func (s *BillingService) GetDailyUsage(ctx context.Context, userID int64) (float64, error) {
	dayKey := fmt.Sprintf("usage:%d:%s", userID, time.Now().Format("2006-01-02"))
	val, err := s.rdb.Get(ctx, dayKey).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

// FlushWorker runs periodically to flush usage logs from Redis to PostgreSQL.
func (s *BillingService) FlushWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.flushOnce(ctx)
		}
	}
}

// FlushAll drains all remaining entries from the queue (used during shutdown).
func (s *BillingService) FlushAll(ctx context.Context) {
	for {
		data, err := s.rdb.LPop(ctx, "usage_log_queue").Result()
		if err != nil {
			break
		}
		var ul model.UsageLog
		if err := parseUsageLog(data, &ul); err != nil {
			log.Printf("[billing] parse error during flush: %v", err)
			continue
		}
		if err := s.db.Create(&ul).Error; err != nil {
			log.Printf("[billing] db insert error during flush: %v", err)
			break
		}
	}
}

func (s *BillingService) flushOnce(ctx context.Context) {
	// Pop up to 100 entries at a time
	for i := 0; i < 100; i++ {
		data, err := s.rdb.LPop(ctx, "usage_log_queue").Result()
		if err == redis.Nil {
			break
		}
		if err != nil {
			log.Printf("[billing] redis lpop error: %v", err)
			break
		}

		var ul model.UsageLog
		if err := parseUsageLog(data, &ul); err != nil {
			log.Printf("[billing] parse error: %v", err)
			continue
		}

		if err := s.db.Create(&ul).Error; err != nil {
			log.Printf("[billing] db insert error: %v", err)
			// Push back to queue on failure
			s.rdb.RPush(ctx, "usage_log_queue", data)
			break
		}
	}
}

var jsonUnmarshal = json.Unmarshal

func parseUsageLog(data string, ul *model.UsageLog) error {
	type logEntry struct {
		UserID           int64   `json:"user_id"`
		APIKeyID         int64   `json:"api_key_id"`
		ModelName        string  `json:"model_name"`
		ProviderName     string  `json:"provider_name"`
		PromptTokens     int     `json:"prompt_tokens"`
		CompletionTokens int     `json:"completion_tokens"`
		TotalTokens      int     `json:"total_tokens"`
		Cost             float64 `json:"cost"`
		Latency          float64 `json:"latency"`
		Status           string  `json:"status"`
		HTTPStatus       int     `json:"http_status"`
		IsStream         bool    `json:"is_stream"`
	}

	var entry logEntry
	if err := jsonUnmarshal([]byte(data), &entry); err != nil {
		return err
	}

	ul.UserID = entry.UserID
	ul.APIKeyID = entry.APIKeyID
	ul.ModelName = entry.ModelName
	ul.ProviderName = entry.ProviderName
	ul.PromptTokens = entry.PromptTokens
	ul.CompletionTokens = entry.CompletionTokens
	ul.TotalTokens = entry.TotalTokens
	ul.Cost = entry.Cost
	ul.Latency = entry.Latency
	ul.Status = entry.Status
	ul.HTTPStatus = entry.HTTPStatus
	ul.IsStream = entry.IsStream
	return nil
}
