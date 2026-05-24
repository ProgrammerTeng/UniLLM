package catalog

import "context"

// Repository loads provider and model configuration.
type Repository interface {
	ListActiveProviders(ctx context.Context) ([]Provider, error)
	ListActiveKeys(ctx context.Context, providerID int64) ([]string, error)
	FindModelByPublicName(ctx context.Context, name string) (*ModelConfig, error)
	FindProviderByID(ctx context.Context, id int64) (*Provider, error)
	ListActiveModels(ctx context.Context) ([]ModelConfig, error)
}
