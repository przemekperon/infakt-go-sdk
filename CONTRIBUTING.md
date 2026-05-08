# Contributing

Thanks for your interest in improving `infakt-go-sdk`. This document covers the basics of working in this repository.

## Prerequisites

- Go 1.19+ (the package documentation uses `# Heading` godoc syntax introduced in Go 1.19; CI exercises a wider matrix)
- `golangci-lint` for static analysis (configuration lives in [.golangci.yml](.golangci.yml))
- `gofmt` (ships with the Go toolchain)

No other tooling is required: the SDK has zero external runtime dependencies and the test suite uses only the standard library.

## Local Development

From the repository root:

```bash
go build ./...        # build everything
go test ./...         # run unit tests and runnable examples
go vet ./...          # standard static analysis
gofmt -s -w .         # format and simplify
golangci-lint run     # run the linters configured in .golangci.yml
```

Please make sure all of the above succeed before opening a pull request.

## Code Style

- Idiomatic Go: prefer the standard library, avoid clever tricks, follow the patterns already used in the package.
- Every exported identifier must have a godoc comment.
- Resource types (`Invoice`, `ClientEntity`, `Product`, etc.) should keep field-level godoc comments where the field's meaning isn't obvious from its name and JSON tag.
- Keep the package flat (no sub-packages) — this is a deliberate design choice.
- Run `gofmt -s -w .` and `goimports` (the repository uses `local-prefixes: github.com/przemekperon/infakt-go-sdk`).

## Dependencies

This SDK has a **zero external runtime dependency** policy. Do not add modules to `go.mod` (other than `golang.org/x/...` if absolutely required, and only after discussion in an issue first). Test-only dependencies are similarly discouraged — the existing tests use `net/http/httptest` and `testing` only.

## Testing

- Unit tests live in `*_test.go` next to the code they cover. They spin up an `httptest.NewServer` and verify request/response handling — they never call the real inFakt API.
- Runnable examples (`Example*` functions) live in `example_test.go` and double as user-facing documentation on pkg.go.dev.
- `_live_test.go` is a manual smoke-test program guarded by `//go:build ignore`. It is **not** part of `go test ./...` and must be invoked explicitly with `go run _live_test.go` and a real API key in `INFAKT_KEY`. Do not point it at production data.

When fixing a bug, add a regression test that fails before your change.

## Commit Messages

Match the existing convention in `git log`:

- Short, imperative subject line (e.g. `Fix resource leaks in retry logic, add Close(), doRaw() and linter fixes`).
- Group related changes in one commit; avoid noisy "wip" commits in the final history.
- Reference issues or PRs in the body when useful.

A quick `git log --oneline -20` is the source of truth for the local style.

## Pull Request Process

1. Branch from `master`.
2. Make your change with tests and updated documentation as needed.
3. Run `go build ./...`, `go test ./...`, `go vet ./...`, `gofmt -l .` (should be empty), and `golangci-lint run`.
4. Open a pull request against `master`. Describe the change and link to any related issues.
5. CI must be green. Expect review feedback; please be patient and responsive.

## Reporting Bugs and Requesting Features

Open a GitHub issue with:

- The Go version (`go version`) and operating system you saw the problem on.
- The smallest reproduction you can produce.
- For API behavior questions, the inFakt endpoint and (redacted) request/response involved.

For security-sensitive reports, please follow [SECURITY.md](SECURITY.md) instead of opening a public issue.
