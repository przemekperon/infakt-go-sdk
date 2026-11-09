package infakt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testCostUUID = "57639119-f155-42c3-b2c5-cba29409978d"

// realisticListBody mirrors the JSON envelope returned by the live
// `GET /v3/documents/costs.json` endpoint, captured against production for
// pinned fixture coverage.
const realisticListBody = `{
  "metainfo":{"count":2,"total_count":2377},
  "entities":[
    {
      "uuid":"57639119-f155-42c3-b2c5-cba29409978d",
      "number":"2026-05-0047463-Z",
      "net_price":4309,
      "gross_price":5300,
      "tax_price":991,
      "currency":"PLN",
      "accounted_at":null,
      "issue_date":"2026-05-05",
      "received_date":"2026-05-05",
      "due_date":"2026-05-04",
      "seller_name":"VIKINGCO POLAND Sp. z o.o.",
      "seller_tax_code":"8971793639",
      "seller_bank_account":"",
      "description":null,
      "added_by":"",
      "category":"Wydatki przedsiębiorcy",
      "kind":"document_scan",
      "notes_count":0,
      "duplicated":false,
      "created_at":"2026-05-05 15:31:58 +0200",
      "source":"ksef",
      "reconciliation_id":null,
      "statuses":[{"symbol":"cost_not_accounted","name":"Nie zaksięgowano","group":"cost_accounting"}]
    },
    {
      "uuid":"00000000-0000-0000-0000-000000000002",
      "number":"2026-05-0047464-Z",
      "currency":"PLN",
      "issue_date":"2026-05-04",
      "received_date":"2026-05-04",
      "due_date":"2026-05-04",
      "kind":"document_scan",
      "duplicated":false,
      "created_at":"2026-05-04 12:00:00 +0200",
      "source":"upload",
      "seller_name":null,
      "seller_tax_code":null,
      "seller_bank_account":null,
      "description":null,
      "category":null,
      "accounted_at":null,
      "reconciliation_id":null,
      "statuses":[]
    }
  ]
}`

func TestCostInvoiceService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v3/documents/costs.json" {
			t.Errorf("expected /v3/documents/costs.json, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(realisticListBody))
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	costs, meta, err := c.CostInvoices.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Count != 2 {
		t.Errorf("expected Count=2, got %d", meta.Count)
	}
	if meta.TotalCount != 2377 {
		t.Errorf("expected TotalCount=2377, got %d", meta.TotalCount)
	}
	if len(costs) != 2 {
		t.Fatalf("expected 2 costs, got %d", len(costs))
	}

	first := costs[0]
	if first.UUID != testCostUUID {
		t.Errorf("expected UUID %q, got %q", testCostUUID, first.UUID)
	}
	if first.Number != "2026-05-0047463-Z" {
		t.Errorf("expected Number, got %q", first.Number)
	}
	if first.NetPrice == nil || *first.NetPrice != 4309 {
		t.Errorf("expected NetPrice=4309, got %v", first.NetPrice)
	}
	if first.GrossPrice == nil || *first.GrossPrice != 5300 {
		t.Errorf("expected GrossPrice=5300, got %v", first.GrossPrice)
	}
	if first.TaxPrice == nil || *first.TaxPrice != 991 {
		t.Errorf("expected TaxPrice=991, got %v", first.TaxPrice)
	}
	if first.SellerName == nil || *first.SellerName != "VIKINGCO POLAND Sp. z o.o." {
		t.Errorf("expected SellerName populated, got %v", first.SellerName)
	}
	if first.SellerTaxCode == nil || *first.SellerTaxCode != "8971793639" {
		t.Errorf("expected SellerTaxCode 8971793639, got %v", first.SellerTaxCode)
	}
	if first.Category == nil || *first.Category != "Wydatki przedsiębiorcy" {
		t.Errorf("expected Category, got %v", first.Category)
	}
	if first.Description != nil {
		t.Errorf("expected Description nil, got %q", *first.Description)
	}
	if first.AccountedAt != nil {
		t.Errorf("expected AccountedAt nil, got %q", *first.AccountedAt)
	}
	if first.ReconciliationID != nil {
		t.Errorf("expected ReconciliationID nil, got %v", *first.ReconciliationID)
	}
	if len(first.Statuses) != 1 || first.Statuses[0].Symbol != "cost_not_accounted" {
		t.Errorf("expected one status with symbol cost_not_accounted, got %+v", first.Statuses)
	}
	if first.Source != "ksef" {
		t.Errorf("expected Source=ksef, got %q", first.Source)
	}

	// Second entity should round-trip nullable fields as nil pointers.
	second := costs[1]
	if second.SellerName != nil {
		t.Errorf("expected SellerName nil on second entity, got %q", *second.SellerName)
	}
	if second.Category != nil {
		t.Errorf("expected Category nil on second entity, got %q", *second.Category)
	}
}

func TestCostInvoiceService_List_Pagination(t *testing.T) {
	hits := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		offset := r.URL.Query().Get("offset")
		limit := r.URL.Query().Get("limit")
		if limit != "2" {
			t.Errorf("expected limit=2, got %q", limit)
		}

		w.Header().Set("Content-Type", "application/json")

		switch offset {
		case "", "0":
			_ = json.NewEncoder(w).Encode(costInvoiceListRoot{
				MetaInfo: MetaInfo{Count: 2, TotalCount: 3},
				Entities: []CostInvoice{
					{UUID: "u1", Number: "n1"},
					{UUID: "u2", Number: "n2"},
				},
			})
		case "2":
			_ = json.NewEncoder(w).Encode(costInvoiceListRoot{
				MetaInfo: MetaInfo{Count: 1, TotalCount: 3},
				Entities: []CostInvoice{
					{UUID: "u3", Number: "n3"},
				},
			})
		default:
			t.Errorf("unexpected offset %q", offset)
		}
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))

	var all []CostInvoice
	page := 0
	for {
		opts := &CostInvoiceListOptions{
			ListOptions: ListOptions{Offset: page * 2, Limit: 2},
		}
		costs, meta, err := c.CostInvoices.List(context.Background(), opts)
		if err != nil {
			t.Fatalf("unexpected error on page %d: %v", page, err)
		}
		all = append(all, costs...)
		if len(all) >= meta.TotalCount {
			break
		}
		page++
		if page > 10 {
			t.Fatal("pagination did not terminate")
		}
	}

	if len(all) != 3 {
		t.Errorf("expected 3 entities across pages, got %d", len(all))
	}
	if hits != 2 {
		t.Errorf("expected 2 server hits, got %d", hits)
	}
}

func TestCostInvoiceService_List_Filters(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if got := q.Get("q[issue_date_gteq]"); got != "2026-01-01" {
			t.Errorf("expected q[issue_date_gteq]=2026-01-01, got %q", got)
		}
		if got := q.Get("q[issue_date_lteq]"); got != "2026-12-31" {
			t.Errorf("expected q[issue_date_lteq]=2026-12-31, got %q", got)
		}
		if got := q.Get("q[seller_tax_code_eq]"); got != "8971793639" {
			t.Errorf("expected q[seller_tax_code_eq]=8971793639, got %q", got)
		}
		if got := q.Get("q[currency_eq]"); got != "PLN" {
			t.Errorf("expected q[currency_eq]=PLN, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(costInvoiceListRoot{
			MetaInfo: MetaInfo{Count: 0, TotalCount: 0},
			Entities: []CostInvoice{},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	_, _, err := c.CostInvoices.List(context.Background(), &CostInvoiceListOptions{
		DateFrom:      "2026-01-01",
		DateTo:        "2026-12-31",
		SellerTaxCode: "8971793639",
		Currency:      "PLN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCostInvoiceService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/v3/documents/costs/" + testCostUUID + ".json"
		if r.URL.Path != expected {
			t.Errorf("expected path %q, got %q", expected, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CostInvoice{
			UUID:   testCostUUID,
			Number: "2026-05-0047463-Z",
			Source: "ksef",
			Kind:   "document_scan",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	cost, err := c.CostInvoices.Get(context.Background(), testCostUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cost.UUID != testCostUUID {
		t.Errorf("expected UUID %q, got %q", testCostUUID, cost.UUID)
	}
	if cost.Number != "2026-05-0047463-Z" {
		t.Errorf("unexpected Number %q", cost.Number)
	}
}

func TestCostInvoiceService_Get_URLEscaping(t *testing.T) {
	// A UUID containing characters that require percent-encoding ensures
	// we always go through url.PathEscape.
	weirdUUID := "abc 123/with?bad#chars"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The server-side unescaped path should match the original UUID.
		want := "/v3/documents/costs/" + weirdUUID + ".json"
		if r.URL.Path != want {
			t.Errorf("expected unescaped path %q, got %q", want, r.URL.Path)
		}
		// And the raw path should contain the percent-encoded form.
		if !strings.Contains(r.URL.EscapedPath(), "%20") {
			t.Errorf("expected raw path to contain %%20, got %q", r.URL.EscapedPath())
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CostInvoice{UUID: weirdUUID})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	if _, err := c.CostInvoices.Get(context.Background(), weirdUUID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCostInvoiceService_Get_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error":"Resource not found"}`)
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	_, err := c.CostInvoices.Get(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected errors.Is(err, ErrNotFound) to be true, got %v", err)
	}

	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *ErrorResponse, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected StatusCode 404, got %d", apiErr.StatusCode)
	}
	if apiErr.Message != "Resource not found" {
		t.Errorf("expected message %q, got %q", "Resource not found", apiErr.Message)
	}
}
