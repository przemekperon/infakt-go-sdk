package infakt

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Invoice represents an invoice in the inFakt system.
type Invoice struct {
	ID                        int64          `json:"id,omitempty"`
	UUID                      string         `json:"uuid,omitempty"`
	ParentID                  *int64         `json:"parent_id,omitempty"`
	Number                    string         `json:"number,omitempty"`
	Currency                  string         `json:"currency,omitempty"`
	PaidPrice                 int            `json:"paid_price,omitempty"`
	Notes                     string         `json:"notes,omitempty"`
	Kind                      string         `json:"kind,omitempty"`
	PaymentMethod             string         `json:"payment_method,omitempty"`
	SplitPayment              bool           `json:"split_payment,omitempty"`
	RecipientSignature        string         `json:"recipient_signature,omitempty"`
	SellerSignature           string         `json:"seller_signature,omitempty"`
	InvoiceDate               string         `json:"invoice_date,omitempty"`
	SaleDate                  string         `json:"sale_date,omitempty"`
	PaymentDate               string         `json:"payment_date,omitempty"`
	Status                    string         `json:"status,omitempty"`
	PaidDate                  string         `json:"paid_date,omitempty"`
	NetPrice                  int            `json:"net_price,omitempty"`
	TaxPrice                  int            `json:"tax_price,omitempty"`
	GrossPrice                int            `json:"gross_price,omitempty"`
	LeftToPay                 int            `json:"left_to_pay,omitempty"`
	ClientID                  int64          `json:"client_id,omitempty"`
	ClientUUID                string         `json:"client_uuid,omitempty"`
	ClientCompanyName         string         `json:"client_company_name,omitempty"`
	ClientFirstName           string         `json:"client_first_name,omitempty"`
	ClientLastName            string         `json:"client_last_name,omitempty"`
	ClientBusinessActivityKind string        `json:"client_business_activity_kind,omitempty"`
	ClientStreet              string         `json:"client_street,omitempty"`
	ClientStreetNumber        string         `json:"client_street_number,omitempty"`
	ClientFlatNumber          string         `json:"client_flat_number,omitempty"`
	ClientCity                string         `json:"client_city,omitempty"`
	ClientPostCode            string         `json:"client_post_code,omitempty"`
	ClientTaxCode             string         `json:"client_tax_code,omitempty"`
	ClientCountry             string         `json:"client_country,omitempty"`
	BankName                  string         `json:"bank_name,omitempty"`
	BankAccount               string         `json:"bank_account,omitempty"`
	Swift                     string         `json:"swift,omitempty"`
	SaleType                  string         `json:"sale_type,omitempty"`
	InvoiceDateKind           string         `json:"invoice_date_kind,omitempty"`
	VatExemptionReason        string         `json:"vat_exemption_reason,omitempty"`
	SalesKind                 string         `json:"sales_kind,omitempty"`
	AmountInWords             string         `json:"amount_in_words,omitempty"`
	CreatedAt                 string         `json:"created_at,omitempty"`
	Services                  []ServiceEntry `json:"services,omitempty"`
	Extensions                *Extensions    `json:"extensions,omitempty"`
}

// ServiceEntry represents a line item on an invoice.
type ServiceEntry struct {
	ID                         int64   `json:"id,omitempty"`
	Name                       string  `json:"name,omitempty"`
	TaxSymbol                  string  `json:"tax_symbol,omitempty"`
	Unit                       string  `json:"unit,omitempty"`
	Quantity                   float64 `json:"quantity,omitempty"`
	UnitNetPrice               int     `json:"unit_net_price,omitempty"`
	UnitNetPriceBeforeDiscount int     `json:"unit_net_price_before_discount,omitempty"`
	NetPrice                   int     `json:"net_price,omitempty"`
	GrossPrice                 int     `json:"gross_price,omitempty"`
	TaxPrice                   int     `json:"tax_price,omitempty"`
	Symbol                     string  `json:"symbol,omitempty"`
	PKWiU                      string  `json:"pkwiu,omitempty"`
	Discount                   string  `json:"discount,omitempty"`
}

// Extensions represents additional invoice settings.
type Extensions struct {
	PaymentOnline *PaymentOnline `json:"payment_online,omitempty"`
}

// PaymentOnline represents online payment configuration.
type PaymentOnline struct {
	Enabled bool `json:"enabled,omitempty"`
}

// InvoiceRequest is used for creating and updating invoices.
type InvoiceRequest struct {
	Currency           *string               `json:"currency,omitempty"`
	Notes              *string               `json:"notes,omitempty"`
	Kind               *string               `json:"kind,omitempty"`
	PaymentMethod      *string               `json:"payment_method,omitempty"`
	RecipientSignature *string               `json:"recipient_signature,omitempty"`
	SellerSignature    *string               `json:"seller_signature,omitempty"`
	InvoiceDate        *string               `json:"invoice_date,omitempty"`
	SaleDate           *string               `json:"sale_date,omitempty"`
	PaymentDate        *string               `json:"payment_date,omitempty"`
	ClientID           *int64                `json:"client_id,omitempty"`
	BankAccount        *string               `json:"bank_account,omitempty"`
	SaleType           *string               `json:"sale_type,omitempty"`
	InvoiceDateKind    *string               `json:"invoice_date_kind,omitempty"`
	Services           []ServiceEntryRequest `json:"services,omitempty"`
}

// ServiceEntryRequest is used for creating and updating service line items.
type ServiceEntryRequest struct {
	Name         *string  `json:"name,omitempty"`
	TaxSymbol    *string  `json:"tax_symbol,omitempty"`
	Unit         *string  `json:"unit,omitempty"`
	Quantity     *float64 `json:"quantity,omitempty"`
	UnitNetPrice *int     `json:"unit_net_price,omitempty"`
	Symbol       *string  `json:"symbol,omitempty"`
	PKWiU        *string  `json:"pkwiu,omitempty"`
	Discount     *string  `json:"discount,omitempty"`
}

// InvoiceListOptions specifies the optional parameters to the
// InvoiceService.List method.
type InvoiceListOptions struct {
	ListOptions

	// Filter fields using inFakt query format (q[field_predicate]=value)
	DateFrom string // q[invoice_date_gteq]
	DateTo   string // q[invoice_date_lteq]
	ClientID int64  // q[client_id_eq]
	Status   string // q[status_eq]
}

// InvoiceService handles communication with the invoice related
// methods of the inFakt API.
type InvoiceService struct {
	client *Client
}

type invoiceRoot struct {
	Invoice Invoice `json:"invoice"`
}

type invoiceListRoot struct {
	MetaInfo MetaInfo   `json:"metainfo"`
	Entities []Invoice  `json:"entities"`
}

// List returns a list of invoices.
func (s *InvoiceService) List(ctx context.Context, opts *InvoiceListOptions) ([]Invoice, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/invoices.json", apiVersion)

	if opts != nil {
		var err error
		path, err = addOptions(path, &opts.ListOptions)
		if err != nil {
			return nil, nil, err
		}
		path = addInvoiceFilters(path, opts)
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var root invoiceListRoot
	_, err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single invoice by ID.
func (s *InvoiceService) Get(ctx context.Context, id int64) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	_, err = s.client.do(req, &invoice)
	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

// Create creates a new invoice.
func (s *InvoiceService) Create(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices.json", apiVersion)

	root := &invoiceRoot{Invoice: *invoice}
	req, err := s.client.newRequest(ctx, http.MethodPost, path, root)
	if err != nil {
		return nil, err
	}

	var created Invoice
	_, err = s.client.do(req, &created)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// Update updates an existing invoice.
func (s *InvoiceService) Update(ctx context.Context, id int64, invoice *Invoice) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices/%d.json", apiVersion, id)

	root := &invoiceRoot{Invoice: *invoice}
	req, err := s.client.newRequest(ctx, http.MethodPut, path, root)
	if err != nil {
		return nil, err
	}

	var updated Invoice
	_, err = s.client.do(req, &updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// Delete deletes an invoice.
func (s *InvoiceService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/%s/invoices/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}

// MarkAsPaid marks an invoice as paid.
func (s *InvoiceService) MarkAsPaid(ctx context.Context, id int64, paidDate string) error {
	path := fmt.Sprintf("/%s/invoices/%d/paid.json", apiVersion, id)

	body := map[string]string{}
	if paidDate != "" {
		body["paid_date"] = paidDate
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}

// SendByEmail sends an invoice by email.
func (s *InvoiceService) SendByEmail(ctx context.Context, id int64, emailTo string) error {
	path := fmt.Sprintf("/%s/invoices/%d/deliver_via_email.json", apiVersion, id)

	body := map[string]interface{}{
		"print_type": "original",
	}
	if emailTo != "" {
		body["email_to"] = emailTo
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	_, err = s.client.do(req, nil)
	return err
}

// GetPDF returns the PDF content of an invoice.
func (s *InvoiceService) GetPDF(ctx context.Context, id int64) ([]byte, error) {
	path := fmt.Sprintf("/%s/invoices/%d.pdf", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("infakt: request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("infakt: failed to read PDF response: %w", err)
	}

	return data, nil
}

// NextNumber returns the next available invoice number for the given kind.
type NextNumberResponse struct {
	NextNumber string `json:"next_number"`
}

// GetNextNumber returns the next available invoice number.
func (s *InvoiceService) GetNextNumber(ctx context.Context, kind string) (string, error) {
	path := fmt.Sprintf("/%s/invoices/next_number.json", apiVersion)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	if kind != "" {
		q := req.URL.Query()
		q.Set("kind", kind)
		req.URL.RawQuery = q.Encode()
	}

	var result NextNumberResponse
	_, err = s.client.do(req, &result)
	if err != nil {
		return "", err
	}

	return result.NextNumber, nil
}

func addInvoiceFilters(path string, opts *InvoiceListOptions) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}

	q := u.Query()
	if opts.DateFrom != "" {
		q.Set("q[invoice_date_gteq]", opts.DateFrom)
	}
	if opts.DateTo != "" {
		q.Set("q[invoice_date_lteq]", opts.DateTo)
	}
	if opts.ClientID != 0 {
		q.Set("q[client_id_eq]", fmt.Sprintf("%d", opts.ClientID))
	}
	if opts.Status != "" {
		q.Set("q[status_eq]", opts.Status)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
