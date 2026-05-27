package catalog

import (
	"context"

	corecatalog "github.com/unillm/unillm/core/catalog"
	"github.com/unillm/unillm/infra/crypto"
	"github.com/unillm/unillm/infra/persistence"
	"github.com/unillm/unillm/internal/model"
)

// Repository adapts persistence to core/catalog.Repository.
type Repository struct {
	repo      *persistence.ProviderRepo
	protector *crypto.KeyProtector
}

func NewRepository(repo *persistence.ProviderRepo, protector *crypto.KeyProtector) *Repository {
	return &Repository{repo: repo, protector: protector}
}

func (r *Repository) ListActiveProviders(_ context.Context) ([]corecatalog.Provider, error) {
	providers, err := r.repo.ListActive()
	if err != nil {
		return nil, err
	}
	result := make([]corecatalog.Provider, len(providers))
	for i, p := range providers {
		result[i] = corecatalog.Provider{
			ID:      p.ID,
			Name:    p.Name,
			BaseURL: p.BaseURL,
		}
	}
	return result, nil
}

func (r *Repository) ListActiveKeys(_ context.Context, providerID int64) ([]string, error) {
	keys, err := r.repo.ListActiveKeys(providerID)
	if err != nil {
		return nil, err
	}
	values := make([]string, len(keys))
	for i, k := range keys {
		values[i] = r.protector.Decrypt(k.KeyValue)
	}
	return values, nil
}

func (r *Repository) FindModelByPublicName(_ context.Context, name string) (*corecatalog.ModelConfig, error) {
	m, err := r.repo.FindModelByPublicName(name)
	if err != nil {
		return nil, err
	}
	return toModelConfig(m), nil
}

func (r *Repository) FindProviderByID(_ context.Context, id int64) (*corecatalog.Provider, error) {
	p, err := r.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return &corecatalog.Provider{
		ID:      p.ID,
		Name:    p.Name,
		BaseURL: p.BaseURL,
	}, nil
}

func (r *Repository) ListActiveModels(_ context.Context) ([]corecatalog.ModelConfig, error) {
	models, err := r.repo.ListActiveModels()
	if err != nil {
		return nil, err
	}
	result := make([]corecatalog.ModelConfig, len(models))
	for i, m := range models {
		result[i] = *toModelConfig(&m)
	}
	return result, nil
}

func toModelConfig(m *model.ModelConfig) *corecatalog.ModelConfig {
	return &corecatalog.ModelConfig{
		ID:               m.ID,
		PublicName:       m.PublicName,
		ProviderID:       m.ProviderID,
		UpstreamModel:    m.UpstreamModel,
		InputPricePer1M:  m.InputPricePer1M,
		OutputPricePer1M: m.OutputPricePer1M,
	}
}
