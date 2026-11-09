// Command pdf-export downloads an invoice PDF and writes it to the
// current working directory.
//
// Run with:
//
//	INFAKT_API_KEY=your-key go run ./examples/pdf-export [invoice-uuid]
//
// If invoice-uuid is omitted, the program lists the first invoice in the
// account and uses its UUID.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

	uuid, err := resolveInvoiceUUID(ctx, client)
	if err != nil {
		log.Fatalf("resolve invoice uuid: %v", err)
	}
	if uuid == "" {
		fmt.Println("no invoices available to export — nothing to do.")
		return
	}

	fmt.Printf("requesting PDF for invoice %s ...\n", uuid)
	pdf, err := client.Invoices.GetPDF(ctx, uuid, infakt.PDFDocumentTypeOriginal, "")
	if err != nil {
		log.Fatalf("get pdf for invoice %s: %v", uuid, err)
	}

	path := fmt.Sprintf("invoice-%s.pdf", uuid)
	if err := os.WriteFile(path, pdf, 0o644); err != nil {
		log.Fatalf("write %s: %v", path, err)
	}

	fmt.Printf("wrote %d bytes to %s\n", len(pdf), path)
}

// resolveInvoiceUUID returns the invoice UUID supplied as os.Args[1], or
// falls back to the first invoice from the account. Returns "" if the
// account has no invoices.
func resolveInvoiceUUID(ctx context.Context, client *infakt.Client) (string, error) {
	if len(os.Args) > 1 && os.Args[1] != "" {
		return os.Args[1], nil
	}

	fmt.Println("no invoice uuid supplied; looking up the most recent invoice ...")
	invoices, _, err := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
		ListOptions: infakt.ListOptions{Limit: 1},
	})
	if err != nil {
		return "", err
	}
	if len(invoices) == 0 {
		return "", nil
	}
	return invoices[0].UUID, nil
}
