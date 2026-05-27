package v1

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/unillm/unillm/infra/persistence"
)

type ModelsHandler struct {
	providerRepo *persistence.ProviderRepo
}

func NewModelsHandler(providerRepo *persistence.ProviderRepo) *ModelsHandler {
	return &ModelsHandler{providerRepo: providerRepo}
}

func (h *ModelsHandler) ListModels(c *gin.Context) {
	models, err := h.providerRepo.ListActiveModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list models"})
		return
	}

	data := make([]gin.H, 0, len(models))
	for _, m := range models {
		data = append(data, gin.H{
			"id":       m.PublicName,
			"object":   "model",
			"created":  time.Now().Unix(),
			"owned_by": "unillm",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data":   data,
	})
}

func (h *ModelsHandler) ModelCatalog(c *gin.Context) {
	type catalogItem struct {
		ID               string  `json:"id"`
		Vendor           string  `json:"vendor"`
		InputPricePer1M  float64 `json:"input_price_per_1m"`
		OutputPricePer1M float64 `json:"output_price_per_1m"`
		MaxTokens        int     `json:"max_tokens"`
		SupportsStream   bool    `json:"supports_stream"`
		SupportsTools    bool    `json:"supports_tools"`
		SupportsVision   bool    `json:"supports_vision"`
	}

	models, err := h.providerRepo.ListActiveModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list models"})
		return
	}

	items := make([]catalogItem, 0, len(models))
	for _, m := range models {
		items = append(items, catalogItem{
			ID:               m.PublicName,
			Vendor:           inferVendor(m.PublicName),
			InputPricePer1M:  m.InputPricePer1M,
			OutputPricePer1M: m.OutputPricePer1M,
			MaxTokens:        m.MaxTokens,
			SupportsStream:   m.SupportsStream,
			SupportsTools:    m.SupportsTools,
			SupportsVision:   m.SupportsVision,
		})
	}

	c.JSON(http.StatusOK, gin.H{"models": items})
}

func inferVendor(modelName string) string {
	name := strings.ToLower(modelName)
	switch {
	case strings.HasPrefix(name, "claude"):
		return "Anthropic"
	case strings.HasPrefix(name, "gemini"):
		return "Google"
	case strings.HasPrefix(name, "gpt"), strings.HasPrefix(name, "o1"), strings.HasPrefix(name, "o3"), strings.HasPrefix(name, "o4"):
		return "OpenAI"
	case strings.Contains(name, "deepseek"):
		return "DeepSeek"
	case strings.Contains(name, "qwen"):
		return "Alibaba"
	case strings.Contains(name, "llama"), strings.Contains(name, "meta"):
		return "Meta"
	case strings.Contains(name, "mistral"), strings.Contains(name, "mixtral"):
		return "Mistral"
	default:
		return "Other"
	}
}
