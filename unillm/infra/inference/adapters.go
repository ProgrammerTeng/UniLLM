package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	coreinference "github.com/unillm/unillm/core/inference"
	"github.com/unillm/unillm/infra/provider"
	"github.com/unillm/unillm/pkg/openai"
)

// ProviderRegistry adapts provider.Registry to core/inference.ProviderRegistry.
type ProviderRegistry struct {
	Registry *provider.Registry
}

func (r *ProviderRegistry) Get(name string) (coreinference.ChatProvider, bool) {
	p, ok := r.Registry.Get(name)
	return p, ok
}

// HTTPEmbeddingForwarder forwards embedding requests over HTTP.
type HTTPEmbeddingForwarder struct {
	Client *http.Client
}

func NewHTTPEmbeddingForwarder(timeout time.Duration) *HTTPEmbeddingForwarder {
	return &HTTPEmbeddingForwarder{Client: provider.NewStandardClient(timeout)}
}

func (f *HTTPEmbeddingForwarder) Forward(ctx context.Context, baseURL, apiKey string, req *openai.EmbeddingRequest) (*openai.EmbeddingResponse, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}

	url := baseURL + "/embeddings"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := f.Client.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, resp.StatusCode, nil
	}

	var embResp openai.EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, resp.StatusCode, err
	}
	return &embResp, resp.StatusCode, nil
}

// MetricsAdapter wraps a metrics callback.
type MetricsAdapter struct {
	Record func(modelName, providerName, status string, latency float64, promptTok, completionTok int)
}

func (a *MetricsAdapter) RecordProxy(modelName, providerName, status string, latency float64, promptTok, completionTok int) {
	if a.Record != nil {
		a.Record(modelName, providerName, status, latency, promptTok, completionTok)
	}
}
