package catalog

// Provider describes an upstream AI provider.
type Provider struct {
	ID      int64
	Name    string
	BaseURL string
}

// ModelConfig maps a public model name to upstream routing and pricing.
type ModelConfig struct {
	ID               int64
	PublicName       string
	ProviderID       int64
	UpstreamModel    string
	InputPricePer1M  float64
	OutputPricePer1M float64
}

// ModelRoute is the resolved routing for a public model name.
type ModelRoute struct {
	Model    ModelConfig
	Provider Provider
}

// ModelInfo is a public-facing model listing entry.
type ModelInfo struct {
	ID         string
	Provider   string
	InputPrice float64
	OutputPrice float64
}
