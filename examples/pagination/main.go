// Command pagination demonstrates paging through every invoice in the
// account using ListOptions{Offset, Limit} and MetaInfo.TotalCount.
//
// Run with:
//
//	INFAKT_API_KEY=your-key go run ./examples/pagination
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	infakt "github.com/przemekperon/infakt-go-sdk"
)

const pageSize = 25

func main() {
	apiKey := os.Getenv("INFAKT_API_KEY")
	if apiKey == "" {
		fmt.Println("INFAKT_API_KEY not set — skipping (this is fine for `go vet` / `go build`).")
		return
	}

	client := infakt.NewClient(apiKey, infakt.WithUserAgent("infakt-go-sdk-pagination/0.1"))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var (
		all     []infakt.Invoice
		offset  int
		page    int
		total   int
		fetched bool
	)

	for {
		page++
		invoices, meta, err := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
			ListOptions: infakt.ListOptions{Offset: offset, Limit: pageSize},
		})
		if err != nil {
			log.Fatalf("list page %d (offset=%d): %v", page, offset, err)
		}

		if !fetched {
			total = meta.TotalCount
			fetched = true
		}

		all = append(all, invoices...)
		fmt.Printf("Page %d: got %d (running total %d of %d)\n",
			page, len(invoices), len(all), total)

		// Stop conditions: empty page, or we've collected everything.
		if len(invoices) == 0 {
			break
		}
		if total > 0 && len(all) >= total {
			break
		}

		offset += len(invoices)
	}

	fmt.Printf("\nDone. Collected %d invoice(s) across %d page(s).\n", len(all), page)
}
