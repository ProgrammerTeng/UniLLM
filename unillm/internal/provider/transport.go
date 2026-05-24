package provider

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/http2"
)

// SharedTransport returns a high-performance HTTP transport with connection
// pooling and HTTP/2 support. This eliminates repeated TCP+TLS handshakes
// that add 200-600ms per request to upstream LLM APIs.
func SharedTransport() *http.Transport {
	t := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 50,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		ForceAttemptHTTP2:   true,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	// Explicitly configure HTTP/2
	http2.ConfigureTransport(t)
	return t
}

// NewStandardClient returns an http.Client for non-streaming requests with
// connection pooling, HTTP/2, and a per-request timeout.
func NewStandardClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Transport: SharedTransport(),
		Timeout:   timeout,
	}
}

// NewStreamClient returns an http.Client for streaming (SSE) requests.
// No overall timeout (streams can run for minutes), but with a header timeout
// to detect unresponsive upstreams quickly.
func NewStreamClient() *http.Client {
	t := SharedTransport()
	t.ResponseHeaderTimeout = 30 * time.Second
	return &http.Client{
		Transport: t,
		// No Timeout — streams are long-lived
	}
}
