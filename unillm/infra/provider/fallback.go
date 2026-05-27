package provider

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/unillm/unillm/pkg/openai"
)

// FallbackProvider tries multiple providers in order until one succeeds.
// This enables automatic failover: e.g., gpt-4o → claude-sonnet → gemini-flash.
type FallbackProvider struct {
	name      string
	providers []Provider
	apiKeys   map[string]string // provider name → API key
}

// FallbackConfig defines a fallback chain.
type FallbackConfig struct {
	Name      string            // name for this fallback group
	Providers []Provider        // ordered list of providers to try
	APIKeys   map[string]string // provider name → API key
}

// NewFallbackProvider creates a provider that falls through a chain on failure.
func NewFallbackProvider(cfg FallbackConfig) *FallbackProvider {
	return &FallbackProvider{
		name:      cfg.Name,
		providers: cfg.Providers,
		apiKeys:   cfg.APIKeys,
	}
}

func (p *FallbackProvider) Name() string {
	return p.name
}

func (p *FallbackProvider) ChatCompletion(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	var lastErr error
	for _, prov := range p.providers {
		// Use the provider's own API key if configured, otherwise use the passed key
		key := apiKey
		if k, ok := p.apiKeys[prov.Name()]; ok {
			key = k
		}

		resp, err := prov.ChatCompletion(ctx, key, req)
		if err == nil {
			log.Debug().
				Str("fallback_group", p.name).
				Str("provider_used", prov.Name()).
				Msg("fallback succeeded")
			return resp, nil
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("fallback_group", p.name).
			Str("provider_failed", prov.Name()).
			Msg("fallback trying next provider")
	}

	return nil, fmt.Errorf("all providers in fallback chain '%s' failed: %w", p.name, lastErr)
}

func (p *FallbackProvider) ChatCompletionStream(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error) {
	var lastErr error
	for _, prov := range p.providers {
		key := apiKey
		if k, ok := p.apiKeys[prov.Name()]; ok {
			key = k
		}

		stream, err := prov.ChatCompletionStream(ctx, key, req)
		if err == nil {
			log.Debug().
				Str("fallback_group", p.name).
				Str("provider_used", prov.Name()).
				Msg("fallback stream succeeded")
			return stream, nil
		}

		lastErr = err
		log.Warn().
			Err(err).
			Str("fallback_group", p.name).
			Str("provider_failed", prov.Name()).
			Msg("fallback stream trying next")
	}

	return nil, fmt.Errorf("all providers in fallback chain '%s' failed: %w", p.name, lastErr)
}
