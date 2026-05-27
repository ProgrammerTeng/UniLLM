package provider

import (
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

// GoogleProvider translates OpenAI-compatible requests to Google Gemini API.
// Note: Google also offers an OpenAI-compatible endpoint. This adapter supports
// both the native Gemini API and the OpenAI-compatible mode.
// For simplicity, we use the OpenAI-compatible endpoint when available.
type GoogleProvider struct {
	useNative    bool
	baseURL      string
	client       *http.Client
	streamClient *http.Client
}

func NewGoogleProvider(baseURL string) *GoogleProvider {
	return &GoogleProvider{
		useNative:    false,
		baseURL:      baseURL,
		client:       NewStandardClient(120 * time.Second),
		streamClient: NewStreamClient(),
	}
}

func (p *GoogleProvider) Name() string { return "google" }

// --- Google Gemini native types ---

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
	GenerationConfig *geminiGenerationConfig `json:"generationConfig,omitempty"`
	Tools            []geminiToolDecl       `json:"tools,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text         string                `json:"text,omitempty"`
	FunctionCall *geminiFunctionCall   `json:"functionCall,omitempty"`
	FunctionResp *geminiFunctionResp   `json:"functionResponse,omitempty"`
}

type geminiFunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}

type geminiFunctionResp struct {
	Name     string          `json:"name"`
	Response json.RawMessage `json:"response"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
}

type geminiToolDecl struct {
	FunctionDeclarations []geminiFunction `json:"functionDeclarations"`
}

type geminiFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata *geminiUsage      `json:"usageMetadata,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
	ThoughtsTokenCount   int `json:"thoughtsTokenCount,omitempty"`
}

// --- OpenAI-compatible mode (preferred) ---

func (p *GoogleProvider) ChatCompletion(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	if !p.useNative {
		return p.chatCompletionOpenAICompat(ctx, apiKey, req)
	}
	return p.chatCompletionNative(ctx, apiKey, req)
}

func (p *GoogleProvider) ChatCompletionStream(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error) {
	if !p.useNative {
		return p.chatCompletionStreamOpenAICompat(ctx, apiKey, req)
	}
	return p.chatCompletionStreamNative(ctx, apiKey, req)
}

// OpenAI-compatible mode: Google provides this at /v1beta/openai/
func (p *GoogleProvider) chatCompletionOpenAICompat(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	req.Stream = false
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.baseURL, "/") + "/openai/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result openai.ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

func (p *GoogleProvider) chatCompletionStreamOpenAICompat(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error) {
	req.Stream = true
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := strings.TrimSuffix(p.baseURL, "/") + "/openai/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("google error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return resp.Body, nil
}

// --- Native Gemini API mode (fallback) ---

func (p *GoogleProvider) convertToGemini(req *openai.ChatCompletionRequest) *geminiRequest {
	gr := &geminiRequest{
		GenerationConfig: &geminiGenerationConfig{
			MaxOutputTokens: req.MaxTokens,
			Temperature:     req.Temperature,
			TopP:            req.TopP,
		},
	}

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			gr.SystemInstruction = &geminiContent{
				Parts: []geminiPart{{Text: contentToString(msg.Content)}},
			}
			continue
		}

		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		gc := geminiContent{Role: role}
		text := contentToString(msg.Content)
		if text != "" {
			gc.Parts = append(gc.Parts, geminiPart{Text: text})
		}

		// Handle tool calls
		for _, tc := range msg.ToolCalls {
			gc.Parts = append(gc.Parts, geminiPart{
				FunctionCall: &geminiFunctionCall{
					Name: tc.Function.Name,
					Args: json.RawMessage(tc.Function.Arguments),
				},
			})
		}

		// Handle tool results
		if msg.Role == "tool" {
			gc.Role = "user"
			gc.Parts = []geminiPart{{
				FunctionResp: &geminiFunctionResp{
					Name:     msg.Name,
					Response: json.RawMessage(contentToString(msg.Content)),
				},
			}}
		}

		gr.Contents = append(gr.Contents, gc)
	}

	// Convert tools
	if len(req.Tools) > 0 {
		decl := geminiToolDecl{}
		for _, t := range req.Tools {
			decl.FunctionDeclarations = append(decl.FunctionDeclarations, geminiFunction{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			})
		}
		gr.Tools = []geminiToolDecl{decl}
	}

	return gr
}

func (p *GoogleProvider) convertFromGemini(gr *geminiResponse, model string) *openai.ChatCompletionResponse {
	resp := &openai.ChatCompletionResponse{
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
	}

	if gr.UsageMetadata != nil {
		resp.Usage = &openai.Usage{
			PromptTokens:     gr.UsageMetadata.PromptTokenCount,
			CompletionTokens: gr.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      gr.UsageMetadata.TotalTokenCount,
		}
		// Expose Gemini's reasoning (thinking) tokens — this is the key differentiator
		// that competitors don't provide, letting users see why Gemini costs are high
		if gr.UsageMetadata.ThoughtsTokenCount > 0 {
			resp.Usage.CompletionTokensDetails = &openai.CompletionTokensDetails{
				ReasoningTokens: gr.UsageMetadata.ThoughtsTokenCount,
			}
		}
	}

	for i, cand := range gr.Candidates {
		msg := &openai.Message{Role: "assistant"}
		var textParts []string
		var toolCalls []openai.ToolCall

		for _, part := range cand.Content.Parts {
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
			if part.FunctionCall != nil {
				toolCalls = append(toolCalls, openai.ToolCall{
					ID:   fmt.Sprintf("call_%d", len(toolCalls)),
					Type: "function",
					Function: openai.FunctionCall{
						Name:      part.FunctionCall.Name,
						Arguments: string(part.FunctionCall.Args),
					},
				})
			}
		}

		if len(textParts) > 0 {
			msg.Content = strings.Join(textParts, "")
		}
		msg.ToolCalls = toolCalls

		fr := "stop"
		switch cand.FinishReason {
		case "MAX_TOKENS":
			fr = "length"
		case "STOP":
			fr = "stop"
		}

		resp.Choices = append(resp.Choices, openai.Choice{
			Index:        i,
			Message:      msg,
			FinishReason: fr,
		})
	}

	return resp
}

func (p *GoogleProvider) chatCompletionNative(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	gr := p.convertToGemini(req)
	body, err := json.Marshal(gr)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, req.Model, apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return p.convertFromGemini(&geminiResp, req.Model), nil
}

func (p *GoogleProvider) chatCompletionStreamNative(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error) {
	gr := p.convertToGemini(req)
	body, err := json.Marshal(gr)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/models/%s:streamGenerateContent?alt=sse&key=%s", p.baseURL, req.Model, apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.streamClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("gemini error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return resp.Body, nil
}
