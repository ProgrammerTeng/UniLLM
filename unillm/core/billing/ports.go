package billing

import (
	"context"
	"time"
)

// Store persists usage counters and flushes logs to durable storage.
type Store interface {
	RecordUsage(ctx context.Context, record UsageRecord) error
	GetUserBalance(ctx context.Context, userID int64) (float64, error)
	GetUsedAmount(ctx context.Context, userID int64) (float64, error)
	GetDailyUsage(ctx context.Context, userID int64) (float64, error)
	FlushWorker(ctx context.Context, interval time.Duration)
	FlushAll(ctx context.Context)
}
