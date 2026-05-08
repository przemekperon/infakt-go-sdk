# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-05-08

### Added

- `CONTRIBUTING.md` with local development, testing, and PR workflow.
- `SECURITY.md` with responsible-disclosure policy and API-key handling guidance.
- `examples/` directory with four runnable programs: `quickstart`, `pagination`, `pdf-export`, `error-handling`, plus index `examples/README.md`.
- 22 new runnable Examples in `example_test.go` (29 total, all with `// Output:` and verified by `go test`), backed by an `httptest`-based mock helper.
- Field-level godoc on `Invoice`, `ServiceEntry`, `ClientEntity`, `Product`, `BankAccount`, `VatRate`, and `ErrorResponse` (semantic explanations: ISO 4217 currency, grosze monetary fields, `YYYY-MM-DD` date format, IBAN, Polish NIP/PKWiU/MPP, etc.).
- Service-level godoc on every `*Service` with supported endpoints and links to https://docs.infakt.pl.
- Godoc on previously undocumented `InvoiceRequest` and `ServiceEntryRequest` types.
- Field-level comments on `Client.Invoices`, `Client.Clients`, `Client.Products`, `Client.BankAccounts`, `Client.VatRates`.

### Changed

- `doc.go` expanded with new sections: `Stability and Versioning`, `Context and Cancellation`, `Retries`, `Documentation and References` (link to https://docs.infakt.pl).
- `README.md` extended with `Requirements`, `Documentation`, `Versioning`, `Context Support`, `Testing`, `Contributing`, and `Security` sections; configuration and error-handling sections slimmed to point at `pkg.go.dev` for the full reference.
- Adopted Go 1.19+ doc-link syntax (`[Name]`) across the package for cross-references between types, methods, sentinel errors, and stdlib symbols.
- Documented concurrency guarantees on `Client` and clarified `NewClient` defaults.

### Notes

- No public API changes; this release is documentation-only.

## [0.1.0] - 2026-01-11

### Added

- Core HTTP client with functional options (`WithBaseURL`, `WithHTTPClient`, `WithUserAgent`, `WithRateLimit`)
- Invoice resource with full CRUD operations
- Invoice actions: `MarkAsPaid`, `SendByEmail`, `GetPDF`, `GetNextNumber`
- Client entity resource with full CRUD operations
- Product resource with full CRUD operations
- Bank account resource (read-only, list)
- VAT rate resource (read-only, list)
- Pagination support with `ListOptions` and `MetaInfo`
- Query filtering for invoices (date range, client ID, status), clients (company name, NIP), and products (name)
- Custom error types with HTTP status mapping (`ErrorResponse`)
- Sentinel errors: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, `ErrRateLimited`
- Built-in rate limiting with configurable interval (default: 100ms)
- Automatic retry on HTTP 429 with `Retry-After` header support
- Pointer helper functions: `String()`, `Int()`, `Int64()`, `Bool()`, `Float64()`, `Time()`
- Typed request structs for create/update operations
- Comprehensive package documentation and runnable examples
- GitHub Actions CI with Go 1.22/1.23 matrix testing and golangci-lint

[0.1.1]: https://github.com/przemekperon/infakt-go-sdk/releases/tag/v0.1.1
[0.1.0]: https://github.com/przemekperon/infakt-go-sdk/releases/tag/v0.1.0
