package infakt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// newTestClient creates a client with rate limiting disabled for faster tests.
func newTestClient(apiKey string, opts ...Option) *Client {
	opts = append([]Option{WithRateLimit(0)}, opts...)
	return NewClient(apiKey, opts...)
}

func TestNewClient(t *testing.T) {
	c := NewClient("test-api-key")

	if c.apiKey != "test-api-key" {
		t.Errorf("expected apiKey %q, got %q", "test-api-key", c.apiKey)
	}
	if c.baseURL.String() != defaultBaseURL {
		t.Errorf("expected baseURL %q, got %q", defaultBaseURL, c.baseURL.String())
	}
	if c.userAgent != defaultUserAgent {
		t.Errorf("expected userAgent %q, got %q", defaultUserAgent, c.userAgent)
	}
	if c.rateLimiter == nil {
		t.Error("expected default rate limiter to be set")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	customHTTP := &http.Client{}
	c := NewClient("key",
		WithBaseURL("https://custom.example.com"),
		WithHTTPClient(customHTTP),
		WithUserAgent("custom-agent"),
	)

	if c.baseURL.String() != "https://custom.example.com" {
		t.Errorf("expected baseURL %q, got %q", "https://custom.example.com", c.baseURL.String())
	}
	if c.httpClient != customHTTP {
		t.Error("expected custom HTTP client")
	}
	if c.userAgent != "custom-agent" {
		t.Errorf("expected userAgent %q, got %q", "custom-agent", c.userAgent)
	}
}

func TestWithRateLimit(t *testing.T) {
	c := NewClient("key", WithRateLimit(500*time.Millisecond))
	if c.rateLimiter == nil {
		t.Error("expected rate limiter to be set")
	}

	c2 := NewClient("key", WithRateLimit(0))
	if c2.rateLimiter != nil {
		t.Error("expected rate limiter to be nil when interval is 0")
	}
}

func TestNewRequest(t *testing.T) {
	c := newTestClient("my-key")

	req, err := c.newRequest(context.Background(), http.MethodGet, "/v3/invoices.json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := req.Header.Get("X-inFakt-ApiKey"); got != "my-key" {
		t.Errorf("expected X-inFakt-ApiKey %q, got %q", "my-key", got)
	}
	if got := req.Header.Get("Accept"); got != "application/json" {
		t.Errorf("expected Accept %q, got %q", "application/json", got)
	}
	if got := req.Header.Get("User-Agent"); got != defaultUserAgent {
		t.Errorf("expected User-Agent %q, got %q", defaultUserAgent, got)
	}
}

func TestNewRequestWithBody(t *testing.T) {
	c := newTestClient("key")

	body := map[string]string{"name": "test"}
	req, err := c.newRequest(context.Background(), http.MethodPost, "/v3/clients.json", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", got)
	}
}

func TestDo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"name":"test"}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))

	req, err := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	err = c.do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("expected name %q, got %q", "test", result["name"])
	}
}

func TestDo_RetryOn429(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))

	req, err := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]bool
	err = c.do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("expected 2 attempts, got %d", atomic.LoadInt32(&attempts))
	}
	if !result["ok"] {
		t.Error("expected ok=true in response")
	}
}

func TestDo_RetryOnServerError(t *testing.T) {
	var attempts int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"internal server error"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"recovered":true}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))

	req, err := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]bool
	err = c.do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", atomic.LoadInt32(&attempts))
	}
	if !result["recovered"] {
		t.Error("expected recovered=true in response")
	}
}

func TestErrorResponse_IncludesContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	req, _ := c.newRequest(context.Background(), http.MethodGet, "/v3/invoices/999.json", nil)
	err := c.do(req, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "GET") {
		t.Errorf("expected error to contain method, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "/v3/invoices/999.json") {
		t.Errorf("expected error to contain endpoint, got: %s", errMsg)
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		val      string
		expected time.Duration
	}{
		{"empty string", "", time.Second},
		{"numeric seconds", "5", 5 * time.Second},
		{"invalid value", "abc", time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.val)
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
