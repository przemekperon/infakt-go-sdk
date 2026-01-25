package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientEntityService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v3/clients.json" {
			t.Errorf("expected path /v3/clients.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clientEntityListRoot{
			MetaInfo: MetaInfo{Count: 2, TotalCount: 2},
			Entities: []ClientEntity{
				{ID: 1, CompanyName: "Firma A"},
				{ID: 2, CompanyName: "Firma B"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	entities, meta, err := c.Clients.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(entities))
	}
	if meta.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", meta.TotalCount)
	}
	if entities[0].CompanyName != "Firma A" {
		t.Errorf("expected CompanyName %q, got %q", "Firma A", entities[0].CompanyName)
	}
}

func TestClientEntityService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/clients/42.json" {
			t.Errorf("expected path /v3/clients/42.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ClientEntity{
			ID:          42,
			CompanyName: "Test Corp",
			NIP:         "1234567890",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	entity, err := c.Clients.Get(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.ID != 42 {
		t.Errorf("expected ID 42, got %d", entity.ID)
	}
	if entity.CompanyName != "Test Corp" {
		t.Errorf("expected CompanyName %q, got %q", "Test Corp", entity.CompanyName)
	}
}

func TestClientEntityService_Create(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var root clientEntityRoot
		_ = json.NewDecoder(r.Body).Decode(&root)

		if root.ClientEntity.CompanyName != "New Corp" {
			t.Errorf("expected CompanyName %q, got %q", "New Corp", root.ClientEntity.CompanyName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(ClientEntity{
			ID:          100,
			CompanyName: "New Corp",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	entity, err := c.Clients.Create(context.Background(), &ClientEntity{
		CompanyName: "New Corp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.ID != 100 {
		t.Errorf("expected ID 100, got %d", entity.ID)
	}
}

func TestClientEntityService_Update(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/v3/clients/42.json" {
			t.Errorf("expected path /v3/clients/42.json, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ClientEntity{
			ID:          42,
			CompanyName: "Updated Corp",
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	entity, err := c.Clients.Update(context.Background(), 42, &ClientEntity{
		CompanyName: "Updated Corp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entity.CompanyName != "Updated Corp" {
		t.Errorf("expected CompanyName %q, got %q", "Updated Corp", entity.CompanyName)
	}
}

func TestClientEntityService_Delete(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v3/clients/42.json" {
			t.Errorf("expected path /v3/clients/42.json, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	err := c.Clients.Delete(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientEntityService_ListWithPagination(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset := r.URL.Query().Get("offset")
		limit := r.URL.Query().Get("limit")

		if offset != "10" {
			t.Errorf("expected offset 10, got %s", offset)
		}
		if limit != "5" {
			t.Errorf("expected limit 5, got %s", limit)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(clientEntityListRoot{
			MetaInfo: MetaInfo{Count: 5, TotalCount: 20},
			Entities: []ClientEntity{
				{ID: 11, CompanyName: "Firma K"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	entities, meta, err := c.Clients.List(context.Background(), &ClientEntityListOptions{
		ListOptions: ListOptions{Offset: 10, Limit: 5},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(entities) != 1 {
		t.Errorf("expected 1 entity, got %d", len(entities))
	}
	if meta.TotalCount != 20 {
		t.Errorf("expected TotalCount 20, got %d", meta.TotalCount)
	}
}
