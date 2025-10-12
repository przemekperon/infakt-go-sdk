package infakt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultBaseURL   = "https://api.infakt.pl"
	defaultUserAgent = "golang-infakt"
	apiVersion       = "v3"
)

// Client manages communication with the inFakt API.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		u, err := url.Parse(baseURL)
		if err == nil {
			c.baseURL = u
		}
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithUserAgent sets a custom User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// NewClient creates a new inFakt API client.
func NewClient(apiKey string, opts ...Option) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
		userAgent:  defaultUserAgent,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// newRequest creates an API request.
func (c *Client) newRequest(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("infakt: invalid path %q: %w", path, err)
	}

	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("infakt: failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-inFakt-ApiKey", c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// do sends an API request and decodes the response.
func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infakt: request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return resp, err
	}

	if v != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, fmt.Errorf("infakt: failed to read response body: %w", err)
		}
		if err := json.Unmarshal(data, v); err != nil {
			return resp, fmt.Errorf("infakt: failed to decode response: %w", err)
		}
	}

	return resp, nil
}
