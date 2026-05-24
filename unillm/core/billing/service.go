package billing

import (
	"context"
	"time"
)

// Service handles balance checks and usage recording.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// CheckBalance returns true when the user has sufficient remaining balance.
func (s *Service) CheckBalance(ctx context.Context, userID int64, estimatedCost float64) (bool, error) {
	balance, err := s.store.GetUserBalance(ctx, userID)
	if err != nil {
		return false, err
	}

	used, err := s.store.GetUsedAmount(ctx, userID)
	if err != nil {
		return false, err
	}

	remaining := balance - used
	if remaining <= 0 {
		return false, nil
	}
	return remaining >= estimatedCost, nil
}

// Record persists usage asynchronously via the store.
func (s *Service) Record(ctx context.Context, record UsageRecord) error {
	return s.store.RecordUsage(ctx, record)
}

// RecordUsage is an alias for Record to ease migration from the legacy service.
func (s *Service) RecordUsage(ctx context.Context, record UsageRecord) error {
	return s.Record(ctx, record)
}

// GetDailyUsage returns today's accumulated cost for a user.
func (s *Service) GetDailyUsage(ctx context.Context, userID int64) (float64, error) {
	return s.store.GetDailyUsage(ctx, userID)
}

// FlushWorker periodically drains the usage log queue.
func (s *Service) FlushWorker(ctx context.Context, interval time.Duration) {
	s.store.FlushWorker(ctx, interval)
}

// FlushAll drains all remaining usage logs (e.g. on shutdown).
func (s *Service) FlushAll(ctx context.Context) {
	s.store.FlushAll(ctx)
}
