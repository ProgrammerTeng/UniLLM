package inference

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/unillm/unillm/pkg/openai"
)

// Service orchestrates chat and embedding inference requests.
type Service struct {
	routes    RouteResolver
	registry  ProviderRegistry
	billing   BillingRecorder
	metrics   MetricsRecorder
	embedding EmbeddingForwarder
}

func NewService(
	routes RouteResolver,
	registry ProviderRegistry,
	billing BillingRecorder,
	metrics MetricsRecorder,
	embedding EmbeddingForwarder,
) *Service {
	return &Service{
		routes:    routes,
		registry:  registry,
		billing:   billing,
		metrics:   metrics,
		embedding: embedding,
	}
}

// ChatCompletion handles a non-streaming chat completion.
func (s *Service) ChatCompletion(ctx context.Context, call CallContext, req *openai.ChatCompletionRequest) (*ChatResult, error) {
	publicModel := req.Model
	route, err := s.resolveRoute(ctx, publicModel)
	if err != nil {
		return nil, err
	}

	p, err := s.getProvider(route.ProviderName)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.routes.NextKey(ctx, route.ProviderName)
	if err != nil {
		return nil, &InferenceError{Kind: ErrProviderKey, Message: err.Error(), Cause: err}
	}

	prepareChatRequest(publicModel, req, route)
	start := time.Now()

	resp, err := p.ChatCompletion(ctx, apiKey, req)
	latency := time.Since(start).Seconds()
	if err != nil {
		s.recordUsage(call, publicModel, route.ProviderName, 0, 0, 0, 0, latency, "error", 502, false)
		return nil, &InferenceError{Kind: ErrUpstream, Message: "upstream error", Cause: err}
	}

	resp.Model = publicModel
	cost := calculateCost(route.InputPricePer1M, route.OutputPricePer1M, resp.Usage)

	pt, ct, tt := 0, 0, 0
	if resp.Usage != nil {
		pt = resp.Usage.PromptTokens
		ct = resp.Usage.CompletionTokens
		tt = resp.Usage.TotalTokens
	}
	s.recordUsage(call, publicModel, route.ProviderName, pt, ct, tt, cost, latency, "ok", 200, false)

	return &ChatResult{Response: resp}, nil
}

// ChatCompletionStream handles a streaming chat completion.
func (s *Service) ChatCompletionStream(ctx context.Context, call CallContext, req *openai.ChatCompletionRequest, w StreamWriter) (*StreamResult, error) {
	publicModel := req.Model
	route, err := s.resolveRoute(ctx, publicModel)
	if err != nil {
		return nil, err
	}

	p, err := s.getProvider(route.ProviderName)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.routes.NextKey(ctx, route.ProviderName)
	if err != nil {
		return nil, &InferenceError{Kind: ErrProviderKey, Message: err.Error(), Cause: err}
	}

	prepareChatRequest(publicModel, req, route)
	start := time.Now()

	stream, err := p.ChatCompletionStream(ctx, apiKey, req)
	if err != nil {
		latency := time.Since(start).Seconds()
		s.recordUsage(call, publicModel, route.ProviderName, 0, 0, 0, 0, latency, "error", 502, true)
		return nil, &InferenceError{Kind: ErrUpstream, Message: "upstream error", Cause: err}
	}
	defer stream.Close()

	var streamPT, streamCT, streamTT int
	var totalContentChars int
	scanner := bufio.NewScanner(stream)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") && line != "data: [DONE]" {
			line = replaceModelName(line, publicModel)
			if strings.Contains(line, `"usage"`) {
				streamPT, streamCT, streamTT = extractStreamUsage(line)
			}
			totalContentChars += extractContentLength(line)
		}
		if err := w.WriteLine(line); err != nil {
			break
		}
		if err := w.Flush(); err != nil {
			break
		}
	}

	latency := time.Since(start).Seconds()
	if streamPT == 0 && streamCT == 0 {
		streamPT = estimatePromptTokens(req.Messages)
		streamCT = max(1, totalContentChars/4)
		streamTT = streamPT + streamCT
	}

	cost := calculateTokenCost(route.InputPricePer1M, route.OutputPricePer1M, streamPT, streamCT)
	s.recordUsage(call, publicModel, route.ProviderName, streamPT, streamCT, streamTT, cost, latency, "ok", 200, true)

	return &StreamResult{
		PromptTokens:     streamPT,
		CompletionTokens: streamCT,
		TotalTokens:      streamTT,
		Cost:             cost,
		Latency:          latency,
	}, nil
}

// CreateEmbedding handles an embedding request.
func (s *Service) CreateEmbedding(ctx context.Context, call CallContext, req *openai.EmbeddingRequest) (*EmbeddingResult, error) {
	publicModel := req.Model
	route, err := s.resolveRoute(ctx, publicModel)
	if err != nil {
		return nil, err
	}

	apiKey, err := s.routes.FirstKey(ctx, route.ProviderName)
	if err != nil {
		return nil, &InferenceError{Kind: ErrProviderKey, Message: err.Error(), Cause: err}
	}

	upstreamReq := *req
	upstreamReq.Model = route.UpstreamModel

	start := time.Now()
	resp, statusCode, err := s.embedding.Forward(ctx, route.ProviderBaseURL, apiKey, &upstreamReq)
	latency := time.Since(start).Seconds()
	if err != nil {
		return nil, &InferenceError{Kind: ErrUpstream, Message: err.Error(), Cause: err}
	}
	if statusCode != 200 {
		return nil, &InferenceError{
			Kind:       ErrUpstream,
			Message:    fmt.Sprintf("upstream error (status %d)", statusCode),
			HTTPStatus: statusCode,
		}
	}

	resp.Model = publicModel
	pt := 0
	if resp.Usage != nil {
		pt = resp.Usage.PromptTokens
	}
	cost := float64(pt) * route.InputPricePer1M / 1_000_000

	go func() {
		_ = s.billing.RecordUsage(context.Background(), UsageRecord{
			UserID:       call.UserID,
			APIKeyID:     call.APIKeyID,
			ModelName:    publicModel,
			ProviderName: route.ProviderName,
			PromptTokens: pt,
			TotalTokens:  pt,
			Cost:         cost,
			Latency:      latency,
			Status:       "ok",
			HTTPStatus:   200,
		})
	}()

	return &EmbeddingResult{
		Response: resp,
		Cost:     cost,
		Latency:  latency,
	}, nil
}

func (s *Service) resolveRoute(ctx context.Context, publicName string) (ModelRoute, error) {
	route, err := s.routes.ResolveModel(ctx, publicName)
	if err != nil {
		return ModelRoute{}, &InferenceError{
			Kind:    ErrModelNotFound,
			Message: fmt.Sprintf("model '%s' not found or not available", publicName),
			Cause:   err,
		}
	}
	return route, nil
}

func (s *Service) getProvider(name string) (ChatProvider, error) {
	p, ok := s.registry.Get(name)
	if !ok {
		return nil, &InferenceError{Kind: ErrProviderMissing, Message: "provider not configured"}
	}
	return p, nil
}

func (s *Service) recordUsage(call CallContext, modelName, providerName string,
	promptTok, completionTok, totalTok int, cost, latency float64, status string, httpStatus int, isStream bool) {

	if s.metrics != nil {
		s.metrics.RecordProxy(modelName, providerName, status, latency, promptTok, completionTok)
	}

	go func() {
		_ = s.billing.RecordUsage(context.Background(), UsageRecord{
			UserID:           call.UserID,
			APIKeyID:         call.APIKeyID,
			ModelName:        modelName,
			ProviderName:     providerName,
			PromptTokens:     promptTok,
			CompletionTokens: completionTok,
			TotalTokens:      totalTok,
			Cost:             cost,
			Latency:          latency,
			Status:           status,
			HTTPStatus:       httpStatus,
			IsStream:         isStream,
		})
	}()
}
