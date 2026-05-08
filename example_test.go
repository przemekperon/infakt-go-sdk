package infakt_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	infakt "github.com/przemekperon/infakt-go-sdk"
)

// newMockClient creates an *infakt.Client wired to an httptest.Server
// running the supplied handler. Rate limiting is disabled so Examples
// run instantly. Callers must defer both srv.Close() and client.Close().
func newMockClient(handler http.HandlerFunc) (*infakt.Client, *httptest.Server) {
	srv := httptest.NewServer(handler)
	client := infakt.NewClient("test-key",
		infakt.WithBaseURL(srv.URL),
		infakt.WithRateLimit(0),
	)
	return client, srv
}

// jsonHandler returns an http.HandlerFunc that always replies 200 OK with
// the given JSON body and Content-Type: application/json.
func jsonHandler(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, body)
	}
}

// -----------------------------------------------------------------------------
// Constructor & lifecycle examples
// -----------------------------------------------------------------------------

func ExampleNewClient() {
	client, srv := newMockClient(jsonHandler(`{"next_number":"1/2026"}`))
	defer srv.Close()
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "vat")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("next:", num)
	// Output: next: 1/2026
}

func ExampleNewClient_withOptions() {
	client, srv := newMockClient(jsonHandler(`{"next_number":"42/2026"}`))
	defer srv.Close()
	defer client.Close()

	// Stack multiple options on top of those already set by newMockClient.
	client = infakt.NewClient("test-key",
		infakt.WithBaseURL(srv.URL),
		infakt.WithRateLimit(0),
		infakt.WithUserAgent("my-app/1.0"),
		infakt.WithHTTPClient(&http.Client{Timeout: 5 * time.Second}),
	)
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num)
	// Output: 42/2026
}

func ExampleNewClient_withBaseURL() {
	srv := httptest.NewServer(jsonHandler(`{"next_number":"baseurl-ok"}`))
	defer srv.Close()

	client := infakt.NewClient("test-key",
		infakt.WithBaseURL(srv.URL),
		infakt.WithRateLimit(0),
	)
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num)
	// Output: baseurl-ok
}

func ExampleNewClient_withHTTPClient() {
	srv := httptest.NewServer(jsonHandler(`{"next_number":"http-client-ok"}`))
	defer srv.Close()

	httpClient := &http.Client{Timeout: 10 * time.Second}
	client := infakt.NewClient("test-key",
		infakt.WithBaseURL(srv.URL),
		infakt.WithHTTPClient(httpClient),
		infakt.WithRateLimit(0),
	)
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num)
	// Output: http-client-ok
}

func ExampleNewClient_withRateLimit() {
	srv := httptest.NewServer(jsonHandler(`{"next_number":"rate-limit-ok"}`))
	defer srv.Close()

	// Pass 0 to disable rate limiting (useful in tests).
	client := infakt.NewClient("test-key",
		infakt.WithBaseURL(srv.URL),
		infakt.WithRateLimit(0),
	)
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num)
	// Output: rate-limit-ok
}

func ExampleNewClient_withUserAgent() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"next_number":%q}`, r.Header.Get("User-Agent"))
	}))
	defer srv.Close()

	client := infakt.NewClient("test-key",
		infakt.WithBaseURL(srv.URL),
		infakt.WithRateLimit(0),
		infakt.WithUserAgent("my-app/1.0"),
	)
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num)
	// Output: my-app/1.0
}

func ExampleClient_Close() {
	client, srv := newMockClient(jsonHandler(`{"next_number":"closed-ok"}`))
	defer srv.Close()
	// Idiomatic: defer Close immediately after construction so the
	// rate-limiter ticker is released when the function returns.
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num)
	// Output: closed-ok
}

// -----------------------------------------------------------------------------
// Invoice service examples
// -----------------------------------------------------------------------------

func ExampleInvoiceService_List() {
	body := `{
		"metainfo":{"count":2,"total_count":2,"next":"","previous":""},
		"entities":[
			{"id":1,"number":"FV/1/2026","gross_price":12300},
			{"id":2,"number":"FV/2/2026","gross_price":45600}
		]
	}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	invoices, meta, err := client.Invoices.List(context.Background(), &infakt.InvoiceListOptions{
		ListOptions: infakt.ListOptions{Limit: 10},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d invoices (total: %d)\n", len(invoices), meta.TotalCount)
	for _, inv := range invoices {
		fmt.Printf("  %s - %d\n", inv.Number, inv.GrossPrice)
	}
	// Output:
	// Found 2 invoices (total: 2)
	//   FV/1/2026 - 12300
	//   FV/2/2026 - 45600
}

func ExampleInvoiceService_Get() {
	body := `{"id":7,"number":"FV/7/2026","gross_price":9999,"status":"sent"}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	inv, err := client.Invoices.Get(context.Background(), 7)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d %s %s %d\n", inv.ID, inv.Number, inv.Status, inv.GrossPrice)
	// Output: 7 FV/7/2026 sent 9999
}

func ExampleInvoiceService_Create() {
	body := `{"id":101,"number":"FV/101/2026","status":"draft","gross_price":5000}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	created, err := client.Invoices.Create(context.Background(), &infakt.Invoice{
		ClientID: 42,
		Currency: "PLN",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created %d %s\n", created.ID, created.Number)
	// Output: created 101 FV/101/2026
}

func ExampleInvoiceService_Update() {
	body := `{"id":5,"number":"FV/5/2026","notes":"updated note"}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	updated, err := client.Invoices.Update(context.Background(), 5, &infakt.Invoice{
		Notes: "updated note",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated %d %s\n", updated.ID, updated.Notes)
	// Output: updated 5 updated note
}

func ExampleInvoiceService_Delete() {
	client, srv := newMockClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	defer client.Close()

	if err := client.Invoices.Delete(context.Background(), 5); err != nil {
		log.Fatal(err)
	}
	fmt.Println("deleted")
	// Output: deleted
}

func ExampleInvoiceService_MarkAsPaid() {
	client, srv := newMockClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer client.Close()

	if err := client.Invoices.MarkAsPaid(context.Background(), 5, "2026-05-08"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("paid")
	// Output: paid
}

func ExampleInvoiceService_SendByEmail() {
	client, srv := newMockClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer client.Close()

	if err := client.Invoices.SendByEmail(context.Background(), 5, "client@example.com"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("sent")
	// Output: sent
}

func ExampleInvoiceService_GetPDF() {
	pdfBytes := []byte("%PDF-1.4 fake pdf body bytes")
	client, srv := newMockClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdfBytes)
	}))
	defer srv.Close()
	defer client.Close()

	data, err := client.Invoices.GetPDF(context.Background(), 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("got %d bytes\n", len(data))
	// Output: got 28 bytes
}

func ExampleInvoiceService_GetNextNumber() {
	client, srv := newMockClient(jsonHandler(`{"next_number":"FV/3/2026"}`))
	defer srv.Close()
	defer client.Close()

	num, err := client.Invoices.GetNextNumber(context.Background(), "vat")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(num)
	// Output: FV/3/2026
}

// -----------------------------------------------------------------------------
// Other service examples
// -----------------------------------------------------------------------------

func ExampleClientEntityService_Create() {
	body := `{"id":501,"company_name":"ACME Sp. z o.o.","nip":"1234567890"}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	entity, err := client.Clients.Create(context.Background(), &infakt.ClientEntity{
		CompanyName: "ACME Sp. z o.o.",
		NIP:         "1234567890",
		Street:      "ul. Testowa 1",
		City:        "Warszawa",
		PostalCode:  "00-001",
		Country:     "PL",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created client: %s (ID: %d)\n", entity.CompanyName, entity.ID)
	// Output: Created client: ACME Sp. z o.o. (ID: 501)
}

func ExampleProductService_List() {
	body := `{
		"metainfo":{"count":2,"total_count":2,"next":"","previous":""},
		"entities":[
			{"id":1,"name":"Consulting","unit_net_price":15000},
			{"id":2,"name":"Hosting","unit_net_price":2500}
		]
	}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	products, _, err := client.Products.List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range products {
		fmt.Printf("%s - %d PLN\n", p.Name, p.UnitNetPrice)
	}
	// Output:
	// Consulting - 15000 PLN
	// Hosting - 2500 PLN
}

func ExampleBankAccountService_List() {
	body := `{
		"metainfo":{"count":1,"total_count":1,"next":"","previous":""},
		"entities":[
			{"id":1,"bank_name":"mBank","account_number":"PL00 0000 0000 0000 0000 0000 0000","currency":"PLN","default":true}
		]
	}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	accounts, _, err := client.BankAccounts.List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, a := range accounts {
		fmt.Printf("%s %s default=%t\n", a.BankName, a.Currency, a.Default)
	}
	// Output: mBank PLN default=true
}

func ExampleVatRateService_List() {
	body := `{
		"metainfo":{"count":2,"total_count":2,"next":"","previous":""},
		"entities":[
			{"id":1,"name":"23%","rate":"23","symbol":"vat_23"},
			{"id":2,"name":"8%","rate":"8","symbol":"vat_8"}
		]
	}`
	client, srv := newMockClient(jsonHandler(body))
	defer srv.Close()
	defer client.Close()

	rates, _, err := client.VatRates.List(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range rates {
		fmt.Printf("%s (%s%%)\n", r.Symbol, r.Rate)
	}
	// Output:
	// vat_23 (23%)
	// vat_8 (8%)
}

// -----------------------------------------------------------------------------
// Patterns
// -----------------------------------------------------------------------------

func Example_pagination() {
	// Total of 5 invoices, paginated 2 at a time.
	all := []string{"A", "B", "C", "D", "E"}

	handler := func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 {
			limit = len(all)
		}

		end := offset + limit
		if end > len(all) {
			end = len(all)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"metainfo":{"count":`)
		_, _ = fmt.Fprintf(w, "%d", end-offset)
		_, _ = fmt.Fprint(w, `,"total_count":`)
		_, _ = fmt.Fprintf(w, "%d", len(all))
		_, _ = fmt.Fprint(w, `,"next":"","previous":""},"entities":[`)
		for i := offset; i < end; i++ {
			if i > offset {
				_, _ = fmt.Fprint(w, ",")
			}
			_, _ = fmt.Fprintf(w, `{"id":%d,"number":%q}`, i+1, all[i])
		}
		_, _ = fmt.Fprint(w, `]}`)
	}

	client, srv := newMockClient(handler)
	defer srv.Close()
	defer client.Close()

	const pageSize = 2
	offset := 0
	page := 0
	for {
		page++
		invoices, meta, err := client.Invoices.List(context.Background(), &infakt.InvoiceListOptions{
			ListOptions: infakt.ListOptions{Offset: offset, Limit: pageSize},
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("page %d: %d items (of %d total)\n", page, len(invoices), meta.TotalCount)
		offset += len(invoices)
		if offset >= meta.TotalCount || len(invoices) == 0 {
			break
		}
	}
	// Output:
	// page 1: 2 items (of 5 total)
	// page 2: 2 items (of 5 total)
	// page 3: 1 items (of 5 total)
}

func Example_errorHandling() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message":"invoice not found"}`)
	}
	client, srv := newMockClient(handler)
	defer srv.Close()
	defer client.Close()

	_, err := client.Invoices.Get(context.Background(), 999)
	if err == nil {
		log.Fatal("expected error")
	}

	// Compare against sentinel via errors.Is.
	if errors.Is(err, infakt.ErrNotFound) {
		fmt.Println("matched ErrNotFound")
	}

	// Inspect rich details via errors.As.
	var apiErr *infakt.ErrorResponse
	if errors.As(err, &apiErr) {
		fmt.Printf("status=%d message=%s\n", apiErr.StatusCode, apiErr.Message)
	}
	// Output:
	// matched ErrNotFound
	// status=404 message=invoice not found
}

func Example_contextTimeout() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the caller's context deadline so the
		// request is canceled before the body is written.
		select {
		case <-time.After(500 * time.Millisecond):
		case <-r.Context().Done():
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"next_number":"too-late"}`)
	}
	client, srv := newMockClient(handler)
	defer srv.Close()
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Invoices.GetNextNumber(ctx, "")
	if errors.Is(err, context.DeadlineExceeded) {
		fmt.Println("deadline exceeded")
	}
	// Output: deadline exceeded
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func ExampleString() {
	name := infakt.String("Test Product")
	fmt.Println(*name)
	// Output: Test Product
}

func ExampleInt() {
	price := infakt.Int(10000)
	fmt.Println(*price)
	// Output: 10000
}

func ExampleInt64() {
	id := infakt.Int64(9223372036854775807)
	fmt.Println(*id)
	// Output: 9223372036854775807
}

func ExampleBool() {
	flag := infakt.Bool(true)
	fmt.Println(*flag)
	// Output: true
}

func ExampleFloat64() {
	qty := infakt.Float64(2.5)
	fmt.Println(*qty)
	// Output: 2.5
}

func ExampleTime() {
	t := infakt.Time(time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC))
	fmt.Println(t.Format(time.RFC3339))
	// Output: 2026-05-08T12:00:00Z
}
