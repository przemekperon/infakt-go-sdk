package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVatRateService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v3/vat_rates.json" {
			t.Errorf("expected path /v3/vat_rates.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(vatRateListRoot{
			MetaInfo: MetaInfo{Count: 3, TotalCount: 3},
			Entities: []VatRate{
				{ID: 1, Name: "23%", Rate: "23.0", Symbol: "23"},
				{ID: 2, Name: "8%", Rate: "8.0", Symbol: "8"},
				{ID: 3, Name: "zw", Rate: "0.0", Symbol: "zw"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	rates, meta, err := c.VatRates.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rates) != 3 {
		t.Errorf("expected 3 rates, got %d", len(rates))
	}
	if meta.TotalCount != 3 {
		t.Errorf("expected TotalCount 3, got %d", meta.TotalCount)
	}
	if rates[0].Symbol != "23" {
		t.Errorf("expected Symbol %q, got %q", "23", rates[0].Symbol)
	}
	if rates[2].Symbol != "zw" {
		t.Errorf("expected Symbol %q, got %q", "zw", rates[2].Symbol)
	}
}
