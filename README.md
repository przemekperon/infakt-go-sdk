# golang-infakt

[![Go Reference](https://pkg.go.dev/badge/github.com/przemekperon/golang-infakt.svg)](https://pkg.go.dev/github.com/przemekperon/golang-infakt)
[![Go Report Card](https://goreportcard.com/badge/github.com/przemekperon/golang-infakt)](https://goreportcard.com/report/github.com/przemekperon/golang-infakt)

Go client library for the [inFakt API v3](https://www.infakt.pl/) — a Polish invoicing and accounting service.

## Features

- Full CRUD support for invoices, clients, and products
- Read-only access to bank accounts and VAT rates
- Invoice actions: mark as paid, send by email, download PDF, get next number
- Built-in rate limiting with HTTP 429 retry handling
- Query filtering for list operations
- Pagination support
- Zero external dependencies (pure stdlib)
- Idiomatic Go error handling with sentinel errors

## Installation

```bash
go get github.com/przemekperon/golang-infakt
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/przemekperon/golang-infakt"
)

func main() {
    client := infakt.NewClient("your-api-key")
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

```go
// Custom HTTP client
client := infakt.NewClient("key",
    infakt.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)

// Custom rate limiting
client := infakt.NewClient("key",
    infakt.WithRateLimit(200 * time.Millisecond),
)

// Custom User-Agent
client := infakt.NewClient("key",
    infakt.WithUserAgent("my-app/1.0"),
)
```

## Resources

| Resource | Service | Operations |
|----------|---------|------------|
| Invoices | `client.Invoices` | List, Get, Create, Update, Delete, MarkAsPaid, SendByEmail, GetPDF, GetNextNumber |
| Clients | `client.Clients` | List, Get, Create, Update, Delete |
| Products | `client.Products` | List, Get, Create, Update, Delete |
| Bank Accounts | `client.BankAccounts` | List |
| VAT Rates | `client.VatRates` | List |

## Error Handling

```go
import "errors"

// Sentinel errors
_, err := client.Invoices.Get(ctx, 999)
if errors.Is(err, infakt.ErrNotFound) {
    fmt.Println("Invoice not found")
}

// Detailed error info
var apiErr *infakt.ErrorResponse
if errors.As(err, &apiErr) {
    fmt.Printf("Status: %d, Message: %s\n", apiErr.StatusCode, apiErr.Message)
}
```

Available sentinel errors: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrRateLimited`.

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

## License

[MIT](LICENSE)
