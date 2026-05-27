package inference

import (
	"encoding/json"
	"strings"

	"github.com/unillm/unillm/pkg/openai"
)

func isReasoningModel(model string) bool {
	m := strings.ToLower(model)
	return strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4")
}

func calculateCost(inputPrice, outputPrice float64, usage *openai.Usage) float64 {
	if usage == nil {
		return 0
	}
	inputCost := float64(usage.PromptTokens) * inputPrice / 1_000_000
	outputCost := float64(usage.CompletionTokens) * outputPrice / 1_000_000
	return inputCost + outputCost
}

func calculateTokenCost(inputPrice, outputPrice float64, promptTok, completionTok int) float64 {
	return float64(promptTok)*inputPrice/1_000_000 + float64(completionTok)*outputPrice/1_000_000
}

func prepareChatRequest(publicModel string, req *openai.ChatCompletionRequest, route ModelRoute) {
	req.Model = route.UpstreamModel
	if isReasoningModel(publicModel) && req.MaxTokens > 0 {
		if req.MaxCompletionTokens == 0 {
			req.MaxCompletionTokens = req.MaxTokens
		}
		req.MaxTokens = 0
	}
	if req.Stream {
		if req.StreamOptions == nil {
			req.StreamOptions = &openai.StreamOptions{IncludeUsage: true}
		} else {
			req.StreamOptions.IncludeUsage = true
		}
	}
}

func replaceModelName(line, publicModel string) string {
	const prefix = `"model":"`
	idx := strings.Index(line, prefix)
	if idx < 0 {
		return line
	}
	start := idx + len(prefix)
	end := strings.Index(line[start:], `"`)
	if end < 0 {
		return line
	}
	return line[:start] + publicModel + line[start+end:]
}

func extractContentLength(line string) int {
	data := strings.TrimPrefix(line, "data: ")
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err == nil && len(chunk.Choices) > 0 {
		return len(chunk.Choices[0].Delta.Content)
	}
	return 0
}

func extractStreamUsage(line string) (pt, ct, tt int) {
	data := strings.TrimPrefix(line, "data: ")
	var chunk struct {
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err == nil && chunk.Usage != nil {
		return chunk.Usage.PromptTokens, chunk.Usage.CompletionTokens, chunk.Usage.TotalTokens
	}
	return 0, 0, 0
}

func estimatePromptTokens(messages []openai.Message) int {
	total := 0
	for _, m := range messages {
		switch v := m.Content.(type) {
		case string:
			total += len(v)
		default:
			b, _ := json.Marshal(v)
			total += len(b)
		}
		total += 4
	}
	return max(1, total/4)
}
