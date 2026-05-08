// Command pdf-export downloads an invoice PDF and writes it to the
// current working directory.
//
// Run with:
//
//	INFAKT_API_KEY=your-key go run ./examples/pdf-export [invoice-id]
//
// If invoice-id is omitted, the program lists the first invoice in the
// account and uses its ID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	infakt "github.com/przemekperon/infakt-go-sdk"
)

func main() {
	apiKey := os.Getenv("INFAKT_API_KEY")
	if apiKey == "" {
		fmt.Println("INFAKT_API_KEY not set — skipping (this is fine for `go vet` / `go build`).")
		return
	}

	client := infakt.NewClient(apiKey, infakt.WithUserAgent("infakt-go-sdk-pdf-export/0.1"))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	id, err := resolveInvoiceID(ctx, client)
	if err != nil {
		log.Fatalf("resolve invoice id: %v", err)
	}
	if id == 0 {
		fmt.Println("no invoices available to export — nothing to do.")
		return
	}

	fmt.Printf("downloading PDF for invoice %d ...\n", id)
	pdf, err := client.Invoices.GetPDF(ctx, id)
	if err != nil {
		log.Fatalf("get pdf for invoice %d: %v", id, err)
	}

	path := fmt.Sprintf("invoice-%d.pdf", id)
	if err := os.WriteFile(path, pdf, 0o644); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}

	fmt.Printf("wrote %d bytes to %s\n", len(pdf), path)
}

// resolveInvoiceID returns the invoice ID supplied as os.Args[1], or
// falls back to the first invoice from the account. Returns 0 if the
// account has no invoices.
func resolveInvoiceID(ctx context.Context, client *infakt.Client) (int64, error) {
	if len(os.Args) > 1 && os.Args[1] != "" {
		id, err := strconv.ParseInt(os.Args[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid invoice id %q: %w", os.Args[1], err)
		}
		return id, nil
	}

	fmt.Println("no invoice id supplied; looking up the most recent invoice ...")
	invoices, _, err := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
		ListOptions: infakt.ListOptions{Limit: 1},
	})
	if err != nil {
		return 0, err
	}
	if len(invoices) == 0 {
		return 0, nil
	}
	return invoices[0].ID, nil
}
