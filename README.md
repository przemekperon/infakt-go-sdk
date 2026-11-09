# infakt-go-sdk

[![CI](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/ci.yml)
[![govulncheck](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/govulncheck.yml/badge.svg)](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/govulncheck.yml)
[![CodeQL](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/codeql.yml/badge.svg)](https://github.com/przemekperon/infakt-go-sdk/actions/workflows/codeql.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/przemekperon/infakt-go-sdk.svg)](https://pkg.go.dev/github.com/przemekperon/infakt-go-sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/przemekperon/infakt-go-sdk)](https://goreportcard.com/report/github.com/przemekperon/infakt-go-sdk)

Go client library for the [inFakt API v3](https://www.infakt.pl/) — a Polish invoicing and accounting service.

## Features

- Full CRUD support for invoices, clients, and products
- Bank account CRUD; read-only VAT rates, GTU codes, and cost invoices (faktury kosztowe)
- Invoice actions: mark as paid, send by email, download PDF, get next number
- Context-aware rate limiting and bounded GET/HEAD retries on HTTP 429 and 5xx
- Same-origin redirect restrictions protecting API credentials
- Query filtering for list operations
- Pagination support
- Zero external dependencies (pure stdlib)
- Idiomatic Go error handling with sentinel errors

## Requirements

- Go 1.24+ (as declared in `go.mod`)
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
    entity, err := client.Clients.CreateWithRequest(ctx, &infakt.ClientEntityRequest{
        CompanyName: infakt.String("ACME Sp. z o.o."),
        NIP:         infakt.String("1234567890"),
        Street:      infakt.String("ul. Testowa 1"),
        City:        infakt.String("Warszawa"),
        PostalCode:  infakt.String("00-001"),
        Country:     infakt.String("PL"),
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created client: %s (ID: %d)\n", entity.CompanyName, entity.ID)

    // Create an invoice
    invoice, err := client.Invoices.CreateWithRequest(ctx, &infakt.InvoiceRequest{
        ClientID:    infakt.Int64(entity.ID),
        InvoiceDate: infakt.String("2025-01-15"),
        SaleDate:    infakt.String("2025-01-15"),
        Services: &[]infakt.ServiceEntryRequest{
            {
                Name:         infakt.String("Consulting"),
                TaxSymbol:    infakt.String("23"),
                Unit:         infakt.String("hour"),
                Quantity:     infakt.Float64(10),
                UnitNetPrice: infakt.Int(15000),
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
| Bank Accounts | `client.BankAccounts` | List, Get, Create, Update, Delete |
| VAT Rates | `client.VatRates` | List |
| GTU codes | `client.Gtus` | List, Get, ListSelected |
| Cost Invoices | `client.CostInvoices` | List, Get |

## Error Handling

Use sentinel errors with `errors.Is`, or `errors.As` to inspect the full response:

```go
_, err := client.Invoices.Get(ctx, "999")
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

## Migrating to 0.3.0

This is a minor release with public API and behavior changes under the module's
v0 versioning policy. Use typed `CreateWithRequest` and `UpdateWithRequest`
methods on `Invoices`, `Clients`, and `BankAccounts`. The model-based methods
are deprecated; legacy `Invoices.Update` returns a local migration error.

A nil DTO field omits that field; a pointer to `""`, `0`, or `false` sends that
value explicitly. This controls JSON presence, while the API still validates
business rules. Nil request arguments return local errors. For example, with
`client`, `ctx`, and `invoiceUUID` already defined:

```go
// Clear a note and explicitly confirm there is no JST recipient address
// to retain (or that deleting it is intended).
updated, err := client.Invoices.UpdateWithRequest(ctx, invoiceUUID, &infakt.InvoiceRequest{
    Notes: infakt.String(""),
    RemoveLocalGovernmentRecipientAddress: true,
})
_ = updated
_ = err

// Explicitly unset the default flag on a bank account.
account, err := client.BankAccounts.UpdateWithRequest(ctx, 42, &infakt.BankAccountRequest{
    Default: infakt.Bool(false),
})
_ = account
_ = err
```

To retain or replace a JST recipient, instead supply a complete
`LocalGovernmentRecipientAddress` request DTO. Its field names differ from the
response: `TaxID` corresponds to `NIP`, `BuildingNumber` to `StreetNumber`,
`DoorNumber` to `FlatNumber`, and `PostCode` to `PostalCode`. Supply the documented
integer `CountryID` explicitly when retaining a non-default country (omission
defaults to Poland); the SDK does not guess it from a country code.
Without either an address or explicit removal consent, updates fail locally
before any PUT. Removal uses the API's documented omission behavior, not JSON
null. This also applies when editing an invoice that currently has no JST address.
For example, with `recipientCountryID` resolved from the country catalogue and
all address values taken from the intended current state:

```go
updated, err := client.Invoices.UpdateWithRequest(ctx, invoiceUUID, &infakt.InvoiceRequest{
    Notes: infakt.String(""),
    LocalGovernmentRecipientAddress: &infakt.LocalGovernmentAddressRequest{
        CompanyName: "Example school",
        Street: "Szkolna", BuildingNumber: "1", City: "Warszawa",
        PostCode: "00-001", TaxID: "0000000000",
        CountryID: infakt.Int64(recipientCountryID),
    },
})
_ = updated
_ = err
```

The SDK performs no hidden GET/merge/PUT. Callers must coordinate concurrent
edits; an address copied from an earlier read may already be stale.

`InvoiceRequest.Services` is now `*[]ServiceEntryRequest`. Nil omits the list;
a supplied list replaces **all** existing lines. Provide the complete desired
list with the required fields on each item. An empty list is serialized as
`[]`; documentation does not guarantee that an empty invoice passes business
validation. Do not use a pointer to a nil slice to request JSON null:
that deletion operation is not documented.

`Invoice.DocumentMarkingsIDs` is now `[]string`, e.g. `[]string{"001"}`,
preserving exact strings. Update/create DTOs use a pointer to a string slice;
the documented maximum is one marking. Response `Extensions` now has payment
and share data; write DTOs only expose documented request fields. Existing
model-based Create methods remain migration adapters and cannot express the
same presence decisions as DTOs. Legacy invoice creation with a response JST
address is rejected locally; migrate that address to its request type.

`ClientEntityRequest.DaysToPayment` changes from `*string` to `*int`: use
`infakt.Int(14)` instead of `infakt.String("14")`.
`ServiceEntryRequest.Discount` changes from `*string` to `*float64`: use
`infakt.Float64(5)` instead of `infakt.String("5")`.
`ServiceEntryRequest.VatDateValue` is removed; set the documented invoice-level
`InvoiceRequest.VatDateValue` instead (e.g. `infakt.String("sale_date")`).
Response extensions are `Extensions.Payments *PaymentExtension` and
`Extensions.Shares *ShareExtension`, with nullable links and share validity.
Legacy `PaymentOnline` is deprecated and is absent from the typed write DTO.

Get and Update accept an invoice UUID or numeric ID as a string, e.g. `"999"`.
Action and PDF endpoint paths are unchanged; consult each method for its
accepted identifiers.

Only GET/HEAD requests retry HTTP 429 or 5xx, excluding PDF downloads: rendering
a PDF can mark a draft as printed and consume its print limit. A failed write may already have
been applied. Reconcile its remote outcome before deciding to retry manually;
do not automatically repeat creation or email actions in application code.
There is no exactly-once guarantee. Retries recreate replayable bodies; a body
without `GetBody` stops retries and preserves the original API response.

Redirects are allowed only within the same origin (scheme, hostname and
effective port), and HTTPS-to-HTTP redirects are rejected. `WithHTTPClient`
copies the supplied client without modifying it, preserves its configuration,
and enforces origin/chain limits before honoring its `CheckRedirect` callback.
Do not rely on pointer identity with the supplied client. Close the SDK client
after in-flight calls finish; concurrent `Close` remains outside the contract.

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
