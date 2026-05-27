package billing

// UsageRecord captures one API call for billing and analytics.
type UsageRecord struct {
	UserID           int64
	APIKeyID         int64
	ModelName        string
	ProviderName     string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Cost             float64
	Latency          float64
	Status           string
	HTTPStatus       int
	IsStream         bool
}
