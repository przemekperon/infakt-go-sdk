package infakt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGtuService_List(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/gtus.json" {
			t.Errorf("expected path /v3/gtus.json, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gtuListRoot{
			MetaInfo: MetaInfo{Count: 2, TotalCount: 2},
			Entities: []Gtu{
				{ID: 1, Name: "GTU_01", ShortDescription: "Napoje alkoholowe"},
				{ID: 7, Name: "GTU_07", ShortDescription: "Pojazdy"},
			},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	gtus, meta, err := c.Gtus.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gtus) != 2 {
		t.Errorf("expected 2 gtus, got %d", len(gtus))
	}
	if meta.TotalCount != 2 {
		t.Errorf("expected TotalCount 2, got %d", meta.TotalCount)
	}
	if gtus[0].Name != "GTU_01" {
		t.Errorf("expected Name %q, got %q", "GTU_01", gtus[0].Name)
	}
}

func TestGtuService_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/gtus/3.json" {
			t.Errorf("expected path /v3/gtus/3.json, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Gtu{ID: 3, Name: "GTU_03"})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	g, err := c.Gtus.Get(context.Background(), 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Name != "GTU_03" {
		t.Errorf("expected Name %q, got %q", "GTU_03", g.Name)
	}
}

func TestGtuService_ListSelected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/gtus/selected.json" {
			t.Errorf("expected path /v3/gtus/selected.json, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(gtuListRoot{
			Entities: []Gtu{{ID: 1, Name: "GTU_01"}},
		})
	}))
	defer ts.Close()

	c := newTestClient("key", WithBaseURL(ts.URL))
	gtus, err := c.Gtus.ListSelected(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gtus) != 1 {
		t.Errorf("expected 1 selected gtu, got %d", len(gtus))
	}
}
