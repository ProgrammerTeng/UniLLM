package v1

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/unillm/unillm/core/catalog"
	"github.com/unillm/unillm/infra/provider"
	"github.com/unillm/unillm/pkg/openai"
)

// ProviderHealth holds the health status for a provider.
type ProviderHealth struct {
	Name      string  `json:"name"`
	Status    string  `json:"status"` // up, down, degraded
	Latency   float64 `json:"latency_ms"`
	Circuit   string  `json:"circuit"` // closed, open, half-open
	CheckedAt string  `json:"checked_at"`
	Message   string  `json:"message,omitempty"`
}

// HealthRecord stores a single probe result for history.
type HealthRecord struct {
	Timestamp time.Time `json:"timestamp"`
	Provider  string    `json:"provider"`
	Status    string    `json:"status"`
	LatencyMs float64   `json:"latency_ms"`
}

// StatusHandler provides provider health endpoints with active probing.
type StatusHandler struct {
	registry    *provider.Registry
	catalog     *catalog.Service
	healthCache map[string]*ProviderHealth
	history     []HealthRecord // ring buffer of last 10080 records (7 days @ 1/min)
	historyIdx  int
	mu          sync.RWMutex
}

const maxHistoryRecords = 10080 // 7 days * 24 hours * 60 minutes

func NewStatusHandler(registry *provider.Registry, catalogSvc *catalog.Service) *StatusHandler {
	return &StatusHandler{
		registry:    registry,
		catalog:     catalogSvc,
		healthCache: make(map[string]*ProviderHealth),
		history:     make([]HealthRecord, 0, maxHistoryRecords),
	}
}

// StartHealthChecks runs periodic health checks with active probing.
func (h *StatusHandler) StartHealthChecks(interval time.Duration) {
	h.checkAll()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			h.checkAll()
		}
	}()
}

func (h *StatusHandler) checkAll() {
	ctx := context.Background()
	providers, err := h.catalog.ListActiveProviders(ctx)
	if err != nil {
		return
	}

	for _, p := range providers {
		prov, ok := h.registry.Get(p.Name)
		if !ok {
			continue
		}

		health := &ProviderHealth{
			Name:      p.Name,
			CheckedAt: time.Now().UTC().Format(time.RFC3339),
		}

		// Check circuit breaker state
		if rp, ok := prov.(*provider.ResilientProvider); ok {
			switch rp.CircuitBreakerState() {
			case provider.CircuitOpen:
				health.Circuit = "open"
			case provider.CircuitHalfOpen:
				health.Circuit = "half-open"
			default:
				health.Circuit = "closed"
			}
		} else {
			health.Circuit = "none"
		}

		latency, probeErr := h.activeProbe(ctx, p.Name, prov)
		health.Latency = latency
		if probeErr != nil {
			if health.Circuit == "open" {
				health.Status = "down"
			} else {
				health.Status = "degraded"
			}
			health.Message = probeErr.Error()
		} else {
			health.Status = "up"
		}

		log.Debug().Str("provider", p.Name).Str("status", health.Status).
			Float64("latency_ms", health.Latency).Msg("health check")

		h.mu.Lock()
		h.healthCache[p.Name] = health
		h.appendHistory(HealthRecord{
			Timestamp: time.Now(),
			Provider:  p.Name,
			Status:    health.Status,
			LatencyMs: health.Latency,
		})
		h.mu.Unlock()
	}
}

// activeProbe sends a lightweight request to test provider availability.
// Uses max_tokens=1 to minimize cost (~$0.00001 per probe).
func (h *StatusHandler) activeProbe(ctx context.Context, providerName string, prov provider.Provider) (float64, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	apiKey, err := h.catalog.NextKey(probeCtx, providerName)
	if err != nil {
		return 0, err
	}

	probeModel := "claude-haiku-4.5"
	models, _ := h.catalog.ListActiveModels(probeCtx)
	for _, m := range models {
		route, err := h.catalog.ResolveModel(probeCtx, m.PublicName)
		if err != nil {
			continue
		}
		if route.Provider.Name == providerName {
			probeModel = route.Model.UpstreamModel
			break
		}
	}

	req := &openai.ChatCompletionRequest{
		Model: probeModel,
		Messages: []openai.Message{{
			Role:    "user",
			Content: "hi",
		}},
		MaxTokens: 1,
	}

	start := time.Now()
	_, err = prov.ChatCompletion(probeCtx, apiKey, req)
	latency := float64(time.Since(start).Milliseconds())

	return latency, err
}

func (h *StatusHandler) appendHistory(r HealthRecord) {
	if len(h.history) < maxHistoryRecords {
		h.history = append(h.history, r)
	} else {
		h.history[h.historyIdx%maxHistoryRecords] = r
	}
	h.historyIdx++
}

// Status returns current health of all providers.
func (h *StatusHandler) Status(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	statuses := make([]ProviderHealth, 0, len(h.healthCache))
	allUp := true
	for _, s := range h.healthCache {
		statuses = append(statuses, *s)
		if s.Status != "up" {
			allUp = false
		}
	}

	overall := "operational"
	if !allUp {
		overall = "degraded"
	}
	if len(statuses) == 0 {
		overall = "unknown"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    overall,
		"providers": statuses,
		"checked":   time.Now().UTC().Format(time.RFC3339),
	})
}

// StatusHistory returns hourly uptime data for the last N days.
func (h *StatusHandler) StatusHistory(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Aggregate by provider+hour
	type hourKey struct {
		Provider string
		Hour     string
	}
	type hourStats struct {
		Total   int     `json:"total"`
		Up      int     `json:"up"`
		Uptime  float64 `json:"uptime_pct"`
		AvgMs   float64 `json:"avg_latency_ms"`
		sumMs   float64
	}

	buckets := make(map[hourKey]*hourStats)
	for _, r := range h.history {
		key := hourKey{
			Provider: r.Provider,
			Hour:     r.Timestamp.UTC().Truncate(time.Hour).Format("2006-01-02T15:00Z"),
		}
		if _, ok := buckets[key]; !ok {
			buckets[key] = &hourStats{}
		}
		b := buckets[key]
		b.Total++
		if r.Status == "up" {
			b.Up++
		}
		b.sumMs += r.LatencyMs
	}

	// Build response
	type historyEntry struct {
		Provider string  `json:"provider"`
		Hour     string  `json:"hour"`
		Uptime   float64 `json:"uptime_pct"`
		AvgMs    float64 `json:"avg_latency_ms"`
		Checks   int     `json:"checks"`
	}
	entries := make([]historyEntry, 0, len(buckets))
	for key, b := range buckets {
		if b.Total > 0 {
			b.Uptime = float64(b.Up) / float64(b.Total) * 100
			b.AvgMs = b.sumMs / float64(b.Total)
		}
		entries = append(entries, historyEntry{
			Provider: key.Provider,
			Hour:     key.Hour,
			Uptime:   b.Uptime,
			AvgMs:    b.AvgMs,
			Checks:   b.Total,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"history": entries,
	})
}
