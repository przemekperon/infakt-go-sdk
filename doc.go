// Package infakt provides a Go client for the inFakt API v3.
//
// inFakt is a Polish invoicing and accounting service. This package implements
// a client for the inFakt REST API v3, providing access to invoices, clients,
// products, bank accounts, VAT rates, GTU reference codes, and cost invoices
// (faktury kosztowe).
//
// # Authentication
//
// All API requests require an API key, which can be obtained from your inFakt
// account settings. Pass the key when creating a new client with [NewClient]:
//
//	client := infakt.NewClient("your-api-key")
//	defer client.Close()
//
// # Usage
//
// The [Client] exposes API resources through service fields:
//
//	// List invoices
//	invoices, meta, err := client.Invoices.List(ctx, nil)
//
//	// Create a client entity
//	entity, err := client.Clients.Create(ctx, &infakt.ClientEntity{
//	    CompanyName: "ACME Sp. z o.o.",
//	    NIP:         "1234567890",
//	})
//
//	// Get a product
//	product, err := client.Products.Get(ctx, 42)
//
//	// List cost invoices (faktury kosztowe)
//	costs, _, err := client.CostInvoices.List(ctx, &infakt.CostInvoiceListOptions{
//	    DateFrom: "2026-01-01",
//	    DateTo:   "2026-12-31",
//	})
//
// Call [Client.Close] when the client is no longer needed to release the
// internal rate-limiter ticker.
//
// # Pagination
//
// List methods accept resource-specific options that embed [ListOptions] for
// paging. See for example [InvoiceListOptions]:
//
//	invoices, meta, err := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
//	    ListOptions: infakt.ListOptions{Offset: 0, Limit: 25},
//	})
//	fmt.Printf("Page has %d items, total: %d\n", meta.Count, meta.TotalCount)
//
// # Error Handling
//
// The package exposes sentinel errors for common HTTP status codes; use
// [errors.Is] to test for them:
//
//	_, err := client.Invoices.Get(ctx, 999)
//	if errors.Is(err, infakt.ErrNotFound) {
//	    // handle 404
//	}
//
// For detailed error information, use [errors.As] with [*ErrorResponse]:
//
//	var apiErr *infakt.ErrorResponse
//	if errors.As(err, &apiErr) {
//	    fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
//	}
//
// # Rate Limiting
//
// The client paces outgoing requests with an internal ticker (default: 100ms
// between requests). Configure the interval with [WithRateLimit]; pass 0 to
// disable pacing.
//
//	client := infakt.NewClient("key", infakt.WithRateLimit(200*time.Millisecond))
//
// Other client behavior can be tuned with [WithBaseURL], [WithHTTPClient], and
// [WithUserAgent].
//
// # Retries
//
// Retry behavior is built in and is not user-configurable:
//
//   - HTTP 429 (Too Many Requests): the client retries the request once,
//     honoring the Retry-After response header. The wait is capped at 60s.
//   - HTTP 5xx (server errors): the client retries up to 3 attempts total
//     (the initial request plus two retries) with exponential backoff
//     starting at 500ms and doubling between attempts.
//
// In both cases, a canceled or expired [context.Context] aborts pending waits
// and returns the context error to the caller. Network/transport errors are
// not retried.
//
// # Context and Cancellation
//
// Every method on every service takes a [context.Context] as its first
// argument. Canceling the context (or hitting its deadline) aborts the
// in-flight HTTP request and any pending retry wait, and propagates the
// context error back to the caller. This is the recommended way to enforce
// per-call timeouts on top of the HTTP client's own timeout.
//
// # Stability and Versioning
//
// This module follows [Semantic Versioning]. While the major version is 0
// (v0.x), the public API may change in backwards-incompatible ways between
// minor releases. Pin a specific version in your go.mod and consult
// CHANGELOG.md before upgrading.
//
// # Documentation and References
//
// The official inFakt API reference is available at the inFakt developer
// portal. Refer to it for endpoint behavior, field semantics, and request
// limits not documented here.
//
// [Semantic Versioning]: https://semver.org/
//
// [inFakt developer portal]: https://docs.infakt.pl
package infakt
