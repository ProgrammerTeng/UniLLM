package openai

// ChatCompletionRequest is the OpenAI-compatible request format.
type ChatCompletionRequest struct {
	Model            string         `json:"model"`
	Messages         []Message      `json:"messages"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	MaxCompletionTokens int         `json:"max_completion_tokens,omitempty"`
	Temperature      *float64       `json:"temperature,omitempty"`
	TopP             *float64       `json:"top_p,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	StreamOptions    *StreamOptions `json:"stream_options,omitempty"`
	Tools            []Tool         `json:"tools,omitempty"`
	Stop             interface{}    `json:"stop,omitempty"`
	N                int            `json:"n,omitempty"`
	ReasoningEffort  string         `json:"reasoning_effort,omitempty"`
}

// StreamOptions controls stream behavior.
type StreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type Message struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"` // string or []ContentPart
	Name       string      `json:"name,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
}

type ContentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Function struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatCompletionResponse is the OpenAI-compatible response format.
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

type Choice struct {
	Index        int      `json:"index"`
	Message      *Message `json:"message,omitempty"`
	Delta        *Message `json:"delta,omitempty"`
	FinishReason string   `json:"finish_reason,omitempty"`
}

type Usage struct {
	PromptTokens            int                      `json:"prompt_tokens"`
	CompletionTokens        int                      `json:"completion_tokens"`
	TotalTokens             int                      `json:"total_tokens"`
	CompletionTokensDetails *CompletionTokensDetails `json:"completion_tokens_details,omitempty"`
}

// CompletionTokensDetails breaks down completion token usage.
type CompletionTokensDetails struct {
	ReasoningTokens           int `json:"reasoning_tokens,omitempty"`
	AcceptedPredictionTokens  int `json:"accepted_prediction_tokens,omitempty"`
	RejectedPredictionTokens  int `json:"rejected_prediction_tokens,omitempty"`
}

// ErrorResponse is the OpenAI-compatible error format.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}
