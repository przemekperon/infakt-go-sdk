package infakt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL     = "https://api.infakt.pl"
	defaultUserAgent   = "golang-infakt"
	defaultHTTPTimeout = 30 * time.Second
	apiVersion         = "v3"

	defaultRateLimit  = 100 * time.Millisecond
	maxRetryAfterWait = 60 * time.Second
	maxServerRetries  = 3
	initialRetryWait  = 500 * time.Millisecond
)

// Client manages communication with the inFakt API.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string

	rateLimiter *time.Ticker

	Invoices     *InvoiceService
	Clients      *ClientEntityService
	Products     *ProductService
	BankAccounts *BankAccountService
	VatRates     *VatRateService
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

// WithRateLimit sets the minimum interval between API requests.
// The default is 100ms between requests. Set to 0 to disable.
func WithRateLimit(interval time.Duration) Option {
	return func(c *Client) {
		if c.rateLimiter != nil {
			c.rateLimiter.Stop()
		}
		if interval > 0 {
			c.rateLimiter = time.NewTicker(interval)
		} else {
			c.rateLimiter = nil
		}
	}
}

// NewClient creates a new inFakt API client.
func NewClient(apiKey string, opts ...Option) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		httpClient:  &http.Client{Timeout: defaultHTTPTimeout},
		userAgent:   defaultUserAgent,
		rateLimiter: time.NewTicker(defaultRateLimit),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Invoices = &InvoiceService{client: c}
	c.Clients = &ClientEntityService{client: c}
	c.Products = &ProductService{client: c}
	c.BankAccounts = &BankAccountService{client: c}
	c.VatRates = &VatRateService{client: c}

	return c
}

// Close releases resources associated with the Client.
// It stops the internal rate limiter ticker. After Close,
// the Client should not be used.
func (c *Client) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
		c.rateLimiter = nil
	}
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

// do sends an API request and decodes the JSON response into v.
// It applies rate limiting, handles HTTP 429 (Too Many Requests) responses
// by waiting for the duration specified in the Retry-After header, and retries
// on server errors (5xx) with exponential backoff (max 3 attempts).
func (c *Client) do(req *http.Request, v interface{}) error {
	if c.rateLimiter != nil {
		<-c.rateLimiter.C
	}

	resp, err := c.executeWithRetry(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return err
	}

	if v != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("infakt: %s %s: failed to read response body: %w",
				req.Method, req.URL.Path, err)
		}
		if err := json.Unmarshal(data, v); err != nil {
			return fmt.Errorf("infakt: %s %s: failed to decode response: %w",
				req.Method, req.URL.Path, err)
		}
	}

	return nil
}

// doRaw sends an API request and returns the raw response body bytes.
// Used for non-JSON responses like PDF downloads.
func (c *Client) doRaw(req *http.Request) ([]byte, error) {
	if c.rateLimiter != nil {
		<-c.rateLimiter.C
	}

	resp, err := c.executeWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("infakt: %s %s: failed to read response body: %w",
			req.Method, req.URL.Path, err)
	}

	return data, nil
}

// executeWithRetry performs the HTTP request with retry logic for 429 and 5xx.
func (c *Client) executeWithRetry(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infakt: %s %s: request failed: %w",
			req.Method, req.URL.Path, err)
	}

	// Handle rate limiting (HTTP 429) — single retry
	if resp.StatusCode == http.StatusTooManyRequests {
		wait := parseRetryAfter(resp.Header.Get("Retry-After"))
		if wait > maxRetryAfterWait {
			wait = maxRetryAfterWait
		}

		select {
		case <-time.After(wait):
		case <-req.Context().Done():
			_ = resp.Body.Close()
			return nil, req.Context().Err()
		}

		_ = resp.Body.Close()
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("infakt: %s %s: retry request failed: %w",
				req.Method, req.URL.Path, err)
		}
	}

	// Retry on server errors (5xx) with exponential backoff
	if resp.StatusCode >= 500 {
		wait := initialRetryWait
		for attempt := 1; attempt < maxServerRetries; attempt++ {
			select {
			case <-time.After(wait):
			case <-req.Context().Done():
				_ = resp.Body.Close()
				return nil, req.Context().Err()
			}

			_ = resp.Body.Close()
			resp, err = c.httpClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("infakt: %s %s: retry request failed (attempt %d): %w",
					req.Method, req.URL.Path, attempt+1, err)
			}

			if resp.StatusCode < 500 {
				break
			}

			wait *= 2
		}
	}

	return resp, nil
}

// parseRetryAfter parses the Retry-After header value into a duration.
func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return time.Second
	}

	if seconds, err := strconv.Atoi(val); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return time.Second
}
