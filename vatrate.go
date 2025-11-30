package infakt

import (
	"context"
	"fmt"
	"net/http"
)

// VatRate represents a VAT rate in the inFakt system.
type VatRate struct {
	ID     int64  `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
	Rate   int    `json:"rate,omitempty"`
	Symbol string `json:"symbol,omitempty"`
}

// VatRateService handles communication with the VAT rate related
// methods of the inFakt API. VAT rates are read-only.
type VatRateService struct {
	client *Client
}

type vatRateListRoot struct {
	MetaInfo MetaInfo  `json:"metainfo"`
	Entities []VatRate `json:"entities"`
}

// List returns a list of VAT rates.
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
	_, err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}
