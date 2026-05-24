package provider

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/unillm/unillm/infra/persistence"
	"github.com/unillm/unillm/internal/model"
)

// RegisterAll registers upstream providers with resilience and optional fallback chains.
//
// Fallback chains are configured via FALLBACK_CHAIN env var:
//   FALLBACK_CHAIN=openai:anthropic,google
// registers a fallback group named "openai" that tries anthropic then google.
func RegisterAll(registry *Registry, providerRepo *persistence.ProviderRepo) {
	providers, err := providerRepo.ListActive()
	if err != nil {
		log.Warn().Err(err).Msg("failed to load providers")
		return
	}

	byName := make(map[string]Provider)
	for _, p := range providers {
		inner := buildProvider(p)
		if inner == nil {
			continue
		}
		wrapped := NewResilientProvider(inner)
		byName[p.Name] = wrapped
		registry.Register(wrapped)
		log.Info().Str("provider", p.Name).Str("url", p.BaseURL).Msg("registered provider")
	}

	registerFallbackChains(registry, byName)
}

func buildProvider(p model.Provider) Provider {
	switch p.Name {
	case "openai", "deepseek", "alibaba", "bytedance", "geneasy":
		return NewOpenAIProvider(p.Name, p.BaseURL)
	case "anthropic":
		return NewAnthropicProvider(p.BaseURL)
	case "google":
		return NewGoogleProvider(p.BaseURL)
	default:
		log.Warn().Str("provider", p.Name).Msg("unknown provider")
		return nil
	}
}

func registerFallbackChains(registry *Registry, byName map[string]Provider) {
	chainSpec := os.Getenv("FALLBACK_CHAIN")
	if chainSpec == "" {
		return
	}

	parts := strings.Split(chainSpec, ":")
	if len(parts) < 2 {
		log.Warn().Str("spec", chainSpec).Msg("invalid FALLBACK_CHAIN, expected name:p1,p2")
		return
	}

	groupName := parts[0]
	names := strings.Split(parts[1], ",")
	var chain []Provider
	for _, name := range names {
		name = strings.TrimSpace(name)
		if p, ok := byName[name]; ok {
			chain = append(chain, p)
		}
	}
	if len(chain) == 0 {
		return
	}

	registry.Register(NewFallbackProvider(FallbackConfig{
		Name:      groupName,
		Providers: chain,
	}))
	log.Info().Str("group", groupName).Int("providers", len(chain)).Msg("registered fallback chain")
}
