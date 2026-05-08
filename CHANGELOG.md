# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/przemekperon/infakt-go-sdk/releases/tag/v0.1.0
