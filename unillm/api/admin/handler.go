package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unillm/unillm/core/catalog"
	"github.com/unillm/unillm/infra/crypto"
	"github.com/unillm/unillm/infra/persistence"
	"github.com/unillm/unillm/internal/model"
)

type Handler struct {
	providerRepo *persistence.ProviderRepo
	userRepo     *persistence.UserRepo
	usageRepo    *persistence.UsageRepo
	catalog      *catalog.Service
	protector    *crypto.KeyProtector
}

func NewHandler(
	providerRepo *persistence.ProviderRepo,
	userRepo *persistence.UserRepo,
	usageRepo *persistence.UsageRepo,
	catalogSvc *catalog.Service,
	protector *crypto.KeyProtector,
) *Handler {
	return &Handler{
		providerRepo: providerRepo,
		userRepo:     userRepo,
		usageRepo:    usageRepo,
		catalog:      catalogSvc,
		protector:    protector,
	}
}

func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.userRepo.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *Handler) UpdateUserBalance(c *gin.Context) {
	var input struct {
		UserID int64   `json:"user_id" binding:"required"`
		Delta  float64 `json:"delta" binding:"required"`
		Reason string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userRepo.UpdateBalance(input.UserID, input.Delta); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user, _ := h.userRepo.FindByID(input.UserID)
	c.JSON(http.StatusOK, gin.H{
		"user_id":     input.UserID,
		"delta":       input.Delta,
		"reason":      input.Reason,
		"new_balance": user.Balance,
	})
}

func (h *Handler) ListProviders(c *gin.Context) {
	providers, err := h.providerRepo.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (h *Handler) CreateProvider(c *gin.Context) {
	var input struct {
		Name    string `json:"name" binding:"required"`
		BaseURL string `json:"base_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := model.Provider{Name: input.Name, BaseURL: input.BaseURL, IsActive: true}
	if err := h.providerRepo.Create(&p); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "provider already exists or " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) ToggleProvider(c *gin.Context) {
	var input struct {
		ID       int64 `json:"id" binding:"required"`
		IsActive bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.providerRepo.SetActive(input.ID, input.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

func (h *Handler) ListModels(c *gin.Context) {
	models, err := h.providerRepo.ListAllModels()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

func (h *Handler) CreateModel(c *gin.Context) {
	var input model.ModelConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.IsActive = true

	if err := h.providerRepo.CreateModel(&input); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "model already exists or " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *Handler) UpdateModel(c *gin.Context) {
	var input struct {
		ID               int64    `json:"id" binding:"required"`
		InputPricePer1M  *float64 `json:"input_price_per_1m"`
		OutputPricePer1M *float64 `json:"output_price_per_1m"`
		IsActive         *bool    `json:"is_active"`
		MaxTokens        *int     `json:"max_tokens"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.InputPricePer1M != nil {
		updates["input_price_per_1m"] = *input.InputPricePer1M
	}
	if input.OutputPricePer1M != nil {
		updates["output_price_per_1m"] = *input.OutputPricePer1M
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.MaxTokens != nil {
		updates["max_tokens"] = *input.MaxTokens
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := h.providerRepo.UpdateModel(input.ID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

func (h *Handler) AddProviderKey(c *gin.Context) {
	var input struct {
		ProviderID int64  `json:"provider_id" binding:"required"`
		KeyValue   string `json:"key_value" binding:"required"`
		RPM        int    `json:"rpm"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rpm := input.RPM
	if rpm == 0 {
		rpm = 60
	}

	stored, err := h.protector.Encrypt(input.KeyValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt key: " + err.Error()})
		return
	}

	pk := model.ProviderKey{
		ProviderID: input.ProviderID,
		KeyValue:   stored,
		IsActive:   true,
		RPM:        rpm,
		Weight:     1,
	}
	if err := h.providerRepo.CreateProviderKey(&pk); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = h.catalog.Reload(c.Request.Context())

	c.JSON(http.StatusCreated, gin.H{
		"id":          pk.ID,
		"provider_id": pk.ProviderID,
		"rpm":         pk.RPM,
		"created":     pk.CreatedAt,
		"encrypted":   h.protector.Enabled(),
	})
}

func (h *Handler) ListProviderKeys(c *gin.Context) {
	keys, err := h.providerRepo.ListAllProviderKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type maskedKey struct {
		ID         int64  `json:"id"`
		ProviderID int64  `json:"provider_id"`
		Prefix     string `json:"key_prefix"`
		RPM        int    `json:"rpm"`
		IsActive   bool   `json:"is_active"`
		Weight     int    `json:"weight"`
	}

	var result []maskedKey
	for _, k := range keys {
		plain := h.protector.Decrypt(k.KeyValue)
		prefix := plain
		if len(prefix) > 12 {
			prefix = prefix[:12] + "..."
		}
		result = append(result, maskedKey{
			ID:         k.ID,
			ProviderID: k.ProviderID,
			Prefix:     prefix,
			RPM:        k.RPM,
			IsActive:   k.IsActive,
			Weight:     k.Weight,
		})
	}
	c.JSON(http.StatusOK, gin.H{"keys": result})
}

func (h *Handler) GlobalStats(c *gin.Context) {
	stats, err := h.usageRepo.PlatformStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total_users":    stats.TotalUsers,
		"total_requests": stats.TotalRequests,
		"total_cost":     stats.TotalCost,
		"total_tokens":   stats.TotalTokens,
		"active_keys":    stats.ActiveKeys,
	})
}
