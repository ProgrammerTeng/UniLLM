package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unillm/unillm/internal/model"
	"github.com/unillm/unillm/internal/repository"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db           *gorm.DB
	providerRepo *repository.ProviderRepo
	userRepo     *repository.UserRepo
}

func NewAdminHandler(db *gorm.DB, providerRepo *repository.ProviderRepo, userRepo *repository.UserRepo) *AdminHandler {
	return &AdminHandler{db: db, providerRepo: providerRepo, userRepo: userRepo}
}

// ListUsers returns all users (admin only).
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []model.User
	if err := h.db.Select("id, email, name, role, balance, created_at, updated_at").
		Order("id ASC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// UpdateUserBalance adjusts a user's balance (admin only).
func (h *AdminHandler) UpdateUserBalance(c *gin.Context) {
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

// ListProviders returns all providers.
func (h *AdminHandler) ListProviders(c *gin.Context) {
	var providers []model.Provider
	if err := h.db.Find(&providers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// CreateProvider adds a new upstream provider.
func (h *AdminHandler) CreateProvider(c *gin.Context) {
	var input struct {
		Name    string `json:"name" binding:"required"`
		BaseURL string `json:"base_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p := model.Provider{Name: input.Name, BaseURL: input.BaseURL, IsActive: true}
	if err := h.db.Create(&p).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "provider already exists or " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// ToggleProvider enables/disables a provider.
func (h *AdminHandler) ToggleProvider(c *gin.Context) {
	var input struct {
		ID       int64 `json:"id" binding:"required"`
		IsActive bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Model(&model.Provider{}).Where("id = ?", input.ID).
		Update("is_active", input.IsActive).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// ListModels returns all model configurations.
func (h *AdminHandler) ListModels(c *gin.Context) {
	var models []model.ModelConfig
	if err := h.db.Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
}

// CreateModel adds a new model configuration.
func (h *AdminHandler) CreateModel(c *gin.Context) {
	var input model.ModelConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	input.IsActive = true

	if err := h.db.Create(&input).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "model already exists or " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

// UpdateModel updates an existing model configuration.
func (h *AdminHandler) UpdateModel(c *gin.Context) {
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

	if err := h.db.Model(&model.ModelConfig{}).Where("id = ?", input.ID).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": true})
}

// AddProviderKey adds an API key for an upstream provider.
func (h *AdminHandler) AddProviderKey(c *gin.Context) {
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

	pk := model.ProviderKey{
		ProviderID: input.ProviderID,
		KeyValue:   input.KeyValue,
		IsActive:   true,
		RPM:        rpm,
		Weight:     1,
	}
	if err := h.db.Create(&pk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":          pk.ID,
		"provider_id": pk.ProviderID,
		"rpm":         pk.RPM,
		"created":     pk.CreatedAt,
	})
}

// ListProviderKeys lists API keys for a provider (masked).
func (h *AdminHandler) ListProviderKeys(c *gin.Context) {
	var keys []model.ProviderKey
	if err := h.db.Find(&keys).Error; err != nil {
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
		prefix := k.KeyValue
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

// GlobalStats returns platform-wide statistics.
func (h *AdminHandler) GlobalStats(c *gin.Context) {
	var totalUsers int64
	h.db.Model(&model.User{}).Count(&totalUsers)

	var totalRequests int64
	h.db.Model(&model.UsageLog{}).Count(&totalRequests)

	var totalCost float64
	h.db.Model(&model.UsageLog{}).Select("COALESCE(SUM(cost), 0)").Scan(&totalCost)

	var totalTokens int64
	h.db.Model(&model.UsageLog{}).Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokens)

	var activeKeys int64
	h.db.Model(&model.APIKey{}).Where("is_active = true").Count(&activeKeys)

	c.JSON(http.StatusOK, gin.H{
		"total_users":    totalUsers,
		"total_requests": totalRequests,
		"total_cost":     totalCost,
		"total_tokens":   totalTokens,
		"active_keys":    activeKeys,
	})
}
