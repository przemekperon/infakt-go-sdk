# infakt-go-sdk

[![CI](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/ci.yml)
[![govulncheck](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/govulncheck.yml)
[![CodeQL](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/codeql.yml/badge.svg)](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/codeql.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/przemekperon/infakt-go-sdk.svg)](https://pkg.go.dev/github.com/przemekperon/infakt-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/przemekperon/infakt-go-sdk)](https://goreportcard.com/report/github.com/przemekperon/infakt-go-sdk)

Go client library for the [inFakt API v3](https://www.infakt.pl/) — a Polish invoicing and accounting service.

## Features

- Full CRUD support for invoices, clients, and products
- Read-only access to bank accounts and VAT rates
- Invoice actions: mark as paid, send by email, download PDF, get next number
- Built-in rate limiting with HTTP 429 retry handling
- Automatic retry on 5xx with exponential backoff
- Query filtering for list operations
- Pagination support
- Zero external dependencies (pure stdlib)
- Idiomatic Go error handling with sentinel errors

## Requirements

- Go 1.19+ (the package documentation uses `# Heading` godoc syntax introduced in Go 1.19)
- Module path: `github.com/przemekperon/infakt-go-sdk`
- Zero external runtime dependencies — only the Go standard library

## Installation

```bash
go get github.com/przemekperon/infakt-go-sdk
```

## Documentation

- API reference for this SDK on [pkg.go.dev](https://pkg.go.dev/github.com/przemekperon/infakt-go-sdk)
- Official inFakt API reference: <https://docs.infakt.pl>

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/przemekperon/infakt-go-sdk"
)

func main() {
    client := infakt.NewClient("your-api-key")
    defer client.Close()
    ctx := context.Background()

    // List invoices
    invoices, meta, err := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
        ListOptions: infakt.ListOptions{Limit: 10},
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d invoices (total: %d)\n", len(invoices), meta.TotalCount)

    // Create a client
    entity, err := client.Clients.Create(ctx, &infakt.ClientEntity{
        CompanyName: "ACME Sp. z o.o.",
        NIP:         "1234567890",
        Street:      "ul. Testowa 1",
        City:        "Warszawa",
        PostalCode:  "00-001",
        Country:     "PL",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created client: %s (ID: %d)\n", entity.CompanyName, entity.ID)

    // Create an invoice
    invoice, err := client.Invoices.Create(ctx, &infakt.Invoice{
        ClientID:    entity.ID,
        InvoiceDate: "2025-01-15",
        SaleDate:    "2025-01-15",
        Services: []infakt.ServiceEntry{
            {
                Name:         "Consulting",
                TaxSymbol:    "23",
                Unit:         "hour",
                Quantity:     10,
                UnitNetPrice: 15000,
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created invoice: %s\n", invoice.Number)
}
```

## Configuration

Functional options on `NewClient` cover the common knobs:

```go
client := infakt.NewClient("key",
    infakt.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
    infakt.WithRateLimit(200 * time.Millisecond),
    infakt.WithUserAgent("my-app/1.0"),
)
```

See [pkg.go.dev](https://pkg.go.dev/github.com/przemekperon/infakt-go-sdk#Option) for the full list of options and their defaults.

## Resources

| Resource | Service | Operations |
|----------|---------|------------|
| Invoices | `client.Invoices` | List, Get, Create, Update, Delete, MarkAsPaid, SendByEmail, GetPDF, GetNextNumber |
| Clients | `client.Clients` | List, Get, Create, Update, Delete |
| Products | `client.Products` | List, Get, Create, Update, Delete |
| Bank Accounts | `client.BankAccounts` | List |
| VAT Rates | `client.VatRates` | List |

## Error Handling

Use sentinel errors with `errors.Is`, or `errors.As` to inspect the full response:

```go
_, err := client.Invoices.Get(ctx, 999)
if errors.Is(err, infakt.ErrNotFound) {
    // handle 404
}

var apiErr *infakt.ErrorResponse
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
}
```

Sentinel errors: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrRateLimited`. See [pkg.go.dev](https://pkg.go.dev/github.com/przemekperon/infakt-go-sdk#pkg-variables) for details.

## Filtering

```go
// Filter invoices by date range and status
invoices, _, _ := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
    DateFrom: "2025-01-01",
    DateTo:   "2025-12-31",
    Status:   "paid",
})

// Filter clients by company name
clients, _, _ := client.Clients.List(ctx, &infakt.ClientEntityListOptions{
    CompanyName: "ACME",
})
```

## Context Support

Every service method takes a `context.Context` as its first argument. Cancelling the context — or letting its deadline expire — aborts the in-flight HTTP request and any pending retry wait, and returns the context error to the caller. This is the recommended way to apply per-call timeouts on top of the HTTP client's own timeout.

## Versioning

This module follows [Semantic Versioning](https://semver.org/). While the major version is `0` (v0.x), the public API may change in backwards-incompatible ways between minor releases. Pin a specific version in your `go.mod` and consult [CHANGELOG.md](CHANGELOG.md) before upgrading.

## Testing

Run unit tests and runnable examples (which use `httptest.NewServer`, no real API):

```bash
go test ./...
```

The repository also contains `_live_test.go`, a manual smoke-test program excluded from normal builds with the `//go:build ignore` constraint. It expects the API key in the `INFAKT_KEY` environment variable and is intended to be run directly against the live inFakt API, e.g.:

```bash
INFAKT_KEY=... go run _live_test.go
```

Do not run live tests against production data without understanding the side effects.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the development workflow, coding style, and PR checklist.

## Security

To report a security issue, see [SECURITY.md](SECURITY.md).

## License

[MIT](LICENSE)
