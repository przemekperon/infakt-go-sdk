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
	defaultUserAgent   = "infakt-go-sdk"
	defaultHTTPTimeout = 30 * time.Second
	apiVersion         = "v3"

	defaultRateLimit  = 100 * time.Millisecond
	maxRetryAfterWait = 60 * time.Second
	maxServerRetries  = 3
	initialRetryWait  = 500 * time.Millisecond
)

// Client manages communication with the inFakt API.
//
// A Client is safe for concurrent use across goroutines once it has been
// constructed by [NewClient]. The provided service fields ([Client.Invoices],
// [Client.Clients], [Client.Products], [Client.BankAccounts],
// [Client.VatRates]) may be invoked from multiple goroutines simultaneously;
// requests are serialized only by the optional rate limiter.
//
// [Client.Close] is NOT safe to call concurrently with other Client methods.
// It should only be called when no requests are in flight, typically via
// `defer client.Close()` after construction.
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string

	rateLimiter *time.Ticker

	// Invoices provides access to invoice-related endpoints.
	// See [InvoiceService].
	Invoices *InvoiceService
	// Clients provides access to client-entity (kontrahent) endpoints.
	// See [ClientEntityService].
	Clients *ClientEntityService
	// Products provides access to product-catalog endpoints.
	// See [ProductService].
	Products *ProductService
	// BankAccounts provides read-only access to bank-account endpoints.
	// See [BankAccountService].
	BankAccounts *BankAccountService
	// VatRates provides read-only access to VAT-rate endpoints.
	// See [VatRateService].
	VatRates *VatRateService
}

// Option is a functional option for configuring the [Client]. Options are
// applied in order by [NewClient]; later options can override earlier ones.
// Built-in options include [WithBaseURL], [WithHTTPClient], [WithUserAgent]
// and [WithRateLimit].
type Option func(*Client)

// WithBaseURL sets a custom base URL for the API and returns an [Option] that
// can be passed to [NewClient]. This is most commonly used to point the
// [Client] at a mock or staging server during tests; production code should
// rely on the default URL ("https://api.infakt.pl").
//
// If baseURL fails to parse, the existing default is preserved. Pair with
// [WithHTTPClient] to fully customize transport behavior.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		u, err := url.Parse(baseURL)
		if err == nil {
			c.baseURL = u
		}
	}
}

// WithHTTPClient sets a custom *http.Client used by the [Client] for all
// outbound requests, and returns an [Option] for [NewClient]. Use this to
// inject custom transports (e.g. for tracing or proxying) or to override the
// default 30-second timeout. Combine with [WithBaseURL] when you need both
// transport-level and URL-level customization.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithUserAgent sets a custom User-Agent header sent with every request and
// returns an [Option] for [NewClient]. The default is "infakt-go-sdk"; setting
// a project-specific value (e.g. "my-app/1.0") helps the inFakt operators
// identify traffic from your integration. See also [WithHTTPClient] and
// [WithRateLimit].
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithRateLimit sets the minimum interval between API requests issued by the
// [Client] and returns an [Option] for [NewClient]. The default is 100ms
// between requests. Pass zero (or any non-positive duration) to disable rate
// limiting entirely — convenient for tests using [WithBaseURL] against a
// local mock server. See also [WithHTTPClient].
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
//
// The apiKey argument is required; it is sent on every request via the
// X-inFakt-ApiKey header. The key is NOT validated locally — an invalid or
// empty key is only detected by the server on the first request, which will
// return [ErrUnauthorized].
//
// Defaults applied before opts run:
//   - base URL:     https://api.infakt.pl (override with [WithBaseURL])
//   - HTTP timeout: 30 seconds            (override with [WithHTTPClient])
//   - User-Agent:   "infakt-go-sdk"       (override with [WithUserAgent])
//   - rate limit:   100ms between calls   (override with [WithRateLimit])
//
// Each opt is applied in order and may override any default. The returned
// [Client] is ready for concurrent use; call [Client.Close] when done to
// release the rate-limiter ticker.
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
