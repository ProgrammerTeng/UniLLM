package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/unillm/unillm/internal/middleware"
	"github.com/unillm/unillm/internal/model"
	"github.com/unillm/unillm/internal/provider"
	"github.com/unillm/unillm/internal/repository"
	"github.com/unillm/unillm/internal/service"
	"github.com/unillm/unillm/pkg/openai"
)

type ProxyHandler struct {
	registry     *provider.Registry
	providerRepo *repository.ProviderRepo
	billing      *service.BillingService
	providerKeys map[string][]string
	keyIndex     map[string]uint64
	mu           sync.Mutex
}

func NewProxyHandler(registry *provider.Registry, providerRepo *repository.ProviderRepo, billing *service.BillingService) *ProxyHandler {
	return &ProxyHandler{
		registry:     registry,
		providerRepo: providerRepo,
		billing:      billing,
		providerKeys: make(map[string][]string),
		keyIndex:     make(map[string]uint64),
	}
}

// LoadProviderKeys loads API keys from the database into memory.
func (h *ProxyHandler) LoadProviderKeys(providerRepo *repository.ProviderRepo) error {
	providers, err := providerRepo.ListActive()
	if err != nil {
		return err
	}
	for _, p := range providers {
		keys, err := providerRepo.ListActiveKeys(p.ID)
		if err != nil {
			continue
		}
		for _, k := range keys {
			h.providerKeys[p.Name] = append(h.providerKeys[p.Name], k.KeyValue)
		}
	}
	return nil
}

// GetNextKey returns the next API key for a provider using round-robin.
func (h *ProxyHandler) GetNextKey(providerName string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	keys := h.providerKeys[providerName]
	if len(keys) == 0 {
		return "", fmt.Errorf("no API keys configured for provider %s", providerName)
	}
	idx := h.keyIndex[providerName] % uint64(len(keys))
	h.keyIndex[providerName]++
	return keys[idx], nil
}

// ChatCompletion handles POST /v1/chat/completions.
func (h *ProxyHandler) ChatCompletion(c *gin.Context) {
	var req openai.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "invalid request: " + err.Error(), Type: "invalid_request_error"},
		})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "model is required", Type: "invalid_request_error"},
		})
		return
	}

	modelCfg, err := h.providerRepo.FindModelByPublicName(req.Model)
	if err != nil {
		c.JSON(http.StatusNotFound, openai.ErrorResponse{
			Error: openai.ErrorBody{
				Message: fmt.Sprintf("model '%s' not found or not available", req.Model),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	providerInfo, err := h.providerRepo.FindByID(modelCfg.ProviderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "provider lookup failed", Type: "server_error"},
		})
		return
	}

	p, ok := h.registry.Get(providerInfo.Name)
	if !ok {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "provider not configured", Type: "server_error"},
		})
		return
	}

	apiKey, err := h.GetNextKey(providerInfo.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: err.Error(), Type: "server_error"},
		})
		return
	}

	publicModel := req.Model
	req.Model = modelCfg.UpstreamModel

	// Reasoning models (o1, o3, etc.) use max_completion_tokens instead of max_tokens
	if isReasoningModel(publicModel) && req.MaxTokens > 0 {
		if req.MaxCompletionTokens == 0 {
			req.MaxCompletionTokens = req.MaxTokens
		}
		req.MaxTokens = 0
	}

	start := time.Now()
	userID := c.GetInt64("user_id")
	apiKeyID := c.GetInt64("api_key_id")

	if req.Stream {
		// Always request usage in stream for billing
		if req.StreamOptions == nil {
			req.StreamOptions = &openai.StreamOptions{IncludeUsage: true}
		} else {
			req.StreamOptions.IncludeUsage = true
		}
		h.handleStream(c, p, apiKey, &req, modelCfg, providerInfo, publicModel, userID, apiKeyID, start)
	} else {
		h.handleNonStream(c, p, apiKey, &req, modelCfg, providerInfo, publicModel, userID, apiKeyID, start)
	}
}

func (h *ProxyHandler) handleNonStream(c *gin.Context, p provider.Provider, apiKey string,
	req *openai.ChatCompletionRequest, modelCfg *model.ModelConfig, providerInfo *model.Provider,
	publicModel string, userID, apiKeyID int64, start time.Time) {

	resp, err := p.ChatCompletion(c.Request.Context(), apiKey, req)
	latency := time.Since(start).Seconds()

	if err != nil {
		log.Error().Str("model", publicModel).Str("provider", providerInfo.Name).Err(err).Msg("proxy request failed")
		// Log error usage
		h.logUsage(userID, apiKeyID, publicModel, providerInfo.Name, 0, 0, 0, 0, latency, "error", 502, false)
		c.JSON(http.StatusBadGateway, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "upstream error: " + err.Error(), Type: "upstream_error"},
		})
		return
	}

	resp.Model = publicModel
	cost := calculateCost(modelCfg, resp.Usage)

	// Async log usage
	pt, ct, tt := 0, 0, 0
	if resp.Usage != nil {
		pt = resp.Usage.PromptTokens
		ct = resp.Usage.CompletionTokens
		tt = resp.Usage.TotalTokens
	}
	h.logUsage(userID, apiKeyID, publicModel, providerInfo.Name, pt, ct, tt, cost, latency, "ok", 200, false)

	c.JSON(http.StatusOK, resp)
}

func (h *ProxyHandler) handleStream(c *gin.Context, p provider.Provider, apiKey string,
	req *openai.ChatCompletionRequest, modelCfg *model.ModelConfig, providerInfo *model.Provider,
	publicModel string, userID, apiKeyID int64, start time.Time) {

	stream, err := p.ChatCompletionStream(c.Request.Context(), apiKey, req)
	if err != nil {
		log.Error().Str("model", publicModel).Str("provider", providerInfo.Name).Err(err).Msg("proxy stream failed")
		latency := time.Since(start).Seconds()
		h.logUsage(userID, apiKeyID, publicModel, providerInfo.Name, 0, 0, 0, 0, latency, "error", 502, true)
		c.JSON(http.StatusBadGateway, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "upstream error: " + err.Error(), Type: "upstream_error"},
		})
		return
	}
	defer stream.Close()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "streaming not supported", Type: "server_error"},
		})
		return
	}

	// Replace upstream model name with public model name in stream chunks,
	// and extract usage from the final chunk for billing.
	// Also estimate tokens from content when upstream doesn't return usage.
	var streamPT, streamCT, streamTT int
	var totalContentChars int
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") && line != "data: [DONE]" {
			// Replace model name — match any "model":"..." pattern
			line = replaceModelName(line, publicModel)
			// Try to extract usage from final chunks
			if strings.Contains(line, `"usage"`) {
				streamPT, streamCT, streamTT = extractStreamUsage(line)
			}
			// Accumulate content characters for token estimation
			totalContentChars += extractContentLength(line)
		}
		if _, err := fmt.Fprintf(c.Writer, "%s\n", line); err != nil {
			break
		}
		flusher.Flush()
	}

	latency := time.Since(start).Seconds()

	// If upstream didn't return usage, estimate tokens from content
	if streamPT == 0 && streamCT == 0 {
		streamPT = estimatePromptTokens(req.Messages)
		streamCT = max(1, totalContentChars/4) // ~4 chars per token
		streamTT = streamPT + streamCT
	}

	cost := float64(streamPT)*modelCfg.InputPricePer1M/1_000_000 + float64(streamCT)*modelCfg.OutputPricePer1M/1_000_000
	h.logUsage(userID, apiKeyID, publicModel, providerInfo.Name, streamPT, streamCT, streamTT, cost, latency, "ok", 200, true)
}

func (h *ProxyHandler) logUsage(userID, apiKeyID int64, modelName, providerName string,
	promptTok, completionTok, totalTok int, cost, latency float64, status string, httpStatus int, isStream bool) {
	// Record Prometheus metrics
	middleware.RecordProxy(modelName, providerName, status, latency, promptTok, completionTok)

	go func() {
		ul := model.UsageLog{
			UserID:           userID,
			APIKeyID:         apiKeyID,
			ModelName:        modelName,
			ProviderName:     providerName,
			PromptTokens:     promptTok,
			CompletionTokens: completionTok,
			TotalTokens:      totalTok,
			Cost:             cost,
			Latency:          latency,
			Status:           status,
			HTTPStatus:       httpStatus,
			IsStream:         isStream,
		}
		if err := h.billing.RecordUsage(context.Background(), ul); err != nil {
			log.Error().Err(err).Msg("billing record failed")
		}
	}()
}

// replaceModelName replaces any "model":"..." value with the public model name.
func replaceModelName(line, publicModel string) string {
	const prefix = `"model":"`
	idx := strings.Index(line, prefix)
	if idx < 0 {
		return line
	}
	start := idx + len(prefix)
	end := strings.Index(line[start:], `"`)
	if end < 0 {
		return line
	}
	return line[:start] + publicModel + line[start+end:]
}

// extractContentLength extracts the length of the "content" field from a stream chunk.
func extractContentLength(line string) int {
	data := strings.TrimPrefix(line, "data: ")
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err == nil && len(chunk.Choices) > 0 {
		return len(chunk.Choices[0].Delta.Content)
	}
	return 0
}

// estimatePromptTokens estimates token count from request messages (~4 chars per token).
func estimatePromptTokens(messages []openai.Message) int {
	total := 0
	for _, m := range messages {
		switch v := m.Content.(type) {
		case string:
			total += len(v)
		default:
			b, _ := json.Marshal(v)
			total += len(b)
		}
		total += 4 // role overhead
	}
	return max(1, total/4)
}

func extractStreamUsage(line string) (pt, ct, tt int) {
	data := strings.TrimPrefix(line, "data: ")
	var chunk struct {
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err == nil && chunk.Usage != nil {
		return chunk.Usage.PromptTokens, chunk.Usage.CompletionTokens, chunk.Usage.TotalTokens
	}
	return 0, 0, 0
}

// isReasoningModel returns true for models that use max_completion_tokens instead of max_tokens.
func isReasoningModel(model string) bool {
	m := strings.ToLower(model)
	return strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4")
}

func calculateCost(modelCfg *model.ModelConfig, usage *openai.Usage) float64 {
	if usage == nil {
		return 0
	}
	inputCost := float64(usage.PromptTokens) * modelCfg.InputPricePer1M / 1_000_000
	outputCost := float64(usage.CompletionTokens) * modelCfg.OutputPricePer1M / 1_000_000
	return inputCost + outputCost
}
