package catalog

import (
	"context"
	"fmt"
	"sync"
)

// Service resolves models and manages the in-memory provider key pool.
type Service struct {
	repo Repository

	mu           sync.Mutex
	providerKeys map[string][]string
	keyIndex     map[string]uint64
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:         repo,
		providerKeys: make(map[string][]string),
		keyIndex:     make(map[string]uint64),
	}
}

// Reload loads active provider keys from persistent storage.
func (s *Service) Reload(ctx context.Context) error {
	providers, err := s.repo.ListActiveProviders(ctx)
	if err != nil {
		return err
	}

	keys := make(map[string][]string)
	for _, p := range providers {
		providerKeys, err := s.repo.ListActiveKeys(ctx, p.ID)
		if err != nil {
			continue
		}
		if len(providerKeys) > 0 {
			keys[p.Name] = append(keys[p.Name], providerKeys...)
		}
	}

	s.mu.Lock()
	s.providerKeys = keys
	s.keyIndex = make(map[string]uint64)
	s.mu.Unlock()
	return nil
}

// ResolveModel returns routing information for a public model name.
func (s *Service) ResolveModel(ctx context.Context, publicName string) (*ModelRoute, error) {
	modelCfg, err := s.repo.FindModelByPublicName(ctx, publicName)
	if err != nil {
		return nil, err
	}

	providerInfo, err := s.repo.FindProviderByID(ctx, modelCfg.ProviderID)
	if err != nil {
		return nil, err
	}

	return &ModelRoute{
		Model:    *modelCfg,
		Provider: *providerInfo,
	}, nil
}

// ListPublicModels returns active models for catalog endpoints.
func (s *Service) ListPublicModels(ctx context.Context) ([]ModelInfo, error) {
	models, err := s.repo.ListActiveModels(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ModelInfo, 0, len(models))
	for _, m := range models {
		providerInfo, err := s.repo.FindProviderByID(ctx, m.ProviderID)
		if err != nil {
			continue
		}
		result = append(result, ModelInfo{
			ID:          m.PublicName,
			Provider:    providerInfo.Name,
			InputPrice:  m.InputPricePer1M,
			OutputPrice: m.OutputPricePer1M,
		})
	}
	return result, nil
}

// ListActiveModels returns raw model configs (used by health probes).
func (s *Service) ListActiveModels(ctx context.Context) ([]ModelConfig, error) {
	return s.repo.ListActiveModels(ctx)
}

// ListActiveProviders returns all active upstream providers.
func (s *Service) ListActiveProviders(ctx context.Context) ([]Provider, error) {
	return s.repo.ListActiveProviders(ctx)
}

// NextKey returns the next API key for a provider using round-robin.
func (s *Service) NextKey(_ context.Context, providerName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys := s.providerKeys[providerName]
	if len(keys) == 0 {
		return "", fmt.Errorf("no API keys configured for provider %s", providerName)
	}
	idx := s.keyIndex[providerName] % uint64(len(keys))
	s.keyIndex[providerName]++
	return keys[idx], nil
}

// FirstKey returns the first configured key (for embedding calls).
func (s *Service) FirstKey(_ context.Context, providerName string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys := s.providerKeys[providerName]
	if len(keys) == 0 {
		return "", fmt.Errorf("no API keys configured for provider %s", providerName)
	}
	return keys[0], nil
}
