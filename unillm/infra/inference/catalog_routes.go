package inference

import (
	"context"

	"github.com/unillm/unillm/core/catalog"
	coreinference "github.com/unillm/unillm/core/inference"
)

// CatalogRoutes adapts catalog.Service to inference.RouteResolver.
type CatalogRoutes struct {
	Catalog *catalog.Service
}

func (a *CatalogRoutes) ResolveModel(ctx context.Context, publicName string) (coreinference.ModelRoute, error) {
	route, err := a.Catalog.ResolveModel(ctx, publicName)
	if err != nil {
		return coreinference.ModelRoute{}, err
	}
	return coreinference.ModelRoute{
		PublicName:       route.Model.PublicName,
		UpstreamModel:    route.Model.UpstreamModel,
		ProviderName:     route.Provider.Name,
		ProviderBaseURL:  route.Provider.BaseURL,
		InputPricePer1M:  route.Model.InputPricePer1M,
		OutputPricePer1M: route.Model.OutputPricePer1M,
	}, nil
}

func (a *CatalogRoutes) NextKey(ctx context.Context, providerName string) (string, error) {
	return a.Catalog.NextKey(ctx, providerName)
}

func (a *CatalogRoutes) FirstKey(ctx context.Context, providerName string) (string, error) {
	return a.Catalog.FirstKey(ctx, providerName)
}
