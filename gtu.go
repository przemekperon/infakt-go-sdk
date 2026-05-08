package infakt

import (
	"context"
	"fmt"
	"net/http"
)

// Gtu represents a GTU code — a "Group of Goods/Services" marking required
// on Polish JPK_V7 invoices for selected commodity classes (e.g. fuel,
// alcohol, electronics). See https://docs.infakt.pl ("Kody GTU") for the
// canonical catalog.
type Gtu struct {
	ID int64 `json:"id,omitempty"`
	// Name is the JPK_V7 marking — the live API stores the code itself
	// in this field (e.g. "GTU_01", "GTU_07").
	Name string `json:"name,omitempty"`
	// ShortDescription is the one-line label (e.g. "Dostawa napojów
	// alkoholowych").
	ShortDescription string `json:"short_description,omitempty"`
	// Description is the long-form regulatory text.
	Description string `json:"description,omitempty"`
}

// GtuService provides read-only access to the GTU reference catalog.
//
// Supported endpoints:
//   - List, Get, ListSelected
//
// Access it through [Client.Gtus].
//
// See https://docs.infakt.pl for the corresponding API reference.
type GtuService struct {
	client *Client
}

type gtuListRoot struct {
	MetaInfo MetaInfo `json:"metainfo"`
	Entities []Gtu    `json:"entities"`
}

// List returns the full GTU catalog, paginated via [ListOptions].
func (s *GtuService) List(ctx context.Context, opts *ListOptions) ([]Gtu, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/gtus.json", apiVersion)

	if opts != nil {
		var err error
		path, err = addOptions(path, opts)
		if err != nil {
			return nil, nil, err
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var root gtuListRoot
	err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single [Gtu] entry by ID.
func (s *GtuService) Get(ctx context.Context, id int64) (*Gtu, error) {
	path := fmt.Sprintf("/%s/gtus/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var gtu Gtu
	err = s.client.do(req, &gtu)
	if err != nil {
		return nil, err
	}

	return &gtu, nil
}

// ListSelected returns the GTU codes the current account has opted into.
// The endpoint does not paginate — the full active selection is returned
// in a single call.
func (s *GtuService) ListSelected(ctx context.Context) ([]Gtu, error) {
	path := fmt.Sprintf("/%s/gtus/selected.json", apiVersion)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var root gtuListRoot
	if err := s.client.do(req, &root); err != nil {
		return nil, err
	}

	return root.Entities, nil
}
