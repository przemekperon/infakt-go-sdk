package infakt

import (
	"net/url"
	"testing"
)

func TestAddOptions_ClampsLimit(t *testing.T) {
	got, err := addOptions("/v3/x.json", &ListOptions{Limit: 250})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if u.Query().Get("limit") != "100" {
		t.Errorf("expected limit clamped to 100, got %q", u.Query().Get("limit"))
	}
}

func TestAddOptions_OrderFieldsFilters(t *testing.T) {
	got, err := addOptions("/v3/x.json", &ListOptions{
		Order:   "name asc",
		Fields:  []string{"id", "name", "services(name,tax_symbol)"},
		Filters: map[string]string{"number_eq": "1/2026"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	q := u.Query()
	if q.Get("order") != "name asc" {
		t.Errorf("expected order, got %q", q.Get("order"))
	}
	if q.Get("fields") != "id,name,services(name,tax_symbol)" {
		t.Errorf("expected fields joined, got %q", q.Get("fields"))
	}
	if q.Get("q[number_eq]") != "1/2026" {
		t.Errorf("expected q[number_eq] filter, got %q", q.Get("q[number_eq]"))
	}
}

func TestAddOptions_Nil(t *testing.T) {
	got, err := addOptions("/v3/x.json", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/v3/x.json" {
		t.Errorf("expected unchanged path, got %q", got)
	}
}
