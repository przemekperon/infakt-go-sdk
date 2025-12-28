// Package infakt provides a Go client for the inFakt API v3 (https://www.infakt.pl/).
//
// inFakt is a Polish invoicing and accounting service. This package implements
// a client for the inFakt REST API v3, providing access to invoices, clients,
// products, bank accounts, and VAT rates.
//
// # Authentication
//
// All API requests require an API key, which can be obtained from your inFakt
// account settings. Pass the key when creating a new client:
//
//	client := infakt.NewClient("your-api-key")
//
// # Usage
//
// The client provides access to different API resources through service fields:
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
// # Pagination
//
// List methods accept ListOptions for pagination:
//
//	invoices, meta, err := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
//	    ListOptions: infakt.ListOptions{Offset: 0, Limit: 25},
//	})
//	fmt.Printf("Page has %d items, total: %d\n", meta.Count, meta.TotalCount)
//
// # Error Handling
//
// The package provides sentinel errors for common HTTP status codes:
//
//	_, err := client.Invoices.Get(ctx, 999)
//	if errors.Is(err, infakt.ErrNotFound) {
//	    // handle 404
//	}
//
// For detailed error information, use errors.As with *ErrorResponse:
//
//	var apiErr *infakt.ErrorResponse
//	if errors.As(err, &apiErr) {
//	    fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
//	}
//
// # Rate Limiting
//
// The client includes built-in rate limiting (default: 100ms between requests)
// and automatic retry on HTTP 429 responses. Configure with WithRateLimit:
//
//	client := infakt.NewClient("key", infakt.WithRateLimit(200 * time.Millisecond))
package infakt
