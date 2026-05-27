package dashboard

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/unillm/unillm/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input service.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Register(input)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authSvc.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.authSvc.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":    user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"role":       user.Role,
		"balance":    user.Balance,
		"created_at": user.CreatedAt,
	})
}

type CreateKeyInput struct {
	Name  string `json:"name" binding:"required"`
	Scope string `json:"scope"`
}

func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var input CreateKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scope := input.Scope
	if scope == "" {
		scope = "full"
	}

	raw, key, err := h.authSvc.CreateAPIKey(userID, input.Name, scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"key":     raw,
		"id":      key.ID,
		"name":    key.Name,
		"prefix":  key.KeyPrefix,
		"scope":   key.Scope,
		"created": key.CreatedAt,
	})
}

func (h *AuthHandler) ListAPIKeys(c *gin.Context) {
	userID := c.GetInt64("user_id")
	keys, err := h.authSvc.ListAPIKeys(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": keys})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.authSvc.ChangePassword(userID, input.OldPassword, input.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"changed": true})
}

func (h *AuthHandler) DeleteAPIKey(c *gin.Context) {
	userID := c.GetInt64("user_id")
	var input struct {
		ID int64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.authSvc.DeleteAPIKey(input.ID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
