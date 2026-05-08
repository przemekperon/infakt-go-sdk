# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go client library for the InFakt API v3 (Polish invoicing service). Flat package structure under `github.com/przemekperon/infakt-go-sdk`.

## Build & Run Commands

```bash
go build ./...          # build all packages
go test ./...           # run all tests
go test -run TestName   # run a single test by name
go vet ./...            # static analysis
```

## Architecture

- Single `infakt` package (flat structure, no sub-packages)
- `Client` struct with service fields (Invoices, Clients, Products, BankAccounts, VatRates, Gtus, CostInvoices)
- Each resource is a separate `*Service` type with CRUD methods (CostInvoices is read-only)
- All methods accept `context.Context` as first parameter
- Zero external dependencies (pure stdlib)
- API authentication via `X-inFakt-ApiKey` header
- Base URL: `https://api.infakt.pl/v3`
