package inference

import (
	"context"
	"io"

	"github.com/unillm/unillm/pkg/openai"
)

// ChatProvider sends chat completion requests to an upstream provider.
type ChatProvider interface {
	Name() string
	ChatCompletion(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)
	ChatCompletionStream(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error)
}

// ProviderRegistry resolves upstream chat providers by name.
type ProviderRegistry interface {
	Get(name string) (ChatProvider, bool)
}

// BillingRecorder records usage for billing.
type BillingRecorder interface {
	RecordUsage(ctx context.Context, record UsageRecord) error
}

// UsageRecord mirrors core/billing.UsageRecord to avoid a core→core import cycle.
type UsageRecord struct {
	UserID           int64
	APIKeyID         int64
	ModelName        string
	ProviderName     string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Cost             float64
	Latency          float64
	Status           string
	HTTPStatus       int
	IsStream         bool
}

// MetricsRecorder records proxy metrics (optional).
type MetricsRecorder interface {
	RecordProxy(modelName, providerName, status string, latency float64, promptTok, completionTok int)
}

// EmbeddingForwarder sends embedding requests to upstream OpenAI-compatible endpoints.
type EmbeddingForwarder interface {
	Forward(ctx context.Context, baseURL, apiKey string, req *openai.EmbeddingRequest) (*openai.EmbeddingResponse, int, error)
}

// StreamWriter receives SSE lines for streaming chat completions.
type StreamWriter interface {
	WriteLine(line string) error
	Flush() error
}

// CallContext identifies the caller for billing.
type CallContext struct {
	UserID   int64
	APIKeyID int64
}

// ChatResult is the outcome of a non-streaming chat completion.
type ChatResult struct {
	Response *openai.ChatCompletionResponse
}

// StreamResult holds token usage after a streaming completion.
type StreamResult struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Cost             float64
	Latency          float64
}

// EmbeddingResult is the outcome of an embedding request.
type EmbeddingResult struct {
	Response *openai.EmbeddingResponse
	Cost     float64
	Latency  float64
}

// RouteResolver resolves public model names to routing metadata.
type RouteResolver interface {
	ResolveModel(ctx context.Context, publicName string) (ModelRoute, error)
	NextKey(ctx context.Context, providerName string) (string, error)
	FirstKey(ctx context.Context, providerName string) (string, error)
}

// ModelRoute contains resolved model and provider info for inference.
type ModelRoute struct {
	PublicName       string
	UpstreamModel    string
	ProviderName     string
	ProviderBaseURL  string
	InputPricePer1M  float64
	OutputPricePer1M float64
}

// Error codes returned by inference operations.
type ErrorKind string

const (
	ErrModelNotFound    ErrorKind = "model_not_found"
	ErrProviderMissing  ErrorKind = "provider_missing"
	ErrProviderKey      ErrorKind = "provider_key"
	ErrUpstream         ErrorKind = "upstream"
	ErrStreamUnsupported ErrorKind = "stream_unsupported"
)

// InferenceError carries a stable error classification for HTTP mapping.
type InferenceError struct {
	Kind       ErrorKind
	Message    string
	HTTPStatus int
	Cause      error
}

func (e *InferenceError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *InferenceError) Unwrap() error {
	return e.Cause
}
