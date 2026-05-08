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
//
// Per-invoice operations on this resource (Get, Update, Delete, MarkAsPaid,
// SendByEmail, GetPDF) address the invoice by [Invoice.UUID] — the integer
// [Invoice.ID] is informational and is not accepted by the API in URL paths.
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
	// SplitPaymentType enables the Polish split-payment mechanism
	// (mechanizm podzielonej płatności). Allowed values per the API are
	// "required" and "optional"; an empty string omits the field.
	SplitPaymentType   string `json:"split_payment_type,omitempty"`
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
	// VatExemptionReason is the integer code identifying the legal basis
	// for VAT exemption; see https://docs.infakt.pl for the catalog.
	VatExemptionReason int    `json:"vat_exemption_reason,omitempty"`
	SalesKind          string `json:"sales_kind,omitempty"`
	AmountInWords      string `json:"amount_in_words,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	// ContinuousServiceStartOn marks the start date of a continuous
	// service period (YYYY-MM-DD).
	ContinuousServiceStartOn string `json:"continuous_service_start_on,omitempty"`
	// ContinuousServiceEndOn marks the end date of a continuous service
	// period (YYYY-MM-DD).
	ContinuousServiceEndOn string `json:"continuous_service_end_on,omitempty"`
	// BDOCode is the BDO registry number.
	BDOCode string `json:"bdo_code,omitempty"`
	// DocumentMarkingsIDs references markings (oznaczenia dokumentów) by
	// ID, used for JPK_V7 reporting.
	DocumentMarkingsIDs []int64 `json:"document_markings_ids,omitempty"`
	// ReceiptNumber is the cash-register receipt number associated with
	// this invoice.
	ReceiptNumber string `json:"receipt_number,omitempty"`
	// CheckDuplicateNumber asks the server to validate that the assigned
	// number is unique. Sent only on Create/Update.
	CheckDuplicateNumber bool `json:"check_duplicate_number,omitempty"`
	// Services is the list of line items on the invoice. See [ServiceEntry].
	Services []ServiceEntry `json:"services,omitempty"`
	// Extensions carries optional invoice features. See [Extensions].
	Extensions *Extensions `json:"extensions,omitempty"`
	// KsefData carries KSeF (Krajowy System e-Faktur) integration metadata.
	KsefData *KsefData `json:"ksef_data,omitempty"`
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
	// FlatRateTaxSymbol is the VAT rate symbol used when the seller is on
	// the Polish flat-rate tax scheme (ryczałt).
	FlatRateTaxSymbol string `json:"flat_rate_tax_symbol,omitempty"`
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
	TaxPrice int `json:"tax_price,omitempty"`
	// PKWiU is the Polish Classification of Products and Services code
	// (Polska Klasyfikacja Wyrobów i Usług).
	PKWiU string `json:"pkwiu,omitempty"`
	// CN is the Combined Nomenclature code (Nomenklatura Scalona).
	CN string `json:"cn,omitempty"`
	// PKOB is the Polish Classification of Construction Objects code.
	PKOB string `json:"pkob,omitempty"`
	// GtuID references a GTU marking; see [GtuService] for the catalog.
	GtuID int64 `json:"gtu_id,omitempty"`
	// VatDateValue lets the line override the invoice-level VAT recognition
	// date (YYYY-MM-DD). Empty means "use the invoice default".
	VatDateValue string `json:"vat_date_value,omitempty"`
	// Discount is the line-level discount as returned by the live API
	// in decimal-string form ("0.0", "10.5"). docs.infakt.pl labels the
	// field "integer (percent)" but the wire format is a string —
	// keep the SDK aligned with the actual response.
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

// KsefData carries KSeF (Krajowy System e-Faktur) integration metadata
// returned by the API for invoices that have been pushed to or pulled from
// KSeF. Refer to https://docs.infakt.pl ("KSeF") for the per-field
// semantics; the SDK preserves the raw payload as-is.
type KsefData struct {
	Status        string `json:"status,omitempty"`
	ReferenceNumber string `json:"reference_number,omitempty"`
	IssueDate     string `json:"issue_date,omitempty"`
	IssueTime     string `json:"issue_time,omitempty"`
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
	SplitPaymentType   *string               `json:"split_payment_type,omitempty"`
	VatExemptionReason *int                  `json:"vat_exemption_reason,omitempty"`
	BDOCode            *string               `json:"bdo_code,omitempty"`
	ReceiptNumber      *string               `json:"receipt_number,omitempty"`
	CheckDuplicateNumber *bool               `json:"check_duplicate_number,omitempty"`
	Services           []ServiceEntryRequest `json:"services,omitempty"`
}

// ServiceEntryRequest is the partial-update payload for invoice line items.
// Pointer fields let callers distinguish unset from zero-value: for
// example, [ServiceEntryRequest.Quantity] of nil means "do not change",
// while [Float64](0) means "set to zero". Use the helpers [String], [Int],
// [Int64], [Bool], and [Float64] from helpers.go to build values.
type ServiceEntryRequest struct {
	Name              *string  `json:"name,omitempty"`
	TaxSymbol         *string  `json:"tax_symbol,omitempty"`
	FlatRateTaxSymbol *string  `json:"flat_rate_tax_symbol,omitempty"`
	Unit              *string  `json:"unit,omitempty"`
	Quantity          *float64 `json:"quantity,omitempty"`
	UnitNetPrice      *int     `json:"unit_net_price,omitempty"`
	PKWiU             *string  `json:"pkwiu,omitempty"`
	CN                *string  `json:"cn,omitempty"`
	PKOB              *string  `json:"pkob,omitempty"`
	GtuID             *int64   `json:"gtu_id,omitempty"`
	VatDateValue      *string  `json:"vat_date_value,omitempty"`
	Discount          *string  `json:"discount,omitempty"`
}

// InvoiceListOptions specifies the optional parameters to the
// [InvoiceService.List] method.
//
// Filters are translated into the inFakt query syntax q[field_op]=value.
// Common operators include cont (substring match), eq (exact match),
// gteq (>=), and lteq (<=). The Go fields below map to specific operators
// internally; see the comments next to each field. Use
// [ListOptions.Filters] to express predicates not exposed here.
type InvoiceListOptions struct {
	ListOptions

	// Filter fields using inFakt query format (q[field_predicate]=value)
	DateFrom string // q[invoice_date_gteq]
	DateTo   string // q[invoice_date_lteq]
	ClientID int64  // q[client_id_eq]
	Status   string // q[status_eq]
}

// PDFDocumentTypeOriginal is the value to pass to [InvoiceService.GetPDF]
// (and similar endpoints) to request the original copy of an invoice.
const (
	PDFDocumentTypeOriginal  = "original"
	PDFDocumentTypeDuplicate = "duplicate"
	PDFDocumentTypeCopy      = "copy"
)

// InvoiceService manages invoices on the inFakt API.
//
// Per-invoice operations identify the invoice by [Invoice.UUID] — pass the
// UUID returned in [Invoice.UUID] from List/Create/Get to subsequent
// methods.
//
// Supported endpoints:
//   - List, Get, Create, Update, Delete
//   - MarkAsPaid (async), SendByEmail, GetPDF, GetNextNumber
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

// Get returns a single [Invoice] by UUID. Returns [ErrNotFound] (wrapped in
// an [*ErrorResponse]) if no invoice exists with that UUID.
func (s *InvoiceService) Get(ctx context.Context, uuid string) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices/%s.json", apiVersion, url.PathEscape(uuid))

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
// the server-assigned record (with ID, UUID, number, and computed totals).
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

// Update updates an existing [Invoice] addressed by UUID. Returns
// [ErrNotFound] (wrapped in an [*ErrorResponse]) if no invoice exists.
func (s *InvoiceService) Update(ctx context.Context, uuid string, invoice *Invoice) (*Invoice, error) {
	path := fmt.Sprintf("/%s/invoices/%s.json", apiVersion, url.PathEscape(uuid))

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

// Delete deletes an [Invoice] addressed by UUID. Returns [ErrNotFound]
// (wrapped in an [*ErrorResponse]) if no invoice exists.
func (s *InvoiceService) Delete(ctx context.Context, uuid string) error {
	path := fmt.Sprintf("/%s/invoices/%s.json", apiVersion, url.PathEscape(uuid))

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}

// MarkAsPaid asynchronously marks an [Invoice] as paid via the
// /async/invoices/{uuid}/paid.json endpoint. The server responds 201 and
// processes the request in the background; downstream effects (e.g. KSeF
// notifications) settle within seconds.
//
// paidDate must be in YYYY-MM-DD format; passing an empty string omits
// the field so the server defaults to today's date.
//
// Set allowCorrection to true when paying a corrective invoice — the API
// requires the flag to confirm intent on those documents.
func (s *InvoiceService) MarkAsPaid(ctx context.Context, uuid, paidDate string, allowCorrection bool) error {
	path := fmt.Sprintf("/%s/async/invoices/%s/paid.json", apiVersion, url.PathEscape(uuid))

	body := map[string]interface{}{}
	if paidDate != "" {
		body["paid_date"] = paidDate
	}
	if allowCorrection {
		body["allow_correction"] = true
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}

// SendByEmailOptions captures the optional parameters of
// [InvoiceService.SendByEmail]. Each field is optional: zero values cause
// the SDK to omit the corresponding key and let the inFakt server apply
// its defaults.
type SendByEmailOptions struct {
	// PrintType selects the document variant to attach. Allowed values
	// per https://docs.infakt.pl: "original", "duplicate", "copy".
	// Defaults to [PDFDocumentTypeOriginal] when empty.
	PrintType string
	// Locale forces the language of the rendered PDF, e.g. "pl" or "en".
	// Defaults to the account language when empty.
	Locale string
	// Recipient overrides the default recipient address (the client's
	// stored email).
	Recipient string
	// SendCopy, when true, asks the server to BCC the issuer.
	SendCopy bool
}

// SendByEmail delivers an [Invoice] by email via
// /invoices/{uuid}/deliver_via_email.json. Pass nil to use server defaults
// (original PDF, account locale, stored client email).
func (s *InvoiceService) SendByEmail(ctx context.Context, uuid string, opts *SendByEmailOptions) error {
	path := fmt.Sprintf("/%s/invoices/%s/deliver_via_email.json", apiVersion, url.PathEscape(uuid))

	if opts == nil {
		opts = &SendByEmailOptions{}
	}
	printType := opts.PrintType
	if printType == "" {
		printType = PDFDocumentTypeOriginal
	}

	body := map[string]interface{}{
		"print_type": printType,
	}
	if opts.Locale != "" {
		body["locale"] = opts.Locale
	}
	if opts.Recipient != "" {
		body["recipient"] = opts.Recipient
	}
	if opts.SendCopy {
		body["send_copy"] = true
	}

	req, err := s.client.newRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}

// GetPDF returns the raw PDF bytes for an [Invoice]. The endpoint name
// (.json) is misleading: the live API responds with Content-Type
// application/pdf and the document body inline. Callers are responsible
// for writing the bytes to disk or streaming them onward.
//
// documentType is required by the API; use the [PDFDocumentTypeOriginal] /
// [PDFDocumentTypeDuplicate] / [PDFDocumentTypeCopy] constants. An empty
// value defaults to [PDFDocumentTypeOriginal]. locale ("pl", "en", ...)
// is optional; pass an empty string to use the account default.
//
// Returns [ErrNotFound] (wrapped in an [*ErrorResponse]) if no invoice
// exists with that UUID, [ErrUnprocessableEntity] for a rejected
// documentType.
func (s *InvoiceService) GetPDF(ctx context.Context, uuid, documentType, locale string) ([]byte, error) {
	path := fmt.Sprintf("/%s/invoices/%s/pdf.json", apiVersion, url.PathEscape(uuid))

	if documentType == "" {
		documentType = PDFDocumentTypeOriginal
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("document_type", documentType)
	if locale != "" {
		q.Set("locale", locale)
	}
	req.URL.RawQuery = q.Encode()

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
