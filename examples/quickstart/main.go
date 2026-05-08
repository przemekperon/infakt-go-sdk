// Command quickstart demonstrates end-to-end usage of the infakt-go-sdk:
// constructing a client, listing recent invoices, and fetching the next
// invoice number. It does not create or modify any data.
//
// Run with:
//
//	INFAKT_API_KEY=your-key go run ./examples/quickstart
package main

import (
	"context"
	"errors"
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

	client := infakt.NewClient(apiKey, infakt.WithUserAgent("infakt-go-sdk-quickstart/0.1"))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 1: list the 5 most-recent invoices.
	fmt.Println("== Step 1: list invoices ==")
	invoices, meta, err := client.Invoices.List(ctx, &infakt.InvoiceListOptions{
		ListOptions: infakt.ListOptions{Limit: 5},
	})
	if err != nil {
		log.Printf("list invoices: %v", err)
	} else {
		fmt.Printf("got %d invoice(s) (total in account: %d)\n", len(invoices), meta.TotalCount)
		if len(invoices) > 0 {
			first := invoices[0]
			fmt.Printf("first invoice: id=%d number=%q gross=%d grosze (%.2f PLN)\n",
				first.ID, first.Number, first.GrossPrice, float64(first.GrossPrice)/100.0)
		}
	}

	// Step 2: fetch the next invoice number for kind "vat".
	fmt.Println("\n== Step 2: next invoice number (kind=vat) ==")
	next, err := client.Invoices.GetNextNumber(ctx, "vat")
	if err != nil {
		// Auth or transport problems are useful to surface.
		var apiErr *infakt.ErrorResponse
		if errors.As(err, &apiErr) {
			log.Printf("next number: API error %d: %s", apiErr.StatusCode, apiErr.Message)
		} else {
			log.Printf("next number: %v", err)
		}
	} else {
		fmt.Printf("next number: %s\n", next)
	}

	// Step 3: show what a Create call would look like (without actually
	// creating). The InvoiceRequest type uses pointer fields; the helper
	// constructors String, Int, Int64, Bool, Float64 are exposed by the
	// SDK for that purpose.
	fmt.Println("\n== Step 3: sample InvoiceRequest (not sent) ==")
	today := time.Now().Format("2006-01-02")
	sample := &infakt.InvoiceRequest{
		Kind:          infakt.String("vat"),
		PaymentMethod: infakt.String("transfer"),
		InvoiceDate:   infakt.String(today),
		SaleDate:      infakt.String(today),
		ClientID:      infakt.Int64(123456),
		Services: []infakt.ServiceEntryRequest{
			{
				Name:         infakt.String("Consulting"),
				TaxSymbol:    infakt.String("23"),
				Unit:         infakt.String("hour"),
				Quantity:     infakt.Float64(8),
				UnitNetPrice: infakt.Int(15000), // 150.00 PLN
			},
		},
	}
	fmt.Printf("would create vat invoice for client %d with %d service line(s) on %s\n",
		*sample.ClientID, len(sample.Services), *sample.InvoiceDate)
	fmt.Println("(skipped client.Invoices.Create to avoid side effects in a demo)")
}
