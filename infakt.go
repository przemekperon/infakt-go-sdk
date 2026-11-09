package infakt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	sandboxBaseURL     = "https://api.sandbox-infakt.pl"
	defaultUserAgent   = "infakt-go-sdk"
	defaultHTTPTimeout = 30 * time.Second
	apiVersion         = "v3"

	// defaultRateLimit paces outgoing requests to stay under the inFakt
	// IP-level limits documented at https://docs.infakt.pl (sekcja
	// "Limity"): 300 GETs/min and 150 writes/min. 400ms = 150 req/min
	// is safe for both classes.
	defaultRateLimit  = 400 * time.Millisecond
	maxRetryAfterWait = 60 * time.Second
	maxServerRetries  = 3
	initialRetryWait  = 500 * time.Millisecond
)

// Client manages communication with the inFakt API.
//
// A Client is safe for concurrent use across goroutines once it has been
// constructed by [NewClient]. The provided service fields ([Client.Invoices],
// [Client.Clients], [Client.Products], [Client.BankAccounts],
// [Client.VatRates], [Client.Gtus], [Client.CostInvoices]) may be invoked
// from multiple goroutines simultaneously; requests are serialized only by
// the optional rate limiter.
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
	// BankAccounts provides access to bank-account endpoints.
	// See [BankAccountService].
	BankAccounts *BankAccountService
	// VatRates provides read-only access to VAT-rate endpoints.
	// See [VatRateService].
	VatRates *VatRateService
	// Gtus provides read-only access to GTU (Polish JPK_V7 commodity
	// group) reference codes. See [GtuService].
	Gtus *GtuService
	// CostInvoices provides read-only access to cost-invoice (faktury
	// kosztowe) endpoints. See [CostInvoiceService].
	CostInvoices *CostInvoiceService
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
// [WithHTTPClient] to fully customize transport behavior. To target the
// official inFakt sandbox, prefer [WithSandbox] over passing the URL
// manually.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		u, err := url.Parse(baseURL)
		if err == nil {
			c.baseURL = u
		}
	}
}

// WithSandbox points the [Client] at the inFakt sandbox environment
// (https://api.sandbox-infakt.pl) and returns an [Option] for [NewClient].
// The sandbox accepts the same authentication and resource paths as
// production but operates on isolated data, making it suitable for
// integration tests and demos. Equivalent to passing
// [WithBaseURL]("https://api.sandbox-infakt.pl").
func WithSandbox() Option {
	return WithBaseURL(sandboxBaseURL)
}

// WithHTTPClient sets a custom *http.Client used by the [Client] for all
// outbound requests, and returns an [Option] for [NewClient]. Use this to
// inject custom transports (e.g. for tracing or proxying) or to override the
// default 30-second timeout. Combine with [WithBaseURL] when you need both
// transport-level and URL-level customization. NewClient copies its configuration
// without mutating the supplied client. Redirects must stay on the same origin;
// the supplied CheckRedirect may impose further restrictions.
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
// [Client] and returns an [Option] for [NewClient]. The default is 400ms
// between requests, sized to stay under the documented IP-level limit of
// 150 writes/min (https://docs.infakt.pl, sekcja "Limity"). Pass zero (or
// any non-positive duration) to disable rate limiting entirely — convenient
// for tests using [WithBaseURL] against a local mock server. See also
// [WithHTTPClient].
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
//   - base URL:     https://api.infakt.pl (override with [WithBaseURL] or [WithSandbox])
//   - HTTP timeout: 30 seconds            (override with [WithHTTPClient])
//   - User-Agent:   "infakt-go-sdk"       (override with [WithUserAgent])
//   - rate limit:   400ms between calls   (override with [WithRateLimit])
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

	// Copy configuration: callers may share and already be using this client.
	if c.httpClient == nil {
		c.httpClient = &http.Client{Timeout: defaultHTTPTimeout}
	}
	httpClient := *c.httpClient
	userCheckRedirect := httpClient.CheckRedirect
	// SDK requests start at baseURL. Keep its origin outside the mutable
	// request history passed to user callbacks throughout a redirect chain.
	origin := *c.baseURL
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) == 0 || req.URL == nil || !sameOrigin(&origin, req.URL) {
			return errors.New("infakt: redirect to a different origin is not allowed")
		}
		if len(via) >= 10 {
			return errors.New("infakt: stopped after 10 redirects")
		}
		if userCheckRedirect != nil {
			// Callbacks may modify req or via; enforce the immutable boundary
			// again before allowing a network request.
			if err := userCheckRedirect(req, via); err != nil {
				return err
			}
			if req.URL == nil || !sameOrigin(&origin, req.URL) {
				return errors.New("infakt: redirect to a different origin is not allowed")
			}
			if len(via) >= 10 {
				return errors.New("infakt: stopped after 10 redirects")
			}
		}
		return nil
	}
	c.httpClient = &httpClient

	c.Invoices = &InvoiceService{client: c}
	c.Clients = &ClientEntityService{client: c}
	c.Products = &ProductService{client: c}
	c.BankAccounts = &BankAccountService{client: c}
	c.VatRates = &VatRateService{client: c}
	c.Gtus = &GtuService{client: c}
	c.CostInvoices = &CostInvoiceService{client: c}

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
// It applies rate limiting to each attempt. Only GET and HEAD are retried on
// HTTP 429 or 5xx, for at most three total attempts (excluding invoice PDF
// rendering, which has side effects). A failed write may have
// taken effect; reconcile its outcome before manually repeating it.
func (c *Client) do(req *http.Request, v interface{}) error {
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

// sameOrigin compares schemes, hostnames and effective ports. In particular,
// HTTPS-to-HTTP redirects and redirects to another port are rejected.
func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) &&
		strings.EqualFold(a.Hostname(), b.Hostname()) && effectivePort(a) == effectivePort(b)
}

func effectivePort(u *url.URL) string {
	if port := u.Port(); port != "" {
		return port
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}

// waitForRateLimit is shared by JSON, raw responses, and retries. Close must
// only be called once requests have finished, as documented on Client.
func (c *Client) waitForRateLimit(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.rateLimiter != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.rateLimiter.C:
		}
	}
	return ctx.Err()
}

// retryAllowed deliberately excludes every write, including on HTTP 429:
// neither 429 nor 5xx provides a confirmed guarantee of non-execution.
func retryAllowed(req *http.Request, status int) bool {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return false
	}
	// Rendering an invoice PDF can mark it printed and consume a print limit.
	if strings.HasPrefix(req.URL.Path, "/"+apiVersion+"/invoices/") && strings.HasSuffix(req.URL.Path, "/pdf.json") {
		return false
	}
	return status == http.StatusTooManyRequests || (status >= 500 && status <= 599)
}

// executeWithRetry gives each allowed retry a fresh request and body. Without
// GetBody a consumed body cannot be replayed, so the original response is
// returned for normal API error handling instead of retrying it.
func (c *Client) executeWithRetry(req *http.Request) (*http.Response, error) {
	wait := initialRetryWait
	rateRetried := false
	for attempt := 0; ; attempt++ {
		if err := c.waitForRateLimit(req.Context()); err != nil {
			return nil, err
		}
		attemptReq := req.Clone(req.Context())
		if attempt > 0 && req.Body != nil && req.Body != http.NoBody {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("infakt: failed to replay request body: %w", err)
			}
			attemptReq.Body = body
		}
		// Recreating a body may run caller code and take time. Cancellation
		// during GetBody must prevent dispatch, even with a custom transport.
		if err := req.Context().Err(); err != nil {
			if attemptReq.Body != nil {
				_ = attemptReq.Body.Close()
			}
			return nil, err
		}
		resp, err := c.httpClient.Do(attemptReq)
		if err != nil {
			return nil, fmt.Errorf("infakt: %s %s: request failed: %w",
				req.Method, req.URL.Path, err)
		}
		if !retryAllowed(req, resp.StatusCode) || attempt+1 == maxServerRetries ||
			(req.Body != nil && req.Body != http.NoBody && req.GetBody == nil) {
			return resp, nil
		}
		delay := wait
		if resp.StatusCode == http.StatusTooManyRequests {
			if rateRetried {
				return resp, nil
			}
			rateRetried = true
			delay = parseRetryAfter(resp.Header.Get("Retry-After"))
			if delay > maxRetryAfterWait {
				delay = maxRetryAfterWait
			}
		} else {
			wait *= 2
		}
		_ = resp.Body.Close()
		timer := time.NewTimer(delay)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}
}

// parseRetryAfter parses the Retry-After header value into a duration.
func parseRetryAfter(val string) time.Duration {
	if val == "" {
		return time.Second
	}

	if seconds, err := strconv.Atoi(val); err == nil {
		if seconds <= 0 {
			return 0
		}
		if seconds > int(maxRetryAfterWait/time.Second) {
			return maxRetryAfterWait
		}
		return time.Duration(seconds) * time.Second
	}

	return time.Second
}
