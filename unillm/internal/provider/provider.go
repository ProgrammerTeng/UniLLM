package provider

import (
	"context"
	"io"

	"github.com/unillm/unillm/pkg/openai"
)

// Provider is the interface all upstream AI providers must implement.
type Provider interface {
	// Name returns the provider identifier (e.g. "openai", "anthropic").
	Name() string

	// ChatCompletion sends a non-streaming chat completion request.
	ChatCompletion(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error)

	// ChatCompletionStream sends a streaming chat completion request.
	// Returns a reader that yields SSE-formatted data.
	ChatCompletionStream(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error)
}

// Registry holds all registered providers.
type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
