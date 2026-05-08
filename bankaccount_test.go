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
		_ = json.NewEncoder(w).Encode(bankAccountListRoot{
			MetaInfo: MetaInfo{Count: 2, TotalCount: 2},
			Entities: []BankAccount{
				{ID: 1, BankName: "mBank", AccountNumber: "PL12345678901234567890123456", Default: true},
				{ID: 2, BankName: "ING", AccountNumber: "PL65432109876543210987654321"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
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
	if !accounts[0].Default {
		t.Error("expected first account to be default")
	}
}

func TestBankAccountService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/bank_accounts/7.json" {
			t.Errorf("expected path /v3/bank_accounts/7.json, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BankAccount{ID: 7, BankName: "Santander"})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	acc, err := c.BankAccounts.Get(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.BankName != "Santander" {
		t.Errorf("expected BankName %q, got %q", "Santander", acc.BankName)
	}
}

func TestBankAccountService_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var root bankAccountRoot
		_ = json.NewDecoder(r.Body).Decode(&root)
		if root.BankAccount.AccountNumber != "PL11111111111111111111111111" {
			t.Errorf("expected wrapped account number, got %q", root.BankAccount.AccountNumber)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(BankAccount{ID: 99, AccountNumber: root.BankAccount.AccountNumber})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	acc, err := c.BankAccounts.Create(context.Background(), &BankAccount{
		BankName:      "mBank",
		AccountNumber: "PL11111111111111111111111111",
		Currency:      "PLN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.ID != 99 {
		t.Errorf("expected ID 99, got %d", acc.ID)
	}
}

func TestBankAccountService_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v3/bank_accounts/7.json" {
			t.Errorf("expected path /v3/bank_accounts/7.json, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(BankAccount{ID: 7, CustomName: "Operacyjne"})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	acc, err := c.BankAccounts.Update(context.Background(), 7, &BankAccount{CustomName: "Operacyjne"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.CustomName != "Operacyjne" {
		t.Errorf("expected CustomName %q, got %q", "Operacyjne", acc.CustomName)
	}
}

func TestBankAccountService_Delete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v3/bank_accounts/7.json" {
			t.Errorf("expected path /v3/bank_accounts/7.json, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	if err := c.BankAccounts.Delete(context.Background(), 7); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBankAccountService_ListWithPagination(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limit := r.URL.Query().Get("limit")
		if limit != "1" {
			t.Errorf("expected limit 1, got %s", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(bankAccountListRoot{
			MetaInfo: MetaInfo{Count: 1, TotalCount: 5},
			Entities: []BankAccount{
				{ID: 1, BankName: "PKO"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
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
