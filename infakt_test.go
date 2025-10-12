package infakt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestNewRequest(t *testing.T) {
	c := NewClient("my-key")

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
	c := NewClient("key")

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
		w.Write([]byte(`{"name":"test"}`))
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))

	req, err := c.newRequest(context.Background(), http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	_, err = c.do(req, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("expected name %q, got %q", "test", result["name"])
	}
}
