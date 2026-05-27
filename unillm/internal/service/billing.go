package service

import (
	"github.com/redis/go-redis/v9"
	corebilling "github.com/unillm/unillm/core/billing"
	infrabilling "github.com/unillm/unillm/infra/billing"
	"gorm.io/gorm"
)

// BillingService is deprecated; use core/billing.Service directly.
type BillingService = corebilling.Service

// NewBillingService constructs a billing service backed by Redis and PostgreSQL.
func NewBillingService(rdb *redis.Client, db *gorm.DB) *BillingService {
	return corebilling.NewService(infrabilling.NewStore(rdb, db))
}
