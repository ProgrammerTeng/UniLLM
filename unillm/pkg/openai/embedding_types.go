package openai

// EmbeddingRequest is the OpenAI-compatible embedding request format.
type EmbeddingRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"` // string or []string
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
}

// EmbeddingResponse is the OpenAI-compatible embedding response format.
type EmbeddingResponse struct {
	Object string          `json:"object"` // "list"
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  *EmbeddingUsage `json:"usage,omitempty"`
}

// EmbeddingData represents a single embedding vector.
type EmbeddingData struct {
	Object    string    `json:"object"` // "embedding"
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingUsage tracks token usage for embedding requests.
type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
