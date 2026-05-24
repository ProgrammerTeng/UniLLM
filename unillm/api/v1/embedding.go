package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	coreinference "github.com/unillm/unillm/core/inference"
	"github.com/unillm/unillm/pkg/openai"
)

type EmbeddingHandler struct {
	inference *coreinference.Service
}

func NewEmbeddingHandler(inference *coreinference.Service) *EmbeddingHandler {
	return &EmbeddingHandler{inference: inference}
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

	call := coreinference.CallContext{
		UserID:   c.GetInt64("user_id"),
		APIKeyID: c.GetInt64("api_key_id"),
	}

	result, err := h.inference.CreateEmbedding(c.Request.Context(), call, &req)
	if err != nil {
		writeInferenceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result.Response)
}
