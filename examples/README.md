# Examples

Runnable example programs demonstrating common usage of `infakt-go-sdk`.

Each program reads `INFAKT_API_KEY` from the environment. Without the key, every program prints a friendly message and exits — so `go build ./examples/...` and `go vet ./...` work without credentials.

## Running

    INFAKT_API_KEY=your-key go run ./examples/quickstart

## Programs

| Program | What it demonstrates |
|---------|----------------------|
| [quickstart](./quickstart) | End-to-end usage: client setup, listing invoices, fetching the next invoice number |
| [pagination](./pagination) | Paging through every invoice with `Limit`/`Offset` and `MetaInfo.TotalCount` |
| [pdf-export](./pdf-export) | Downloading invoice PDFs and saving them to disk |
| [error-handling](./error-handling) | Sentinel errors (`ErrNotFound`), typed errors (`ErrorResponse`), and context cancellation |

## Notes

- These programs perform real API calls when a key is provided. `quickstart` and `error-handling` are read-only / safe; `pdf-export` only reads. None of them create data.
- The PDF program writes to the current working directory.
