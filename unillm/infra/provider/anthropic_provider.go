package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/unillm/unillm/pkg/openai"
)

// AnthropicProvider translates OpenAI-compatible requests to Anthropic Messages API.
type AnthropicProvider struct {
	baseURL      string
	client       *http.Client
	streamClient *http.Client
}

func NewAnthropicProvider(baseURL string) *AnthropicProvider {
	return &AnthropicProvider{
		baseURL:      baseURL,
		client:       NewStandardClient(120 * time.Second),
		streamClient: NewStreamClient(),
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

// --- Anthropic API types ---

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream,omitempty"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []anthropicContent
}

type anthropicContent struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Source    *anthropicImage `json:"source,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
}

type anthropicImage struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	InputSchema interface{} `json:"input_schema"`
}

type anthropicResponse struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Role       string             `json:"role"`
	Content    []anthropicContent `json:"content"`
	Model      string             `json:"model"`
	StopReason string             `json:"stop_reason"`
	Usage      anthropicUsage     `json:"usage"`
}

type anthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens,omitempty"`
}

// --- Conversion: OpenAI → Anthropic ---

func (p *AnthropicProvider) convertRequest(req *openai.ChatCompletionRequest) *anthropicRequest {
	ar := &anthropicRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
	}
	if ar.MaxTokens == 0 {
		ar.MaxTokens = 4096
	}

	// Extract system message
	var messages []anthropicMessage
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			ar.System = contentToString(msg.Content)
			continue
		}

		am := anthropicMessage{Role: msg.Role}

		// Handle tool_calls in assistant messages
		if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
			parts := make([]anthropicContent, 0)
			text := contentToString(msg.Content)
			if text != "" {
				parts = append(parts, anthropicContent{Type: "text", Text: text})
			}
			for _, tc := range msg.ToolCalls {
				parts = append(parts, anthropicContent{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: json.RawMessage(tc.Function.Arguments),
				})
			}
			am.Content = parts
			messages = append(messages, am)
			continue
		}

		// Handle tool result messages
		if msg.Role == "tool" {
			am.Role = "user"
			am.Content = []anthropicContent{{
				Type:      "tool_result",
				ToolUseID: msg.ToolCallID,
				Content:   contentToString(msg.Content),
			}}
			messages = append(messages, am)
			continue
		}

		// Regular text message
		am.Content = contentToString(msg.Content)
		messages = append(messages, am)
	}
	ar.Messages = messages

	// Convert tools
	if len(req.Tools) > 0 {
		for _, t := range req.Tools {
			ar.Tools = append(ar.Tools, anthropicTool{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				InputSchema: t.Function.Parameters,
			})
		}
	}

	return ar
}

// --- Conversion: Anthropic → OpenAI ---

func (p *AnthropicProvider) convertResponse(ar *anthropicResponse, publicModel string) *openai.ChatCompletionResponse {
	resp := &openai.ChatCompletionResponse{
		ID:      ar.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   publicModel,
		Usage: &openai.Usage{
			PromptTokens:     ar.Usage.InputTokens,
			CompletionTokens: ar.Usage.OutputTokens,
			TotalTokens:      ar.Usage.InputTokens + ar.Usage.OutputTokens,
		},
	}
	// Anthropic doesn't separate reasoning tokens in the same way,
	// but we track cache usage for transparency
	if ar.Usage.CacheReadInputTokens > 0 || ar.Usage.CacheCreationInputTokens > 0 {
		resp.Usage.CompletionTokensDetails = &openai.CompletionTokensDetails{}
	}

	msg := &openai.Message{Role: "assistant"}
	var textParts []string
	var toolCalls []openai.ToolCall

	for _, c := range ar.Content {
		switch c.Type {
		case "text":
			textParts = append(textParts, c.Text)
		case "tool_use":
			toolCalls = append(toolCalls, openai.ToolCall{
				ID:   c.ID,
				Type: "function",
				Function: openai.FunctionCall{
					Name:      c.Name,
					Arguments: string(c.Input),
				},
			})
		}
	}

	if len(textParts) > 0 {
		combined := strings.Join(textParts, "")
		msg.Content = combined
	}
	if len(toolCalls) > 0 {
		msg.ToolCalls = toolCalls
	}

	finishReason := "stop"
	switch ar.StopReason {
	case "end_turn":
		finishReason = "stop"
	case "max_tokens":
		finishReason = "length"
	case "tool_use":
		finishReason = "tool_calls"
	}

	resp.Choices = []openai.Choice{{
		Index:        0,
		Message:      msg,
		FinishReason: finishReason,
	}}

	return resp
}

// --- API calls ---

func (p *AnthropicProvider) ChatCompletion(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	ar := p.convertRequest(req)
	ar.Stream = false

	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return p.convertResponse(&anthropicResp, req.Model), nil
}

func (p *AnthropicProvider) ChatCompletionStream(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error) {
	ar := p.convertRequest(req)
	ar.Stream = true

	body, err := json.Marshal(ar)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Wrap Anthropic SSE stream to convert to OpenAI SSE format
	return newAnthropicStreamAdapter(resp.Body, req.Model), nil
}

// --- SSE Stream Adapter: Anthropic → OpenAI format ---

type anthropicStreamAdapter struct {
	source  io.ReadCloser
	scanner *bufio.Scanner
	model   string
	buf     bytes.Buffer
	done    bool
}

func newAnthropicStreamAdapter(source io.ReadCloser, model string) *anthropicStreamAdapter {
	scanner := bufio.NewScanner(source)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 64KB initial, 1MB max
	return &anthropicStreamAdapter{source: source, scanner: scanner, model: model}
}

func (a *anthropicStreamAdapter) Read(p []byte) (int, error) {
	// If we have buffered data, return it first
	if a.buf.Len() > 0 {
		return a.buf.Read(p)
	}

	if a.done {
		return 0, io.EOF
	}

	// Read a line using buffered scanner (replaces single-byte reads)
	if !a.scanner.Scan() {
		a.done = true
		if err := a.scanner.Err(); err != nil {
			return 0, err
		}
		return 0, io.EOF
	}

	lineStr := strings.TrimSpace(a.scanner.Text())

	// Pass through empty lines
	if lineStr == "" {
		a.buf.WriteString("\n")
		return a.buf.Read(p)
	}

	// Handle data: prefix
	if strings.HasPrefix(lineStr, "data: ") {
		data := strings.TrimPrefix(lineStr, "data: ")

		// Parse Anthropic event
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			// Pass through as-is if can't parse
			a.buf.WriteString(lineStr + "\n")
			return a.buf.Read(p)
		}

		eventType, _ := event["type"].(string)
		switch eventType {
		case "content_block_delta":
			delta, _ := event["delta"].(map[string]interface{})
			deltaType, _ := delta["type"].(string)
			if deltaType == "text_delta" {
				text, _ := delta["text"].(string)
				chunk := openai.ChatCompletionResponse{
					Object: "chat.completion.chunk",
					Model:  a.model,
					Choices: []openai.Choice{{
						Index: 0,
						Delta: &openai.Message{Role: "", Content: text},
					}},
				}
				chunkJSON, _ := json.Marshal(chunk)
				a.buf.WriteString("data: " + string(chunkJSON) + "\n\n")
			}
		case "message_delta":
			delta, _ := event["delta"].(map[string]interface{})
			stopReason, _ := delta["stop_reason"].(string)
			fr := "stop"
			if stopReason == "max_tokens" {
				fr = "length"
			} else if stopReason == "tool_use" {
				fr = "tool_calls"
			}
			chunk := openai.ChatCompletionResponse{
				Object: "chat.completion.chunk",
				Model:  a.model,
				Choices: []openai.Choice{{
					Index:        0,
					Delta:        &openai.Message{},
					FinishReason: fr,
				}},
			}
			// Extract usage if present
			if usage, ok := event["usage"].(map[string]interface{}); ok {
				outTok, _ := usage["output_tokens"].(float64)
				chunk.Usage = &openai.Usage{CompletionTokens: int(outTok)}
			}
			chunkJSON, _ := json.Marshal(chunk)
			a.buf.WriteString("data: " + string(chunkJSON) + "\n\n")
		case "message_stop":
			a.buf.WriteString("data: [DONE]\n\n")
			a.done = true
		default:
			// Skip other event types (message_start, content_block_start, ping, etc.)
		}

		return a.buf.Read(p)
	}

	// Pass through event: lines
	if strings.HasPrefix(lineStr, "event:") {
		// Don't emit Anthropic event types to client
		a.buf.WriteString("")
		return a.buf.Read(p)
	}

	// Default: pass through
	a.buf.WriteString(lineStr + "\n")
	return a.buf.Read(p)
}

func (a *anthropicStreamAdapter) Close() error {
	return a.source.Close()
}

// --- Helpers ---

func contentToString(content interface{}) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		// Array of content parts
		var parts []string
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if t, ok := m["text"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, "")
	default:
		if v == nil {
			return ""
		}
		b, _ := json.Marshal(v)
		return string(b)
	}
}
