package provider

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/unillm/unillm/pkg/openai"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // normal operation
	CircuitOpen                         // failing, reject requests
	CircuitHalfOpen                     // testing with a single request
)

// CircuitBreaker tracks failures for a provider.
type CircuitBreaker struct {
	mu             sync.Mutex
	state          CircuitState
	failures       int
	threshold      int           // consecutive failures to open
	resetTimeout   time.Duration // how long to stay open before half-open
	lastFailure    time.Time
}

func newCircuitBreaker(threshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        CircuitClosed,
		threshold:    threshold,
		resetTimeout: resetTimeout,
	}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = CircuitClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.threshold {
		cb.state = CircuitOpen
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// ResilientProvider wraps a Provider with retry and circuit breaker logic.
type ResilientProvider struct {
	inner   Provider
	cb      *CircuitBreaker
	retries int
}

// NewResilientProvider wraps a provider with retry (max 2) and circuit breaker (5 failures, 30s reset).
func NewResilientProvider(inner Provider) *ResilientProvider {
	return &ResilientProvider{
		inner:   inner,
		cb:      newCircuitBreaker(5, 30*time.Second),
		retries: 2,
	}
}

func (p *ResilientProvider) Name() string {
	return p.inner.Name()
}

func (p *ResilientProvider) ChatCompletion(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	if !p.cb.Allow() {
		return nil, fmt.Errorf("circuit breaker open for provider %s", p.inner.Name())
	}

	var lastErr error
	for attempt := 0; attempt <= p.retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 500ms, 1s
			backoff := time.Duration(attempt) * 500 * time.Millisecond
			log.Warn().
				Str("provider", p.inner.Name()).
				Int("attempt", attempt+1).
				Dur("backoff", backoff).
				Msg("retrying request")
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		resp, err := p.inner.ChatCompletion(ctx, apiKey, req)
		if err == nil {
			// Validate output is not empty (DeepSeek sometimes returns 200 with empty content)
			if isEmptyOutput(resp) {
				log.Warn().
					Str("provider", p.inner.Name()).
					Int("attempt", attempt+1).
					Msg("empty output detected, retrying")
				lastErr = fmt.Errorf("empty output from provider %s", p.inner.Name())
				continue
			}
			p.cb.RecordSuccess()
			return resp, nil
		}
		lastErr = err
		log.Warn().
			Err(err).
			Str("provider", p.inner.Name()).
			Int("attempt", attempt+1).
			Msg("request failed")
	}

	p.cb.RecordFailure()
	return nil, lastErr
}

func (p *ResilientProvider) ChatCompletionStream(ctx context.Context, apiKey string, req *openai.ChatCompletionRequest) (io.ReadCloser, error) {
	// Streaming: no retry (can't replay partial data), but respect circuit breaker
	if !p.cb.Allow() {
		return nil, fmt.Errorf("circuit breaker open for provider %s", p.inner.Name())
	}

	stream, err := p.inner.ChatCompletionStream(ctx, apiKey, req)
	if err != nil {
		p.cb.RecordFailure()
		return nil, err
	}

	p.cb.RecordSuccess()
	return stream, nil
}

// CircuitBreakerState returns the circuit state for monitoring.
func (p *ResilientProvider) CircuitBreakerState() CircuitState {
	return p.cb.State()
}

// Inner returns the wrapped provider.
func (p *ResilientProvider) Inner() Provider {
	return p.inner
}

// isEmptyOutput returns true if the response has no content and no tool calls.
// This catches the case where upstream returns HTTP 200 but empty message body,
// which happens intermittently with DeepSeek and some reasoning models.
func isEmptyOutput(resp *openai.ChatCompletionResponse) bool {
	if resp == nil || len(resp.Choices) == 0 {
		return true
	}
	msg := resp.Choices[0].Message
	if msg == nil {
		return true
	}
	// If tool calls are present, content may legitimately be empty
	if len(msg.ToolCalls) > 0 {
		return false
	}
	content, _ := msg.Content.(string)
	return content == ""
}
