package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProductService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v3/products.json" {
			t.Errorf("expected path /v3/products.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(productListRoot{
			MetaInfo: MetaInfo{Count: 2, TotalCount: 2},
			Entities: []Product{
				{ID: 1, Name: "Hosting"},
				{ID: 2, Name: "Domain"},
			},
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	products, meta, err := c.Products.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}
	if meta.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", meta.TotalCount)
	}
}

func TestProductService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/products/10.json" {
			t.Errorf("expected path /v3/products/10.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Product{
			ID:   10,
			Name: "Consulting",
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	product, err := c.Products.Get(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if product.Name != "Consulting" {
		t.Errorf("expected Name %q, got %q", "Consulting", product.Name)
	}
}

func TestProductService_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Product{
			ID:   50,
			Name: "New Product",
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	product, err := c.Products.Create(context.Background(), &Product{
		Name:         "New Product",
		UnitNetPrice: 10000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if product.ID != 50 {
		t.Errorf("expected ID 50, got %d", product.ID)
	}
}

func TestProductService_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Product{
			ID:   10,
			Name: "Updated Product",
		})
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	product, err := c.Products.Update(context.Background(), 10, &Product{
		Name: "Updated Product",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if product.Name != "Updated Product" {
		t.Errorf("expected Name %q, got %q", "Updated Product", product.Name)
	}
}

func TestProductService_Delete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := NewClient("key", WithBaseURL(ts.URL))
	err := c.Products.Delete(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
