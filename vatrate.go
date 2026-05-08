package infakt

import (
	"context"
	"fmt"
	"net/http"
)

// VatRate represents a VAT rate in the inFakt system.
type VatRate struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Rate is the percentage value as an integer string
	// (e.g., "23" for 23%).
	Rate string `json:"rate,omitempty"`
	// Symbol is the short symbol used in [Invoice] and [ServiceEntry]
	// (e.g., "23", "8", "zw", "np").
	Symbol string `json:"symbol,omitempty"`
	// ValidFrom is the first day on which this rate applies, formatted as
	// YYYY-MM-DD.
	ValidFrom string `json:"valid_from,omitempty"`
	// ValidUntil is the last day on which this rate applies, formatted as
	// YYYY-MM-DD. It is empty for currently-valid rates.
	ValidUntil string `json:"valid_until,omitempty"`
}

// VatRateService manages VAT rates on the inFakt API. VAT rates are
// read-only.
//
// Supported endpoints:
//   - List
//
// Access it through [Client.VatRates].
//
// See https://docs.infakt.pl for the corresponding API reference.
type VatRateService struct {
	client *Client
}

type vatRateListRoot struct {
	MetaInfo MetaInfo  `json:"metainfo"`
	Entities []VatRate `json:"entities"`
}

// List returns VAT rates, paginated via [ListOptions]. The returned
// [MetaInfo] reports the total count and pagination cursors.
func (s *VatRateService) List(ctx context.Context, opts *ListOptions) ([]VatRate, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/vat_rates.json", apiVersion)

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

	var root vatRateListRoot
	err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}
