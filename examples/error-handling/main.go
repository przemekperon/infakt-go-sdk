// Command error-handling demonstrates the three error shapes you'll
// encounter using the infakt-go-sdk:
//
//  1. Sentinel errors (e.g. infakt.ErrNotFound) detected with errors.Is.
//  2. Typed *infakt.ErrorResponse values for structured introspection.
//  3. Standard context errors propagated from cancelled requests.
//
// Run with:
//
//	INFAKT_API_KEY=your-key go run ./examples/error-handling
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	infakt "github.com/przemekperon/infakt-go-sdk"
)

const separator = "------------------------------------------------------------"

func main() {
	apiKey := os.Getenv("INFAKT_API_KEY")
	if apiKey == "" {
		fmt.Println("INFAKT_API_KEY not set — skipping (this is fine for `go vet` / `go build`).")
		return
	}

	client := infakt.NewClient(apiKey, infakt.WithUserAgent("infakt-go-sdk-error-handling/0.1"))
	defer client.Close()

	demoNotFound(client)
	fmt.Println(separator)
	demoTypedError(client)
	fmt.Println(separator)
	demoCancellation(client)
}

// demoNotFound triggers a 404 and recognises it via errors.Is.
func demoNotFound(client *infakt.Client) {
	fmt.Println("== Demo 1: sentinel ErrNotFound ==")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := client.Invoices.Get(ctx, 1)
	switch {
	case err == nil:
		fmt.Println("unexpected: invoice id=1 actually exists.")
	case errors.Is(err, infakt.ErrNotFound):
		fmt.Println("invoice not found, as expected (errors.Is(err, infakt.ErrNotFound) == true)")
	default:
		fmt.Printf("got a different error: %v\n", err)
	}
}

// demoTypedError shows how to extract the structured fields of
// *infakt.ErrorResponse via errors.As.
func demoTypedError(client *infakt.Client) {
	fmt.Println("== Demo 2: typed *ErrorResponse ==")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := client.Invoices.Get(ctx, 1)
	if err == nil {
		fmt.Println("unexpected: invoice id=1 actually exists; nothing to introspect.")
		return
	}

	var apiErr *infakt.ErrorResponse
	if errors.As(err, &apiErr) {
		fmt.Printf("StatusCode: %d\n", apiErr.StatusCode)
		fmt.Printf("Method:     %s\n", apiErr.Method)
		fmt.Printf("Endpoint:   %s\n", apiErr.Endpoint)
		fmt.Printf("Message:    %s\n", apiErr.Message)
		return
	}
	fmt.Printf("error did not unwrap to *infakt.ErrorResponse: %v\n", err)
}

// demoCancellation shows that ctx.Err() propagates through the SDK.
func demoCancellation(client *infakt.Client) {
	fmt.Println("== Demo 3: context cancellation ==")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before issuing the request

	_, _, err := client.Invoices.List(ctx, nil)
	switch {
	case err == nil:
		fmt.Println("unexpected: request succeeded against a cancelled context.")
	case errors.Is(err, context.Canceled):
		fmt.Println("request returned context.Canceled, as expected.")
	default:
		fmt.Printf("got a different error: %v\n", err)
	}
}
