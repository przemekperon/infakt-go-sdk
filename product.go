package infakt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Product represents a product in the inFakt system.
//
// Monetary fields ([Product.UnitNetPrice], [Product.NetPrice],
// [Product.TaxPrice], [Product.GrossPrice], [Product.PurchaseUnitNetPrice],
// [Product.PurchaseUnitGrossPrice]) are expressed in grosze (1/100 PLN);
// for example 12345 represents 123.45 PLN.
type Product struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// Symbol is a server-derived short identifier composed from PKWiU /
	// CN / PKOB. It is read-only on the API side.
	Symbol string `json:"symbol,omitempty"`
	// Unit is the unit of measure for the product. See [ServiceEntry.Unit].
	Unit string `json:"unit,omitempty"`
	// PKWiU is the Polish Classification of Products and Services code.
	// See [ServiceEntry.PKWiU].
	PKWiU string `json:"pkwiu,omitempty"`
	// CN is the Combined Nomenclature code (Nomenklatura Scalona) used
	// for goods crossing EU borders.
	CN string `json:"cn,omitempty"`
	// PKOB is the Polish Classification of Construction Objects code.
	PKOB string `json:"pkob,omitempty"`
	// GtuID references a GTU (Group of Goods/Services) marking; see
	// [GtuService] for the catalog.
	GtuID int64 `json:"gtu_id,omitempty"`
	// TaxSymbol is the VAT rate symbol applied to this product.
	// See [ServiceEntry.TaxSymbol].
	TaxSymbol string `json:"tax_symbol,omitempty"`
	// FlatRateTaxSymbol is the VAT rate symbol used when the seller is on
	// the Polish flat-rate tax scheme (ryczałt).
	FlatRateTaxSymbol string `json:"flat_rate_tax_symbol,omitempty"`
	// UnitNetPrice is the per-unit net price, in grosze (1/100 PLN).
	UnitNetPrice int `json:"unit_net_price,omitempty"`
	// UnitNetPriceBeforeDiscount is the per-unit net price prior to any
	// discount, in grosze (1/100 PLN).
	UnitNetPriceBeforeDiscount int `json:"unit_net_price_before_discount,omitempty"`
	// PurchaseUnitNetPrice is the per-unit net cost of acquisition, in
	// grosze (1/100 PLN).
	PurchaseUnitNetPrice int `json:"purchase_unit_net_price,omitempty"`
	// PurchaseUnitGrossPrice is the per-unit gross cost of acquisition,
	// in grosze (1/100 PLN).
	PurchaseUnitGrossPrice int `json:"purchase_unit_gross_price,omitempty"`
	// NetPrice is the total net price, in grosze (1/100 PLN).
	NetPrice int `json:"net_price,omitempty"`
	// TaxPrice is the VAT amount, in grosze (1/100 PLN).
	TaxPrice int `json:"tax_price,omitempty"`
	// GrossPrice is the gross price (net + VAT), in grosze (1/100 PLN).
	GrossPrice int `json:"gross_price,omitempty"`
	// Quantity is the product quantity. See [ServiceEntry.Quantity].
	Quantity float64 `json:"quantity,omitempty"`
	// Discount is a discount factor. See [ServiceEntry.Discount].
	Discount string `json:"discount,omitempty"`
}

// ProductRequest is the partial-update payload for creating and updating
// products. Pointer fields let callers distinguish unset from zero-value:
// leaving a field as nil means "do not change", while passing
// [Int](0) sets it to zero. Use the helpers [String], [Int], and
// [Float64] from helpers.go to build values.
type ProductRequest struct {
	Name                   *string  `json:"name,omitempty"`
	Unit                   *string  `json:"unit,omitempty"`
	PKWiU                  *string  `json:"pkwiu,omitempty"`
	CN                     *string  `json:"cn,omitempty"`
	PKOB                   *string  `json:"pkob,omitempty"`
	GtuID                  *int64   `json:"gtu_id,omitempty"`
	TaxSymbol              *string  `json:"tax_symbol,omitempty"`
	FlatRateTaxSymbol      *string  `json:"flat_rate_tax_symbol,omitempty"`
	UnitNetPrice           *int     `json:"unit_net_price,omitempty"`
	PurchaseUnitNetPrice   *int     `json:"purchase_unit_net_price,omitempty"`
	PurchaseUnitGrossPrice *int     `json:"purchase_unit_gross_price,omitempty"`
	Quantity               *float64 `json:"quantity,omitempty"`
	Discount               *string  `json:"discount,omitempty"`
}

// ProductListOptions specifies the optional parameters to the
// [ProductService.List] method.
//
// Filters are translated into the inFakt query syntax q[field_op]=value.
// Common operators include cont (substring match), eq (exact match),
// gteq (>=), and lteq (<=). The Go fields below map to specific operators
// internally; see the comments next to each field. Use
// [ListOptions.Filters] to express predicates not exposed here.
type ProductListOptions struct {
	ListOptions

	// Filter fields using inFakt query format (q[field_predicate]=value)
	Name   string // q[name_cont]
	NameEq string // q[name_eq]
}

// ProductService manages products on the inFakt API.
//
// Supported endpoints:
//   - List, Get, Create, Update, Delete
//
// Access it through [Client.Products].
//
// See https://docs.infakt.pl for the corresponding API reference.
type ProductService struct {
	client *Client
}

type productWriteRoot struct {
	Product *ProductRequest `json:"product"`
}

type productListRoot struct {
	MetaInfo MetaInfo  `json:"metainfo"`
	Entities []Product `json:"entities"`
}

// List returns products, paginated and filtered via [ProductListOptions].
// The returned [MetaInfo] reports the total count and pagination cursors.
func (s *ProductService) List(ctx context.Context, opts *ProductListOptions) ([]Product, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/products.json", apiVersion)

	if opts != nil {
		var err error
		path, err = addOptions(path, &opts.ListOptions)
		if err != nil {
			return nil, nil, err
		}
		path = addProductFilters(path, opts)
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var root productListRoot
	err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single [Product] by ID. Returns [ErrNotFound] (wrapped in
// an [*ErrorResponse]) if no product exists with that ID.
func (s *ProductService) Get(ctx context.Context, id int64) (*Product, error) {
	path := fmt.Sprintf("/%s/products/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var product Product
	err = s.client.do(req, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// Create creates a new [Product] from the supplied [ProductRequest] and
// returns the server-assigned record (with ID and computed totals).
func (s *ProductService) Create(ctx context.Context, req *ProductRequest) (*Product, error) {
	path := fmt.Sprintf("/%s/products.json", apiVersion)

	httpReq, err := s.client.newRequest(ctx, http.MethodPost, path, &productWriteRoot{Product: req})
	if err != nil {
		return nil, err
	}

	var created Product
	if err := s.client.do(httpReq, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// Update updates an existing [Product] using a partial [ProductRequest].
// Returns [ErrNotFound] (wrapped in an [*ErrorResponse]) if no product
// exists with that ID.
func (s *ProductService) Update(ctx context.Context, id int64, req *ProductRequest) (*Product, error) {
	path := fmt.Sprintf("/%s/products/%d.json", apiVersion, id)

	httpReq, err := s.client.newRequest(ctx, http.MethodPut, path, &productWriteRoot{Product: req})
	if err != nil {
		return nil, err
	}

	var updated Product
	if err := s.client.do(httpReq, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

// Delete deletes a [Product] by ID. Returns [ErrNotFound] (wrapped in an
// [*ErrorResponse]) if no product exists with that ID.
func (s *ProductService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/%s/products/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}

func addProductFilters(path string, opts *ProductListOptions) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}

	q := u.Query()
	if opts.Name != "" {
		q.Set("q[name_cont]", opts.Name)
	}
	if opts.NameEq != "" {
		q.Set("q[name_eq]", opts.NameEq)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
