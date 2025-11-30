package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBankAccountService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v3/bank_accounts.json" {
			t.Errorf("expected path /v3/bank_accounts.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bankAccountListRoot{
			MetaInfo: MetaInfo{Count: 2, TotalCount: 2},
			Entities: []BankAccount{
				{ID: 1, BankName: "mBank", AccountNumber: "PL12345678901234567890123456", IsDefault: true},
				{ID: 2, BankName: "ING", AccountNumber: "PL65432109876543210987654321"},
			},
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	accounts, meta, err := c.BankAccounts.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
	if meta.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", meta.TotalCount)
	}
	if accounts[0].BankName != "mBank" {
		t.Errorf("expected BankName %q, got %q", "mBank", accounts[0].BankName)
	}
	if !accounts[0].IsDefault {
		t.Error("expected first account to be default")
	}
}

func TestBankAccountService_ListWithPagination(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "1" {
			t.Errorf("expected limit 1, got %s", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bankAccountListRoot{
			MetaInfo: MetaInfo{Count: 1, TotalCount: 5},
			Entities: []BankAccount{
				{ID: 1, BankName: "PKO"},
			},
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	accounts, meta, err := c.BankAccounts.List(context.Background(), &ListOptions{Limit: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(accounts))
	}
	if meta.TotalCount != 5 {
		t.Errorf("expected TotalCount 5, got %d", meta.TotalCount)
	}
}
