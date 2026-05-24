package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	coreinference "github.com/unillm/unillm/core/inference"
	"github.com/unillm/unillm/pkg/openai"
)

type ProxyHandler struct {
	inference *coreinference.Service
}

func NewProxyHandler(inference *coreinference.Service) *ProxyHandler {
	return &ProxyHandler{inference: inference}
}

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

	call := coreinference.CallContext{
		UserID:   c.GetInt64("user_id"),
		APIKeyID: c.GetInt64("api_key_id"),
	}

	if req.Stream {
		h.handleStream(c, call, &req)
		return
	}
	h.handleNonStream(c, call, &req)
}

func (h *ProxyHandler) handleNonStream(c *gin.Context, call coreinference.CallContext, req *openai.ChatCompletionRequest) {
	result, err := h.inference.ChatCompletion(c.Request.Context(), call, req)
	if err != nil {
		writeInferenceError(c, err)
		return
	}
	c.JSON(http.StatusOK, result.Response)
}

func (h *ProxyHandler) handleStream(c *gin.Context, call coreinference.CallContext, req *openai.ChatCompletionRequest) {
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

	writer := &ginStreamWriter{writer: c.Writer, flusher: flusher}
	_, err := h.inference.ChatCompletionStream(c.Request.Context(), call, req, writer)
	if err != nil {
		if !c.Writer.Written() {
			writeInferenceError(c, err)
		} else {
			log.Error().Err(err).Msg("proxy stream failed after headers sent")
		}
	}
}

type ginStreamWriter struct {
	writer  http.ResponseWriter
	flusher http.Flusher
}

func (w *ginStreamWriter) WriteLine(line string) error {
	_, err := fmt.Fprintf(w.writer, "%s\n", line)
	return err
}

func (w *ginStreamWriter) Flush() error {
	w.flusher.Flush()
	return nil
}

func writeInferenceError(c *gin.Context, err error) {
	infErr, ok := err.(*coreinference.InferenceError)
	if !ok {
		log.Error().Err(err).Msg("proxy request failed")
		c.JSON(http.StatusBadGateway, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "upstream error", Type: "upstream_error"},
		})
		return
	}

	if infErr.Cause != nil {
		log.Error().Str("kind", string(infErr.Kind)).Err(infErr.Cause).Msg("proxy request failed")
	}

	switch infErr.Kind {
	case coreinference.ErrModelNotFound:
		c.JSON(http.StatusNotFound, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: infErr.Message, Type: "invalid_request_error"},
		})
	case coreinference.ErrProviderMissing, coreinference.ErrProviderKey:
		c.JSON(http.StatusInternalServerError, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: infErr.Message, Type: "server_error"},
		})
	default:
		status := http.StatusBadGateway
		if infErr.HTTPStatus > 0 {
			status = infErr.HTTPStatus
		}
		c.JSON(status, openai.ErrorResponse{
			Error: openai.ErrorBody{Message: "upstream error: " + infErr.Message, Type: "upstream_error"},
		})
	}
}
