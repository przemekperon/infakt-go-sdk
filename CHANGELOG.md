# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-06-17

This release reconciles the SDK with the official inFakt API documentation
(https://docs.infakt.pl) and verifies the entire public surface end-to-end
against the official sandbox. It contains breaking changes on the Invoice and
Product write paths; consult the migration notes below before upgrading.

### Breaking Changes

- `Invoice` per-resource operations now identify invoices by **UUID
  (string)** instead of the integer `Invoice.ID`. Affected methods:
  `Invoices.Get`, `Invoices.Update`, `Invoices.Delete`, `Invoices.MarkAsPaid`,
  `Invoices.SendByEmail`, `Invoices.GetPDF`. The API endpoints documented at
  https://docs.infakt.pl are addressed by `{invoice_uuid}` and the previous
  integer paths returned 404 in production.
- `Invoices.MarkAsPaid` now targets `/async/invoices/{uuid}/paid.json` (the
  documented async endpoint) and accepts an additional `allowCorrection bool`
  argument required when paying corrective invoices.
- `Invoices.SendByEmail` signature changed: the second argument is now
  `*SendByEmailOptions` exposing `PrintType`, `Locale`, `Recipient`, and
  `SendCopy`. The previous fixed `print_type=original` and `email_to`
  payload key did not match the documented contract (`recipient`).
- `Invoices.GetPDF` signature changed: now takes `(uuid, documentType, locale)`
  and addresses the documented `/{uuid}/pdf.json` endpoint, which responds with
  the document directly (`Content-Type: application/pdf`). It returns the raw
  PDF bytes (`[]byte`); the previous `/{id}.pdf` path did not exist. (Verified
  against the sandbox: the endpoint returns the binary PDF, not a JSON envelope
  with a download link.)
- `Invoice.SplitPayment bool` removed; replaced by
  `Invoice.SplitPaymentType string` (`"required"` / `"optional"`) per the
  API schema.
- `Invoice.VatExemptionReason` is now `int` (was `string`).
- `ClientEntity.DaysToPayment` is now typed as `FlexString` (was `string`).
  The live API is inconsistent: it returns an empty string (`""`) when the
  term is unset but a bare JSON number (e.g. `14`) once a value is assigned,
  which a plain `string` field cannot decode. `FlexString` accepts both shapes
  and marshals back as a JSON string, which the API accepts on input.
  `ClientEntityRequest.DaysToPayment` stays `*string`.
- `Products.Create` and `Products.Update` now accept `*ProductRequest`
  instead of `*Product`, enabling explicit zero-value writes via pointer
  helpers. `ProductRequest.Description` removed (the field is not part of
  the documented schema).
- `Product.UUID` removed (not part of the documented Product schema).

### Added

- **CostInvoices** read-only access to cost invoices (faktury kosztowe)
  via `Client.CostInvoices`, backed by `GET /v3/documents/costs.json` and
  `GET /v3/documents/costs/{uuid}.json` per
  https://github.com/infakt/API/blob/master/readme.md#koszty. Supports
  paginated listing with `q[issue_date_gteq]` / `q[issue_date_lteq]` /
  `q[seller_tax_code_eq]` / `q[currency_eq]` filters and per-document
  fetch by UUID. Requires the `api:costs:read` scope on the API key.
- **BankAccounts** full CRUD: `Get`, `Create`, `Update`, `Delete` (the API
  exposes `POST/GET/{id}/PUT/DELETE` and the previous "read-only" annotation
  was incorrect). Mutating endpoints require the
  `api:sensitive:bank_accounts:write` scope on the API key.
- **GTU** reference catalog (`Client.Gtus`) backed by `/gtus.json`,
  `/gtus/{id}.json`, `/gtus/selected.json` for JPK_V7 commodity markings.
- `WithSandbox()` client option pointing at
  `https://api.sandbox-infakt.pl`.
- Sentinel errors expanded: `ErrBadRequest` (400), `ErrPaymentRequired`
  (402), `ErrUnprocessableEntity` (422), `ErrLocked` (423). 503 now maps
  to `ErrRateLimited` (matches docs which group it with 429).
- `ErrorResponse.Errors` (`map[string][]string`) captures the per-field
  validation messages returned on 422 responses (e.g.
  `{"bank_account": ["..."]}`); the detail is also appended to
  `ErrorResponse.Error()` so the cause is visible in logs. Previously only
  the generic top-level message was surfaced.
- `ListOptions` extended with `Order`, `Fields`, and `Filters` so all
  resource-list calls can express documented Ransack-style filters
  (`q[<predicate>]`), sort expressions (`order=name asc`), and field
  selection (`fields=name,services(name,tax_symbol)`).
- `Limit` is now clamped to the documented maximum of 100 records per page
  (constant `MaxPageLimit`).
- New Invoice fields: `ContinuousServiceStartOn`, `ContinuousServiceEndOn`,
  `BDOCode`, `DocumentMarkingsIDs`, `ReceiptNumber`,
  `CheckDuplicateNumber`, `KsefData`.
- New ServiceEntry fields: `CN`, `PKOB`, `GtuID`, `FlatRateTaxSymbol`,
  `VatDateValue`.
- New Product fields: `CN`, `PKOB`, `GtuID`, `PurchaseUnitNetPrice`,
  `PurchaseUnitGrossPrice`. `ProductRequest` extended with all editable
  fields (CN, PKOB, GtuID, FlatRateTaxSymbol, Discount, purchase prices).
- `ProductListOptions.NameEq` for `q[name_eq]` filter.
- `PDFDocumentTypeOriginal` / `PDFDocumentTypeDuplicate` /
  `PDFDocumentTypeCopy` constants for `Invoices.GetPDF` and
  `SendByEmailOptions.PrintType`.

### Fixed

- Verified the entire public surface end-to-end against the official sandbox
  (`https://api.sandbox-infakt.pl`), exercising every service method including
  the write paths. The fixes below address decode/encode mismatches found
  during that run.
- `Gtu` response mapping: the struct referenced a non-existent `code`
  attribute and was missing `short_description`. The API returns the JPK_V7
  marking in `name` (e.g. `"GTU_01"`) plus `short_description` and
  `description`; the struct now matches (`Gtu.Code` removed, `Gtu.Name` and
  `Gtu.ShortDescription` are authoritative).
- `ClientEntity.DaysToPayment` decode failure: a populated client returns
  `days_to_payment` as a JSON number, which the `string` field rejected with
  "cannot unmarshal number into Go struct field". Fixed via the new
  `FlexString` type (see Breaking Changes).
- `Invoices.GetPDF` decode failure: the endpoint returns a binary PDF, which
  the JSON-envelope decode rejected with "invalid character '%'". `GetPDF`
  now returns the raw bytes (see Breaking Changes).

### Changed

- Default rate limit raised from 100 ms to **400 ms** to honour the
  documented IP-level limit of 150 writes/minute (https://docs.infakt.pl,
  sekcja "Limity"). The previous 100 ms (600 req/min) reliably tripped 429
  on write-heavy workloads.
- `ErrorResponse` now extracts `message` from the JSON `error` key (the
  format documented and used by the live API). The legacy `message` key
  is still parsed as a fallback.
- Service-level godoc on `BankAccountService` updated to reflect full CRUD
  surface; "read-only" wording removed.

### CI

- Bumped GitHub Actions: `actions/checkout` v4→v6,
  `actions/upload-artifact` v4→v7, `github/codeql-action` v3→v4 (#6).
- Fixed `errcheck` lint findings by handling `Fprint`/`Fprintf` return
  values in test mocks.
- Ignored per-developer Claude Code local settings.

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
- Multi-OS CI matrix (Linux / macOS / Windows × Go 1.22 / 1.23) with `bash` shell pinned on Windows for portability.
- `govulncheck` workflow scanning every push and pull request for known Go vulnerabilities.
- CodeQL static-analysis workflow on a weekly schedule plus per-PR runs.
- Dependabot configuration tracking Go modules and GitHub Actions updates.

### Changed

- Renamed Go module from `github.com/przemekperon/golang-infakt` to
  `github.com/przemekperon/infakt-go-sdk` to match the published name on
  pkg.go.dev.
- `doc.go` expanded with new sections: `Stability and Versioning`, `Context and Cancellation`, `Retries`, `Documentation and References` (link to https://docs.infakt.pl).
- `README.md` extended with `Requirements`, `Documentation`, `Versioning`, `Context Support`, `Testing`, `Contributing`, and `Security` sections; configuration and error-handling sections slimmed to point at `pkg.go.dev` for the full reference.
- Adopted Go 1.19+ doc-link syntax (`[Name]`) across the package for cross-references between types, methods, sentinel errors, and stdlib symbols.
- Documented concurrency guarantees on `Client` and clarified `NewClient` defaults.
- Lowered minimum `go` directive in `go.mod` for broader downstream compatibility; pinned latest Go patch version in CI matrix.

### Notes

- No public API changes; this release is documentation, infrastructure, and packaging.

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

[Unreleased]: https://github.com/przemekperon/infakt-go-sdk/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/przemekperon/infakt-go-sdk/releases/tag/v0.2.0
[0.1.1]: https://github.com/przemekperon/infakt-go-sdk/releases/tag/v0.1.1
[0.1.0]: https://github.com/przemekperon/infakt-go-sdk/releases/tag/v0.1.0
