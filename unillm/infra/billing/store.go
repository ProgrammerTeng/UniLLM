package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/unillm/unillm/core/billing"
	"github.com/unillm/unillm/internal/model"
	"gorm.io/gorm"
)

// Store implements billing.Store with Redis counters and PostgreSQL flush.
type Store struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewStore(rdb *redis.Client, db *gorm.DB) *Store {
	return &Store{rdb: rdb, db: db}
}

func (s *Store) RecordUsage(ctx context.Context, record billing.UsageRecord) error {
	pipe := s.rdb.Pipeline()

	dayKey := fmt.Sprintf("usage:%d:%s", record.UserID, time.Now().Format("2006-01-02"))
	pipe.IncrByFloat(ctx, dayKey, record.Cost)
	pipe.Expire(ctx, dayKey, 48*time.Hour)

	totalKey := fmt.Sprintf("balance_used:%d", record.UserID)
	pipe.IncrByFloat(ctx, totalKey, record.Cost)

	modelKey := fmt.Sprintf("model_reqs:%s:%s", record.ModelName, time.Now().Format("2006-01-02-15"))
	pipe.Incr(ctx, modelKey)
	pipe.Expire(ctx, modelKey, 25*time.Hour)

	logBytes, err := json.Marshal(map[string]interface{}{
		"user_id":           record.UserID,
		"api_key_id":        record.APIKeyID,
		"model_name":        record.ModelName,
		"provider_name":     record.ProviderName,
		"prompt_tokens":     record.PromptTokens,
		"completion_tokens": record.CompletionTokens,
		"total_tokens":      record.TotalTokens,
		"cost":              record.Cost,
		"latency":           record.Latency,
		"status":            record.Status,
		"http_status":       record.HTTPStatus,
		"is_stream":         record.IsStream,
	})
	if err != nil {
		return fmt.Errorf("marshal usage log: %w", err)
	}
	pipe.RPush(ctx, "usage_log_queue", string(logBytes))

	_, execErr := pipe.Exec(ctx)
	return execErr
}

func (s *Store) GetUserBalance(ctx context.Context, userID int64) (float64, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Select("balance").First(&user, userID).Error; err != nil {
		return 0, err
	}
	return user.Balance, nil
}

func (s *Store) GetUsedAmount(ctx context.Context, userID int64) (float64, error) {
	totalKey := fmt.Sprintf("balance_used:%d", userID)
	usedStr, err := s.rdb.Get(ctx, totalKey).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	if usedStr == "" {
		return 0, nil
	}
	return strconv.ParseFloat(usedStr, 64)
}

func (s *Store) GetDailyUsage(ctx context.Context, userID int64) (float64, error) {
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

func (s *Store) FlushWorker(ctx context.Context, interval time.Duration) {
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

func (s *Store) FlushAll(ctx context.Context) {
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
		if err := s.db.WithContext(ctx).Create(&ul).Error; err != nil {
			log.Printf("[billing] db insert error during flush: %v", err)
			break
		}
	}
}

func (s *Store) flushOnce(ctx context.Context) {
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

		if err := s.db.WithContext(ctx).Create(&ul).Error; err != nil {
			log.Printf("[billing] db insert error: %v", err)
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
