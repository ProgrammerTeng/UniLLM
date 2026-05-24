package inference

import (
	"testing"

	"github.com/unillm/unillm/pkg/openai"
)

func TestIsReasoningModel(t *testing.T) {
	cases := map[string]bool{
		"o1-preview":  true,
		"o3-mini":     true,
		"O4-mini":     true,
		"gpt-4o":      false,
		"claude-3":    false,
	}
	for model, want := range cases {
		if got := isReasoningModel(model); got != want {
			t.Fatalf("isReasoningModel(%q) = %v, want %v", model, got, want)
		}
	}
}

func TestCalculateCost(t *testing.T) {
	cost := calculateCost(3, 6, &openai.Usage{
		PromptTokens:     1_000_000,
		CompletionTokens: 500_000,
	})
	if cost != 6 {
		t.Fatalf("expected cost 6, got %v", cost)
	}
}

func TestEstimatePromptTokens(t *testing.T) {
	tokens := estimatePromptTokens([]openai.Message{
		{Role: "user", Content: "hello world"},
	})
	if tokens < 1 {
		t.Fatalf("expected at least 1 token, got %d", tokens)
	}
}

func TestReplaceModelName(t *testing.T) {
	line := `data: {"model":"gpt-4o-mini","choices":[]}`
	got := replaceModelName(line, "public-model")
	want := `data: {"model":"public-model","choices":[]}`
	if got != want {
		t.Fatalf("replaceModelName() = %q, want %q", got, want)
	}
}
