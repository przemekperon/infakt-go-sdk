package infakt

import (
	"context"
	"fmt"
	"net/http"
)

// Product represents a product in the inFakt system.
type Product struct {
	ID           int64   `json:"id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Description  string  `json:"description,omitempty"`
	Unit         string  `json:"unit,omitempty"`
	PKWiU        string  `json:"pkwiu,omitempty"`
	TaxSymbol    string  `json:"tax_symbol,omitempty"`
	UnitNetPrice int     `json:"unit_net_price,omitempty"`
	Quantity     float64 `json:"quantity,omitempty"`
}

// ProductListOptions specifies the optional parameters to the
// ProductService.List method.
type ProductListOptions struct {
	ListOptions
}

// ProductService handles communication with the product related
// methods of the inFakt API.
type ProductService struct {
	client *Client
}

type productRoot struct {
	Product Product `json:"product"`
}

type productListRoot struct {
	MetaInfo MetaInfo  `json:"metainfo"`
	Entities []Product `json:"entities"`
}

// List returns a list of products.
func (s *ProductService) List(ctx context.Context, opts *ProductListOptions) ([]Product, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/products.json", apiVersion)

	if opts != nil {
		var err error
		path, err = addOptions(path, &opts.ListOptions)
		if err != nil {
			return nil, nil, err
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var root productListRoot
	_, err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single product by ID.
func (s *ProductService) Get(ctx context.Context, id int64) (*Product, error) {
	path := fmt.Sprintf("/%s/products/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var product Product
	_, err = s.client.do(req, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// Create creates a new product.
func (s *ProductService) Create(ctx context.Context, product *Product) (*Product, error) {
	path := fmt.Sprintf("/%s/products.json", apiVersion)

	root := &productRoot{Product: *product}
	req, err := s.client.newRequest(ctx, http.MethodPost, path, root)
	if err != nil {
		return nil, err
	}

	var created Product
	_, err = s.client.do(req, &created)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// Update updates an existing product.
func (s *ProductService) Update(ctx context.Context, id int64, product *Product) (*Product, error) {
	path := fmt.Sprintf("/%s/products/%d.json", apiVersion, id)

	root := &productRoot{Product: *product}
	req, err := s.client.newRequest(ctx, http.MethodPut, path, root)
	if err != nil {
		return nil, err
	}

	var updated Product
	_, err = s.client.do(req, &updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// Delete deletes a product.
func (s *ProductService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/%s/products/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}
