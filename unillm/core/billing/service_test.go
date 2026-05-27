package billing

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockStore struct {
	balance    float64
	used       float64
	balanceErr error
	usedErr    error
}

func (m *mockStore) RecordUsage(context.Context, UsageRecord) error { return nil }
func (m *mockStore) GetUserBalance(context.Context, int64) (float64, error) {
	return m.balance, m.balanceErr
}
func (m *mockStore) GetUsedAmount(context.Context, int64) (float64, error) {
	return m.used, m.usedErr
}
func (m *mockStore) GetDailyUsage(context.Context, int64) (float64, error) { return 0, nil }
func (m *mockStore) FlushWorker(context.Context, time.Duration) {}
func (m *mockStore) FlushAll(context.Context)                                {}

func TestCheckBalance_Sufficient(t *testing.T) {
	svc := NewService(&mockStore{balance: 10, used: 3})
	ok, err := svc.CheckBalance(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected sufficient balance")
	}
}

func TestCheckBalance_Insufficient(t *testing.T) {
	svc := NewService(&mockStore{balance: 10, used: 9})
	ok, err := svc.CheckBalance(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected insufficient balance")
	}
}

func TestCheckBalance_ZeroRemaining(t *testing.T) {
	svc := NewService(&mockStore{balance: 10, used: 10})
	ok, err := svc.CheckBalance(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected zero remaining balance to fail")
	}
}

func TestCheckBalance_StoreError(t *testing.T) {
	svc := NewService(&mockStore{balanceErr: errors.New("db down")})
	_, err := svc.CheckBalance(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}
