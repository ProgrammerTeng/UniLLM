package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "unillm_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "unillm_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10, 30, 60, 120},
	}, []string{"method", "path"})

	ProxyRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "unillm_proxy_requests_total",
		Help: "Total proxy requests by model and provider",
	}, []string{"model", "provider", "status"})

	ProxyDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "unillm_proxy_duration_seconds",
		Help:    "Proxy request duration by model",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
	}, []string{"model", "provider"})

	TokensTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "unillm_tokens_total",
		Help: "Total tokens processed",
	}, []string{"model", "type"})

	UpstreamErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "unillm_upstream_errors_total",
		Help: "Total upstream provider errors",
	}, []string{"provider"})

	ActiveConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "unillm_active_connections",
		Help: "Current active connections",
	})
)

// Metrics middleware records HTTP request metrics.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		ActiveConnections.Inc()
		start := time.Now()

		c.Next()

		ActiveConnections.Dec()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		RequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		RequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// RecordProxy records proxy-specific metrics.
func RecordProxy(model, provider, status string, duration float64, promptTokens, completionTokens int) {
	ProxyRequestsTotal.WithLabelValues(model, provider, status).Inc()
	ProxyDuration.WithLabelValues(model, provider).Observe(duration)
	if promptTokens > 0 {
		TokensTotal.WithLabelValues(model, "prompt").Add(float64(promptTokens))
	}
	if completionTokens > 0 {
		TokensTotal.WithLabelValues(model, "completion").Add(float64(completionTokens))
	}
	if status == "error" {
		UpstreamErrors.WithLabelValues(provider).Inc()
	}
}
