package infakt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Invoice represents an invoice in the inFakt system.
//
// Monetary fields ([Invoice.NetPrice], [Invoice.TaxPrice], [Invoice.GrossPrice],
// [Invoice.PaidPrice], [Invoice.LeftToPay]) are expressed in grosze (1/100 PLN);
// for example 12345 represents 123.45 PLN. Date fields are formatted as
// YYYY-MM-DD strings.
type Invoice struct {
	ID       int64  `json:"id,omitempty"`
	UUID     string `json:"uuid,omitempty"`
	ParentID *int64 `json:"parent_id,omitempty"`
	Number   string `json:"number,omitempty"`
	// Currency is the ISO 4217 currency code (e.g., "PLN", "EUR").
	Currency string `json:"currency,omitempty"`
	// PaidPrice is the amount already paid, in grosze (1/100 PLN).
	PaidPrice int    `json:"paid_price,omitempty"`
	Notes     string `json:"notes,omitempty"`
	// Kind selects the invoice type. Typical values include "vat",
	// "proforma", "advance", "final", "correction". See
	// https://docs.infakt.pl for the complete list.
	Kind string `json:"kind,omitempty"`
	// PaymentMethod identifies how the invoice should be paid
	// (e.g., "transfer", "cash", "card"). See https://docs.infakt.pl.
	PaymentMethod string `json:"payment_method,omitempty"`
	// SplitPayment, when true, enables the Polish split-payment mechanism
	// (mechanizm podzielonej płatności).
	SplitPayment       bool   `json:"split_payment,omitempty"`
	RecipientSignature string `json:"recipient_signature,omitempty"`
	SellerSignature    string `json:"seller_signature,omitempty"`
	// InvoiceDate is the issue date, formatted as YYYY-MM-DD.
	InvoiceDate string `json:"invoice_date,omitempty"`
	// SaleDate is the date of sale, formatted as YYYY-MM-DD.
	SaleDate string `json:"sale_date,omitempty"`
	// PaymentDate is the due date for payment, formatted as YYYY-MM-DD.
	PaymentDate string `json:"payment_date,omitempty"`
	// Status is the current invoice state (e.g., "draft", "sent", "paid").
	// See https://docs.infakt.pl for the full set of values.
	Status string `json:"status,omitempty"`
	// PaidDate is when the invoice was marked as paid, formatted as YYYY-MM-DD.
	PaidDate string `json:"paid_date,omitempty"`
	// NetPrice is the total net amount, in grosze (1/100 PLN).
	NetPrice int `json:"net_price,omitempty"`
	// TaxPrice is the total VAT amount, in grosze (1/100 PLN).
	TaxPrice int `json:"tax_price,omitempty"`
	// GrossPrice is the total gross amount (net + VAT), in grosze (1/100 PLN).
	GrossPrice int `json:"gross_price,omitempty"`
	// LeftToPay is the outstanding balance, in grosze (1/100 PLN).
	LeftToPay                  int    `json:"left_to_pay,omitempty"`
	ClientID                   int64  `json:"client_id,omitempty"`
	ClientUUID                 string `json:"client_uuid,omitempty"`
	ClientCompanyName          string `json:"client_company_name,omitempty"`
	ClientFirstName            string `json:"client_first_name,omitempty"`
	ClientLastName             string `json:"client_last_name,omitempty"`
	ClientBusinessActivityKind string `json:"client_business_activity_kind,omitempty"`
	ClientStreet               string `json:"client_street,omitempty"`
	ClientStreetNumber         string `json:"client_street_number,omitempty"`
	ClientFlatNumber           string `json:"client_flat_number,omitempty"`
	ClientCity                 string `json:"client_city,omitempty"`
	ClientPostCode             string `json:"client_post_code,omitempty"`
	ClientTaxCode              string `json:"client_tax_code,omitempty"`
	ClientCountry              string `json:"client_country,omitempty"`
	BankName                   string `json:"bank_name,omitempty"`
	BankAccount                string `json:"bank_account,omitempty"`
	Swift                      string `json:"swift,omitempty"`
	SaleType                   string `json:"sale_type,omitempty"`
	InvoiceDateKind            string `json:"invoice_date_kind,omitempty"`
	VatExemptionReason         string `json:"vat_exemption_reason,omitempty"`
	SalesKind                  string `json:"sales_kind,omitempty"`
	AmountInWords              string `json:"amount_in_words,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
	// Services is the list of line items on the invoice. See [ServiceEntry].
	Services []ServiceEntry `json:"services,omitempty"`
	// Extensions carries optional invoice features. See [Extensions].
	Extensions *Extensions `json:"extensions,omitempty"`
}

// ServiceEntry represents a line item on an invoice.
//
// Monetary fields ([ServiceEntry.UnitNetPrice], [ServiceEntry.NetPrice],
// [ServiceEntry.TaxPrice], [ServiceEntry.GrossPrice]) are expressed in grosze
// (1/100 PLN).
type ServiceEntry struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	// TaxSymbol is the VAT rate symbol applied to the line
	// (e.g., "23", "8", "5", "0", "zw", "np"). See [VatRate.Symbol].
	TaxSymbol string `json:"tax_symbol,omitempty"`
	// Unit is the unit of measure for the line (e.g., "hour", "piece",
	// "kg"). It is a free-form string as defined by inFakt.
	Unit string `json:"unit,omitempty"`
	// Quantity is the line quantity, encoded as defined by
	// https://docs.infakt.pl.
	Quantity float64 `json:"quantity,omitempty"`
	// UnitNetPrice is the per-unit net price, in grosze (1/100 PLN).
	UnitNetPrice int `json:"unit_net_price,omitempty"`
	// UnitNetPriceBeforeDiscount is the per-unit net price prior to any
	// discount, in grosze (1/100 PLN).
	UnitNetPriceBeforeDiscount int `json:"unit_net_price_before_discount,omitempty"`
	// NetPrice is the line net total, in grosze (1/100 PLN).
	NetPrice int `json:"net_price,omitempty"`
	// GrossPrice is the line gross total (net + VAT), in grosze (1/100 PLN).
	GrossPrice int `json:"gross_price,omitempty"`
	// TaxPrice is the line VAT amount, in grosze (1/100 PLN).
	TaxPrice int    `json:"tax_price,omitempty"`
	Symbol   string `json:"symbol,omitempty"`
	// PKWiU is the Polish Classification of Products and Services code
	// (Polska Klasyfikacja Wyrobów i Usług).
	PKWiU string `json:"pkwiu,omitempty"`
	// Discount is a discount factor whose encoding is defined by
	// https://docs.infakt.pl.
	Discount string `json:"discount,omitempty"`
}

// Extensions represents additional invoice settings.
type Extensions struct {
	PaymentOnline *PaymentOnline `json:"payment_online,omitempty"`
}

// PaymentOnline represents online payment configuration.
type PaymentOnline struct {
	Enabled bool `json:"enabled,omitempty"`
}

// InvoiceRequest is the partial-update payload for creating and updating
// invoices. Pointer fields let callers distinguish unset from zero-value:
// for example, leaving [InvoiceRequest.Currency] as nil means "do not
// change", while passing [String]("") sets it to the empty string. Use
// the helpers [String], [Int], [Int64], [Bool], and [Float64] from
// helpers.go to build values.
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

// ServiceEntryRequest is the partial-update payload for invoice line items.
// Pointer fields let callers distinguish unset from zero-value: for
// example, [ServiceEntryRequest.Quantity] of nil means "do not change",
// while [Float64](0) means "set to zero". Use the helpers [String], [Int],
// [Bool], and [Float64] from helpers.go to build values.
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
// [InvoiceService.List] method.
//
// Filters are translated into the inFakt query syntax q[field_op]=value.
// Common operators include cont (substring match), eq (exact match),
// gteq (>=), and lteq (<=). The Go fields below map to specific operators
// internally; see the comments next to each field.
type InvoiceListOptions struct {
	ListOptions

	// Filter fields using inFakt query format (q[field_predicate]=value)
	DateFrom string // q[invoice_date_gteq]
	DateTo   string // q[invoice_date_lteq]
	ClientID int64  // q[client_id_eq]
	Status   string // q[status_eq]
}

// InvoiceService manages invoices on the inFakt API.
//
// Supported endpoints:
//   - List, Get, Create, Update, Delete
//   - MarkAsPaid, SendByEmail, GetPDF, GetNextNumber
//
// Access it through [Client.Invoices].
//
// See https://docs.infakt.pl for the corresponding API reference.
type InvoiceService struct {
	client *Client
}

type invoiceRoot struct {
	Invoice Invoice `json:"invoice"`
}

type invoiceListRoot struct {
	MetaInfo MetaInfo  `json:"metainfo"`
	Entities []Invoice `json:"entities"`
}

// List returns invoices, paginated and filtered via [InvoiceListOptions].
// The returned [MetaInfo] reports the total count and pagination cursors.
// On API errors a typed [*ErrorResponse] is returned.
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
	err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single [Invoice] by ID. Returns [ErrNotFound] (wrapped in
// an [*ErrorResponse]) if no invoice exists with that ID.
func (s *InvoiceService) Get(ctx context.Context, id int64) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var invoice Invoice
	err = s.client.do(req, &invoice)
	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

// Create creates a new [Invoice] from the supplied prototype and returns
// the server-assigned record (with ID, number, and computed totals).
func (s *InvoiceService) Create(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices.json", apiVersion)

	root := &invoiceRoot{Invoice: *invoice}
	req, err := s.client.newRequest(ctx, http.MethodPost, path, root)
	if err != nil {
		return nil, err
	}

	var created Invoice
	err = s.client.do(req, &created)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// Update updates an existing [Invoice]. Returns [ErrNotFound] (wrapped in
// an [*ErrorResponse]) if no invoice exists with that ID.
func (s *InvoiceService) Update(ctx context.Context, id int64, invoice *Invoice) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices/%d.json", apiVersion, id)

	root := &invoiceRoot{Invoice: *invoice}
	req, err := s.client.newRequest(ctx, http.MethodPut, path, root)
	if err != nil {
		return nil, err
	}

	var updated Invoice
	err = s.client.do(req, &updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// Delete deletes an [Invoice] by ID. Returns [ErrNotFound] (wrapped in an
// [*ErrorResponse]) if no invoice exists with that ID.
func (s *InvoiceService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/%s/invoices/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}

// MarkAsPaid marks an [Invoice] as paid. The paidDate argument must be in
// YYYY-MM-DD format; passing an empty string omits the field from the
// request, letting the inFakt server use today's date.
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

	return s.client.do(req, nil)
}

// SendByEmail sends an [Invoice] to a recipient by email. The print_type
// is fixed to "original". When emailTo is empty, the email_to field is
// omitted from the request and the inFakt server applies its default
// recipient (typically the stored client email). When non-empty, the
// supplied address overrides the default.
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

	return s.client.do(req, nil)
}

// GetPDF returns the raw PDF bytes for an [Invoice]. The caller is
// responsible for writing the bytes to disk or streaming them to a
// client. Returns [ErrNotFound] (wrapped in an [*ErrorResponse]) if no
// invoice exists with that ID.
func (s *InvoiceService) GetPDF(ctx context.Context, id int64) ([]byte, error) {
	path := fmt.Sprintf("/%s/invoices/%d.pdf", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.doRaw(req)
}

// NextNumberResponse is the JSON envelope returned by the next-number
// endpoint of the inFakt API.
type NextNumberResponse struct {
	NextNumber string `json:"next_number"`
}

// GetNextNumber returns the next available invoice number. The kind
// argument selects the invoice number series (typical values: "vat",
// "proforma"); passing an empty string omits the parameter so the server
// returns the default series.
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
	err = s.client.do(req, &result)
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
