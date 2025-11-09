package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInvoiceService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v3/invoices.json" {
			t.Errorf("expected path /v3/invoices.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(invoiceListRoot{
			MetaInfo: MetaInfo{Count: 1, TotalCount: 1},
			Entities: []Invoice{
				{ID: 1, Number: "FV/2025/01/001", GrossPrice: 12300},
			},
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	invoices, meta, err := c.Invoices.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(invoices) != 1 {
		t.Errorf("expected 1 invoice, got %d", len(invoices))
	}
	if meta.TotalCount != 1 {
		t.Errorf("expected TotalCount 1, got %d", meta.TotalCount)
	}
	if invoices[0].Number != "FV/2025/01/001" {
		t.Errorf("expected Number %q, got %q", "FV/2025/01/001", invoices[0].Number)
	}
}

func TestInvoiceService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/invoices/99.json" {
			t.Errorf("expected path /v3/invoices/99.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Invoice{
			ID:     99,
			Number: "FV/2025/02/001",
			Services: []ServiceEntry{
				{Name: "Consulting", Quantity: 10, UnitNetPrice: 15000},
			},
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	invoice, err := c.Invoices.Get(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.ID != 99 {
		t.Errorf("expected ID 99, got %d", invoice.ID)
	}
	if len(invoice.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(invoice.Services))
	}
	if invoice.Services[0].Name != "Consulting" {
		t.Errorf("expected service name %q, got %q", "Consulting", invoice.Services[0].Name)
	}
}

func TestInvoiceService_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var root invoiceRoot
		json.NewDecoder(r.Body).Decode(&root)

		if root.Invoice.ClientID != 42 {
			t.Errorf("expected ClientID 42, got %d", root.Invoice.ClientID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Invoice{
			ID:       200,
			ClientID: 42,
			Number:   "FV/2025/03/001",
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	invoice, err := c.Invoices.Create(context.Background(), &Invoice{
		ClientID:    42,
		InvoiceDate: "2025-03-01",
		SaleDate:    "2025-03-01",
		Services: []ServiceEntry{
			{Name: "Development", Quantity: 1, UnitNetPrice: 50000},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.ID != 200 {
		t.Errorf("expected ID 200, got %d", invoice.ID)
	}
}

func TestInvoiceService_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Invoice{
			ID:     99,
			Notes:  "Updated notes",
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	invoice, err := c.Invoices.Update(context.Background(), 99, &Invoice{
		Notes: "Updated notes",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.Notes != "Updated notes" {
		t.Errorf("expected Notes %q, got %q", "Updated notes", invoice.Notes)
	}
}

func TestInvoiceService_Delete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	err := c.Invoices.Delete(context.Background(), 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
