package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testInvoiceUUID = "abc123def4567890"

func TestInvoiceService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v3/invoices.json" {
			t.Errorf("expected path /v3/invoices.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(invoiceListRoot{
			MetaInfo: MetaInfo{Count: 1, TotalCount: 1},
			Entities: []Invoice{
				{ID: 1, UUID: testInvoiceUUID, Number: "FV/2025/01/001", GrossPrice: 12300},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
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
	if invoices[0].UUID != testInvoiceUUID {
		t.Errorf("expected UUID %q, got %q", testInvoiceUUID, invoices[0].UUID)
	}
}

func TestInvoiceService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/invoices/"+testInvoiceUUID+".json" {
			t.Errorf("expected path with UUID, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Invoice{
			ID:     99,
			UUID:   testInvoiceUUID,
			Number: "FV/2025/02/001",
			Services: []ServiceEntry{
				{Name: "Consulting", Quantity: 10, UnitNetPrice: 15000, Discount: "5.0"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	invoice, err := c.Invoices.Get(context.Background(), testInvoiceUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if invoice.UUID != testInvoiceUUID {
		t.Errorf("expected UUID %q, got %q", testInvoiceUUID, invoice.UUID)
	}
	if len(invoice.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(invoice.Services))
	}
	if invoice.Services[0].Discount != "5.0" {
		t.Errorf("expected Discount %q, got %q", "5.0", invoice.Services[0].Discount)
	}
}

func TestInvoiceService_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var root invoiceRoot
		_ = json.NewDecoder(r.Body).Decode(&root)

		if root.Invoice.ClientID != 42 {
			t.Errorf("expected ClientID 42, got %d", root.Invoice.ClientID)
		}
		if root.Invoice.SplitPaymentType != "required" {
			t.Errorf("expected SplitPaymentType %q, got %q", "required", root.Invoice.SplitPaymentType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Invoice{
			ID:       200,
			UUID:     testInvoiceUUID,
			ClientID: 42,
			Number:   "FV/2025/03/001",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	invoice, err := c.Invoices.Create(context.Background(), &Invoice{
		ClientID:         42,
		InvoiceDate:      "2025-03-01",
		SaleDate:         "2025-03-01",
		SplitPaymentType: "required",
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
		if r.URL.Path != "/v3/invoices/"+testInvoiceUUID+".json" {
			t.Errorf("expected UUID path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Invoice{
			ID:    99,
			UUID:  testInvoiceUUID,
			Notes: "Updated notes",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	invoice, err := c.Invoices.Update(context.Background(), testInvoiceUUID, &Invoice{
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
		if r.URL.Path != "/v3/invoices/"+testInvoiceUUID+".json" {
			t.Errorf("expected UUID path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	err := c.Invoices.Delete(context.Background(), testInvoiceUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvoiceService_MarkAsPaid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		expected := "/v3/async/invoices/" + testInvoiceUUID + "/paid.json"
		if r.URL.Path != expected {
			t.Errorf("expected async paid path %q, got %s", expected, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["paid_date"] != "2025-11-16" {
			t.Errorf("expected paid_date %q, got %v", "2025-11-16", body["paid_date"])
		}
		if body["allow_correction"] != true {
			t.Errorf("expected allow_correction=true, got %v", body["allow_correction"])
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	err := c.Invoices.MarkAsPaid(context.Background(), testInvoiceUUID, "2025-11-16", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvoiceService_SendByEmail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		expected := "/v3/invoices/" + testInvoiceUUID + "/deliver_via_email.json"
		if r.URL.Path != expected {
			t.Errorf("expected deliver_via_email path %q, got %s", expected, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["print_type"] != "duplicate" {
			t.Errorf("expected print_type duplicate, got %v", body["print_type"])
		}
		if body["recipient"] != "client@example.com" {
			t.Errorf("expected recipient set, got %v", body["recipient"])
		}
		if body["locale"] != "pl" {
			t.Errorf("expected locale pl, got %v", body["locale"])
		}
		if body["send_copy"] != true {
			t.Errorf("expected send_copy=true, got %v", body["send_copy"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	err := c.Invoices.SendByEmail(context.Background(), testInvoiceUUID, &SendByEmailOptions{
		PrintType: PDFDocumentTypeDuplicate,
		Locale:    "pl",
		Recipient: "client@example.com",
		SendCopy:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvoiceService_SendByEmail_Defaults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["print_type"] != "original" {
			t.Errorf("expected default print_type=original, got %v", body["print_type"])
		}
		if _, ok := body["locale"]; ok {
			t.Errorf("expected locale to be omitted, got %v", body["locale"])
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	if err := c.Invoices.SendByEmail(context.Background(), testInvoiceUUID, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvoiceService_GetPDF(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := "/v3/invoices/" + testInvoiceUUID + "/pdf.json"
		if r.URL.Path != expected {
			t.Errorf("expected pdf.json path %q, got %s", expected, r.URL.Path)
		}
		if r.URL.Query().Get("document_type") != "duplicate" {
			t.Errorf("expected document_type=duplicate, got %q", r.URL.Query().Get("document_type"))
		}
		if r.URL.Query().Get("locale") != "en" {
			t.Errorf("expected locale=en, got %q", r.URL.Query().Get("locale"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PDFResponse{
			DownloadLink: "https://example.test/some/signed.pdf",
			Status:       "ready",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	pdf, err := c.Invoices.GetPDF(context.Background(), testInvoiceUUID, PDFDocumentTypeDuplicate, "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pdf.DownloadLink == "" {
		t.Error("expected DownloadLink populated")
	}
	if pdf.Status != "ready" {
		t.Errorf("expected status=ready, got %q", pdf.Status)
	}
}

func TestInvoiceService_GetPDF_DefaultsDocumentType(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("document_type") != "original" {
			t.Errorf("expected default document_type=original, got %q", r.URL.Query().Get("document_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PDFResponse{DownloadLink: "https://example.test/x.pdf"})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	if _, err := c.Invoices.GetPDF(context.Background(), testInvoiceUUID, "", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInvoiceService_GetNextNumber(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/invoices/next_number.json" {
			t.Errorf("expected path /v3/invoices/next_number.json, got %s", r.URL.Path)
		}

		kind := r.URL.Query().Get("kind")
		if kind != "vat" {
			t.Errorf("expected kind %q, got %q", "vat", kind)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(NextNumberResponse{
			NextNumber: "FV/2025/11/002",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	number, err := c.Invoices.GetNextNumber(context.Background(), "vat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if number != "FV/2025/11/002" {
		t.Errorf("expected number %q, got %q", "FV/2025/11/002", number)
	}
}
