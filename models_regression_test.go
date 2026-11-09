package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestInvoiceUpdateAddressIntent(t *testing.T) {
	original := &LocalGovernmentAddressRequest{CompanyName: "School", Street: "Main", BuildingNumber: "1", City: "Warsaw", TaxID: "0001", DoorNumber: "2", PostCode: "00-001", CountryID: Int64(1)}
	current := original
	services := []ServiceEntryRequest{{Name: String("Old"), TaxSymbol: String("23")}}
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var root struct {
			Invoice InvoiceRequest `json:"invoice"`
		}
		if err := json.NewDecoder(r.Body).Decode(&root); err != nil {
			t.Error(err)
		}
		// Reproduce the documented destructive omission and full-list replacement.
		current = root.Invoice.LocalGovernmentRecipientAddress
		if root.Invoice.Services != nil {
			services = *root.Invoice.Services
		}
		_ = json.NewEncoder(w).Encode(Invoice{Notes: ""})
	}))
	defer ts.Close()
	c := newTestClient("test", WithBaseURL(ts.URL))
	defer c.Close()
	ctx := context.Background()
	if _, err := c.Invoices.UpdateWithRequest(ctx, "1", &InvoiceRequest{Notes: String("")}); err == nil {
		t.Fatal("missing address intent accepted")
	}
	if _, err := c.Invoices.Update(ctx, "1", &Invoice{Notes: "edit", LocalGovernmentRecipientAddress: &LocalGovernmentAddress{}}); err == nil {
		t.Fatal("legacy update accepted")
	}
	if calls != 0 {
		t.Fatalf("ambiguous updates made %d calls", calls)
	}
	if _, err := c.Invoices.UpdateWithRequest(ctx, "1", &InvoiceRequest{Notes: String(""), LocalGovernmentRecipientAddress: original}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(current, original) {
		t.Fatalf("address changed: %#v", current)
	}
	if len(services) != 1 || *services[0].Name != "Old" {
		t.Fatal("omitted services changed old list")
	}
	replacement := []ServiceEntryRequest{{Name: String("New A"), TaxSymbol: String("8")}, {Name: String("New B"), TaxSymbol: String("23")}}
	changed := &LocalGovernmentAddressRequest{CompanyName: "School 2", Street: "Other", BuildingNumber: "3", City: "Cracow"}
	if _, err := c.Invoices.UpdateWithRequest(ctx, "1", &InvoiceRequest{LocalGovernmentRecipientAddress: changed, Services: &replacement}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(current, changed) || !reflect.DeepEqual(services, replacement) {
		t.Fatal("replacement not preserved")
	}
	if _, err := c.Invoices.UpdateWithRequest(ctx, "1", &InvoiceRequest{RemoveLocalGovernmentRecipientAddress: true}); err != nil {
		t.Fatal(err)
	}
	if current != nil {
		t.Fatal("explicit removal retained address")
	}
	if _, err := c.Invoices.UpdateWithRequest(ctx, "1", &InvoiceRequest{RemoveLocalGovernmentRecipientAddress: true}); err != nil {
		t.Fatal(err)
	}
	if current != nil {
		t.Fatal("address unexpectedly created")
	}
	before := calls
	if _, err := c.Invoices.UpdateWithRequest(ctx, "1", &InvoiceRequest{LocalGovernmentRecipientAddress: original, RemoveLocalGovernmentRecipientAddress: true}); err == nil {
		t.Fatal("conflicting intent accepted")
	}
	if calls != before {
		t.Fatal("conflicting intent sent a request")
	}
}

func TestTypedWriteZeroValues(t *testing.T) {
	var captured map[string]json.RawMessage
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var root map[string]map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&root); err != nil {
			t.Error(err)
		}
		for _, v := range root {
			captured = v
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer ts.Close()
	c := newTestClient("test", WithBaseURL(ts.URL))
	defer c.Close()
	assert := func(fields map[string]string, omitted ...string) {
		t.Helper()
		for k, want := range fields {
			if string(captured[k]) != want {
				t.Errorf("%s=%s, want %s", k, captured[k], want)
			}
		}
		for _, k := range omitted {
			if _, ok := captured[k]; ok {
				t.Errorf("unexpected %s", k)
			}
		}
	}
	ctx := context.Background()
	for _, create := range []bool{false, true} {
		req := &InvoiceRequest{Notes: String(""), PaidPrice: Int(0), CheckDuplicateNumber: Bool(false), RemoveLocalGovernmentRecipientAddress: true}
		var err error
		if create {
			_, err = c.Invoices.CreateWithRequest(ctx, req)
		} else {
			_, err = c.Invoices.UpdateWithRequest(ctx, "1", req)
		}
		if err != nil {
			t.Fatal(err)
		}
		assert(map[string]string{"notes": `""`, "paid_price": "0", "check_duplicate_number": "false"}, "currency", "services", "RemoveLocalGovernmentRecipientAddress", "remove_local_government_recipient_address")
		client := &ClientEntityRequest{Note: String(""), DaysToPayment: Int(0), SameForwardAddress: Bool(false)}
		if create {
			_, err = c.Clients.CreateWithRequest(ctx, client)
		} else {
			_, err = c.Clients.UpdateWithRequest(ctx, 1, client)
		}
		if err != nil {
			t.Fatal(err)
		}
		assert(map[string]string{"note": `""`, "days_to_payment": "0", "same_forward_address": "false"}, "company_name")
		bank := &BankAccountRequest{CustomName: String(""), Default: Bool(false)}
		if create {
			_, err = c.BankAccounts.CreateWithRequest(ctx, bank)
		} else {
			_, err = c.BankAccounts.UpdateWithRequest(ctx, 1, bank)
		}
		if err != nil {
			t.Fatal(err)
		}
		assert(map[string]string{"custom_name": `""`, "default": "false"}, "currency")
	}
	// Empty arrays can be serialized; this test does not assert business acceptance.
	empty := []ServiceEntryRequest{}
	if _, err := c.Invoices.CreateWithRequest(ctx, &InvoiceRequest{Services: &empty}); err != nil {
		t.Fatal(err)
	}
	assert(map[string]string{"services": "[]"})
	if _, err := c.Invoices.CreateWithRequest(ctx, &InvoiceRequest{}); err != nil {
		t.Fatal(err)
	}
	assert(nil, "notes", "paid_price", "services")
}

func TestWriteNilArguments(t *testing.T) {
	c := newTestClient("test", WithHTTPClient(&http.Client{Transport: modelRoundTripFunc(func(*http.Request) (*http.Response, error) { t.Error("nil request reached transport"); return nil, nil })}))
	defer c.Close()
	ctx := context.Background()
	cases := []func() error{
		func() error { _, e := c.Invoices.Create(ctx, nil); return e }, func() error { _, e := c.Invoices.Update(ctx, "1", nil); return e },
		func() error { _, e := c.Invoices.CreateWithRequest(ctx, nil); return e }, func() error { _, e := c.Invoices.UpdateWithRequest(ctx, "1", nil); return e },
		func() error { _, e := c.Clients.Create(ctx, nil); return e }, func() error { _, e := c.Clients.Update(ctx, 1, nil); return e },
		func() error { _, e := c.Clients.CreateWithRequest(ctx, nil); return e }, func() error { _, e := c.Clients.UpdateWithRequest(ctx, 1, nil); return e },
		func() error { _, e := c.BankAccounts.Create(ctx, nil); return e }, func() error { _, e := c.BankAccounts.Update(ctx, 1, nil); return e },
		func() error { _, e := c.BankAccounts.CreateWithRequest(ctx, nil); return e }, func() error { _, e := c.BankAccounts.UpdateWithRequest(ctx, 1, nil); return e },
	}
	for i, call := range cases {
		if err := call(); err == nil {
			t.Errorf("case %d accepted nil", i)
		}
	}
}

type modelRoundTripFunc func(*http.Request) (*http.Response, error)

func (f modelRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestInvoiceDocumentMarkingsAndExtensions(t *testing.T) {
	variants := []string{`"document_markings_ids":["001","ABC"],`, `"document_markings_ids":null,`, `"document_markings_ids":[],`, ``}
	extensions := []string{`"extensions":{"payments":{"link":"https://example.test/pay","available":true},"shares":{"link":"https://example.test/share","available":true,"valid_until":"2030-01-01"}},`, `"extensions":null,`, `"extensions":{},`, ``}
	for i, markings := range variants {
		for j, ext := range extensions {
			t.Run(string(rune('A'+i))+string(rune('A'+j)), func(t *testing.T) {
				body := `{` + markings + ext + `"id":1,"local_government_recipient_address":{"company_name":"School","nip":"0001","street":"Main","street_number":"1","flat_number":null,"postal_code":"00-001","city":"Warsaw","country":"PL"}}`
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.HasSuffix(r.URL.Path, "/invoices.json") {
						_, _ = w.Write([]byte(`{"entities":[` + body + `]}`))
					} else {
						_, _ = w.Write([]byte(body))
					}
				}))
				defer ts.Close()
				c := newTestClient("test", WithBaseURL(ts.URL))
				defer c.Close()
				one, err := c.Invoices.Get(context.Background(), "1")
				if err != nil {
					t.Fatal(err)
				}
				list, _, err := c.Invoices.List(context.Background(), nil)
				if err != nil {
					t.Fatal(err)
				}
				for _, v := range []*Invoice{one, &list[0]} {
					if i == 0 && !reflect.DeepEqual(v.DocumentMarkingsIDs, []string{"001", "ABC"}) {
						t.Fatal("markings changed")
					}
					if (i == 1 || i == 3) && v.DocumentMarkingsIDs != nil {
						t.Fatal("absent/null markings should be nil")
					}
					if i == 2 && (v.DocumentMarkingsIDs == nil || len(v.DocumentMarkingsIDs) != 0) {
						t.Fatal("empty markings lost")
					}
					if (j == 1 || j == 3) && v.Extensions != nil {
						t.Fatal("absent/null extensions should be nil")
					}
					if j == 0 {
						if v.Extensions == nil || v.Extensions.Payments == nil || *v.Extensions.Payments.Link != "https://example.test/pay" || !v.Extensions.Payments.Available || v.Extensions.Shares == nil || *v.Extensions.Shares.Link != "https://example.test/share" || !v.Extensions.Shares.Available || *v.Extensions.Shares.ValidUntil != "2030-01-01" {
							t.Fatal("extension links lost")
						}
					}
					if v.LocalGovernmentRecipientAddress == nil || *v.LocalGovernmentRecipientAddress.NIP != "0001" || *v.LocalGovernmentRecipientAddress.StreetNumber != "1" || *v.LocalGovernmentRecipientAddress.Country != "PL" {
						t.Fatal("JST response lost")
					}
				}
			})
		}
	}
}

func TestInvoiceRejectsNullServices(t *testing.T) {
	c := newTestClient("test", WithHTTPClient(&http.Client{Transport: modelRoundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("null services reached transport")
		return nil, nil
	})}))
	defer c.Close()
	var items []ServiceEntryRequest
	if _, err := c.Invoices.CreateWithRequest(context.Background(), &InvoiceRequest{Services: &items}); err == nil {
		t.Fatal("null services accepted")
	}
}

func TestInvoiceRejectsNumericDocumentMarkings(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(`{"document_markings_ids":[1]}`)) }))
	defer ts.Close()
	c := newTestClient("test", WithBaseURL(ts.URL))
	defer c.Close()
	if _, err := c.Invoices.Get(context.Background(), "1"); err == nil {
		t.Fatal("numeric marking accepted")
	}
}
