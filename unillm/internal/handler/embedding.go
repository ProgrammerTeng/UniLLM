package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/unillm/unillm/internal/model"
	"github.com/unillm/unillm/internal/provider"
	"github.com/unillm/unillm/internal/repository"
	"github.com/unillm/unillm/internal/service"
	"github.com/unillm/unillm/pkg/openai"
)

// EmbeddingHandler handles POST /v1/embeddings.
type EmbeddingHandler struct {
	providerRepo *repository.ProviderRepo
	billing      *service.BillingService
	providerKeys map[string][]string
	client       *http.Client
}

func NewEmbeddingHandler(providerRepo *repository.ProviderRepo, billing *service.BillingService) *EmbeddingHandler {
	return &EmbeddingHandler{
		providerRepo: providerRepo,
		billing:      billing,
		providerKeys: make(map[string][]string),
		client:       provider.NewStandardClient(30 * time.Second),
	}
}

// LoadProviderKeys loads API keys for embedding providers.
func (h *EmbeddingHandler) LoadProviderKeys(providerRepo *repository.ProviderRepo) error {
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

func (h *EmbeddingHandler) CreateEmbedding(c *gin.Context) {
	var req openai.EmbeddingRequest
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
				Message: fmt.Sprintf("embedding model '%s' not found", req.Model),
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

	keys := h.providerKeys[providerInfo.Name]
	if len(keys) == 0 {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "no API keys for provider", Type: "server_error"},
		})
		return
	}
	apiKey := keys[0]

	// Forward to upstream
	upstreamReq := req
	upstreamReq.Model = modelCfg.UpstreamModel

	body, err := json.Marshal(upstreamReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "marshal failed", Type: "server_error"},
		})
		return
	}

	url := providerInfo.BaseURL + "/embeddings"
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "POST", url, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "create request failed", Type: "server_error"},
		})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	start := time.Now()
	resp, err := h.client.Do(httpReq)
	latency := time.Since(start).Seconds()

	if err != nil {
		log.Error().Err(err).Str("model", req.Model).Msg("embedding upstream error")
		c.JSON(http.StatusBadGateway, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "upstream error: " + err.Error(), Type: "upstream_error"},
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Error().Int("status", resp.StatusCode).Str("model", req.Model).Str("body", string(respBody)[:200]).Msg("embedding upstream error")
		c.JSON(resp.StatusCode, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: fmt.Sprintf("upstream error (status %d)", resp.StatusCode), Type: "upstream_error"},
		})
		return
	}

	var embResp openai.EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "decode response failed", Type: "server_error"},
		})
		return
	}

	// Replace model name with public name
	embResp.Model = req.Model

	// Log usage
	userID := c.GetInt64("user_id")
	apiKeyID := c.GetInt64("api_key_id")
	pt := 0
	if embResp.Usage != nil {
		pt = embResp.Usage.PromptTokens
	}
	cost := float64(pt) * modelCfg.InputPricePer1M / 1_000_000

	go func() {
		ul := model.UsageLog{
			UserID:       userID,
			APIKeyID:     apiKeyID,
			ModelName:    req.Model,
			ProviderName: providerInfo.Name,
			PromptTokens: pt,
			TotalTokens:  pt,
			Cost:         cost,
			Latency:      latency,
			Status:       "ok",
			HTTPStatus:   200,
		}
		if err := h.billing.RecordUsage(context.Background(), ul); err != nil {
			log.Error().Err(err).Msg("embedding billing record failed")
		}
	}()

	c.JSON(http.StatusOK, embResp)
}
