package inference

import (
	"context"

	corebilling "github.com/unillm/unillm/core/billing"
	coreinference "github.com/unillm/unillm/core/inference"
)

// BillingRecorder adapts core/billing.Service to inference.BillingRecorder.
type BillingRecorder struct {
	Billing *corebilling.Service
}

func (b *BillingRecorder) RecordUsage(ctx context.Context, record coreinference.UsageRecord) error {
	return b.Billing.Record(ctx, corebilling.UsageRecord{
		UserID:           record.UserID,
		APIKeyID:         record.APIKeyID,
		ModelName:        record.ModelName,
		ProviderName:     record.ProviderName,
		PromptTokens:     record.PromptTokens,
		CompletionTokens: record.CompletionTokens,
		TotalTokens:      record.TotalTokens,
		Cost:             record.Cost,
		Latency:          record.Latency,
		Status:           record.Status,
		HTTPStatus:       record.HTTPStatus,
		IsStream:         record.IsStream,
	})
}
