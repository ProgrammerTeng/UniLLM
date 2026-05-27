package provider

import (
	"context"
	"fmt"
	"io"

	"github.com/unillm/unillm/pkg/openai"
	"golang.org/x/sync/semaphore"
)

// ConcurrencyLimitedProvider wraps a Provider with a semaphore-based
// concurrency limiter. When the limit is reached, requests fail fast
// with a 429-style error instead of queuing and causing tail latency blowup.
type ConcurrencyLimitedProvider struct {
	inner Provider
	sem   *semaphore.Weighted
	limit int64
}

// NewConcurrencyLimitedProvider wraps a provider with a max concurrent requests limit.
func NewConcurrencyLimitedProvider(inner Provider, maxConcurrent int64) *ConcurrencyLimitedProvider {
	return &ConcurrencyLimitedProvider{
		inner: inner,
		sem:   semaphore.NewWeighted(maxConcurrent),
		limit: maxConcurrent,
	}
}

func (p *ConcurrencyLimitedProvider) Name() string {
	return p.inner.Name()
}

func (p *ConcurrencyLimitedProvider) ChatCompletion(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	if !p.sem.TryAcquire(1) {
		return nil, fmt.Errorf("provider %s at capacity (%d concurrent), try again later", p.inner.Name(), p.limit)
	}
	defer p.sem.Release(1)
	return p.inner.ChatCompletion(ctx, apiKey, req)
}

func (p *ConcurrencyLimitedProvider) ChatCompletionStream(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error) {
	if !p.sem.TryAcquire(1) {
		return nil, fmt.Errorf("provider %s at capacity (%d concurrent), try again later", p.inner.Name(), p.limit)
	}
	// Note: we release in a wrapper that tracks stream close
	stream, err := p.inner.ChatCompletionStream(ctx, apiKey, req)
	if err != nil {
		p.sem.Release(1)
		return nil, err
	}
	return &semaphoreReleasingReader{ReadCloser: stream, sem: p.sem}, nil
}

// semaphoreReleasingReader releases the semaphore when the stream is closed.
type semaphoreReleasingReader struct {
	io.ReadCloser
	sem      *semaphore.Weighted
	released bool
}

func (r *semaphoreReleasingReader) Close() error {
	err := r.ReadCloser.Close()
	if !r.released {
		r.sem.Release(1)
		r.released = true
	}
	return err
}
