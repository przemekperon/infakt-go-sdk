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
// Get, Update, SendByEmail, MarkAsPaid, and GetPDF accept either a UUID
// or a numeric ID encoded as a string.
type Invoice struct {
	LocalGovernmentRecipientAddress *LocalGovernmentAddress `json:"local_government_recipient_address,omitempty"`
	LocalGovernmentSellerAddress    *LocalGovernmentAddress `json:"local_government_seller_address,omitempty"`

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
	DocumentMarkingsIDs []string `json:"document_markings_ids,omitempty"`
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
	// FlatRateTaxSymbol is the income-tax flat-rate symbol (ryczałt),
	// separate from the VAT rate in TaxSymbol.
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

// Extensions represents additional invoice response data.
type Extensions struct {
	Payments *PaymentExtension `json:"payments,omitempty"`
	Shares   *ShareExtension   `json:"shares,omitempty"`
	// Deprecated: payment_online is not part of the current documented response.
	PaymentOnline *PaymentOnline `json:"payment_online,omitempty"`
}

// PaymentOnline represents legacy online payment configuration.
// Deprecated: current invoice request schemas do not document payment_online.
type PaymentOnline struct {
	Enabled bool `json:"enabled,omitempty"`
}

// KsefData carries KSeF (Krajowy System e-Faktur) integration metadata
// returned by the API for invoices that have been pushed to or pulled from
// KSeF. Refer to https://docs.infakt.pl ("KSeF") for the per-field
// semantics; the SDK preserves the raw payload as-is.
type KsefData struct {
	Status          string `json:"status,omitempty"`
	ReferenceNumber string `json:"reference_number,omitempty"`
	IssueDate       string `json:"issue_date,omitempty"`
	IssueTime       string `json:"issue_time,omitempty"`
}

// InvoiceRequest is the partial-update payload for creating and updating
// invoices. Pointer fields let callers distinguish unset from zero-value:
// for example, leaving [InvoiceRequest.Currency] as nil means "do not
// change", while passing [String]("") sets it to the empty string. Use
// the helpers [String], [Int], [Int64], [Bool], and [Float64] from
// helpers.go to build values.
type InvoiceRequest struct {
	Number                          *string                        `json:"number,omitempty"`
	CheckDuplicateNumber            *bool                          `json:"check_duplicate_number,omitempty"`
	Kind                            *string                        `json:"kind,omitempty"`
	Currency                        *string                        `json:"currency,omitempty"`
	PaidPrice                       *int                           `json:"paid_price,omitempty"`
	LeftToPay                       *int                           `json:"left_to_pay,omitempty"`
	Notes                           *string                        `json:"notes,omitempty"`
	PaymentMethod                   *string                        `json:"payment_method,omitempty"`
	SplitPayment                    *bool                          `json:"split_payment,omitempty"`
	SplitPaymentType                *string                        `json:"split_payment_type,omitempty"`
	RecipientSignature              *string                        `json:"recipient_signature,omitempty"`
	SellerSignature                 *string                        `json:"seller_signature,omitempty"`
	InvoiceDate                     *string                        `json:"invoice_date,omitempty"`
	SaleDate                        *string                        `json:"sale_date,omitempty"`
	VatDateValue                    *string                        `json:"vat_date_value,omitempty"`
	ContinuousServiceStartOn        *string                        `json:"continuous_service_start_on,omitempty"`
	ContinuousServiceEndOn          *string                        `json:"continuous_service_end_on,omitempty"`
	OccasionalSale                  *bool                          `json:"occasional_sale,omitempty"`
	PaymentDate                     *string                        `json:"payment_date,omitempty"`
	PaidDate                        *string                        `json:"paid_date,omitempty"`
	ClientID                        *int64                         `json:"client_id,omitempty"`
	ClientCompanyName               *string                        `json:"client_company_name,omitempty"`
	ClientFirstName                 *string                        `json:"client_first_name,omitempty"`
	ClientLastName                  *string                        `json:"client_last_name,omitempty"`
	ClientBusinessActivityKind      *string                        `json:"client_business_activity_kind,omitempty"`
	ClientStreet                    *string                        `json:"client_street,omitempty"`
	ClientStreetNumber              *string                        `json:"client_street_number,omitempty"`
	ClientFlatNumber                *string                        `json:"client_flat_number,omitempty"`
	ClientCity                      *string                        `json:"client_city,omitempty"`
	ClientPostCode                  *string                        `json:"client_post_code,omitempty"`
	ClientTaxCode                   *string                        `json:"client_tax_code,omitempty"`
	ClientCountry                   *string                        `json:"client_country,omitempty"`
	BankAccount                     *string                        `json:"bank_account,omitempty"`
	BankName                        *string                        `json:"bank_name,omitempty"`
	Swift                           *string                        `json:"swift,omitempty"`
	VatExemptionReason              *int                           `json:"vat_exemption_reason,omitempty"`
	SaleType                        *string                        `json:"sale_type,omitempty"`
	InvoiceDateKind                 *string                        `json:"invoice_date_kind,omitempty"`
	DocumentMarkingsIDs             *[]string                      `json:"document_markings_ids,omitempty"`
	TransactionKindID               *int64                         `json:"transaction_kind_id,omitempty"`
	BDOCode                         *string                        `json:"bdo_code,omitempty"`
	ReceiptNumber                   *string                        `json:"receipt_number,omitempty"`
	NotIncome                       *bool                          `json:"not_income,omitempty"`
	VatExchangeDateKind             *string                        `json:"vat_exchange_date_kind,omitempty"`
	SalesKind                       *string                        `json:"sales_kind,omitempty"`
	KsefNumber                      *string                        `json:"ksef_number,omitempty"`
	TransactionTerms                *TransactionTermsRequest       `json:"transaction_terms,omitempty"`
	LocalGovernmentRecipientAddress *LocalGovernmentAddressRequest `json:"local_government_recipient_address,omitempty"`
	LocalGovernmentSellerAddress    *LocalGovernmentAddressRequest `json:"local_government_seller_address,omitempty"`
	ThirdPartyAddresses             *[]ThirdPartyAddressRequest    `json:"third_party_addresses,omitempty"`
	ScheduledDates                  *[]string                      `json:"scheduled_dates,omitempty"`
	AllowCorrection                 *bool                          `json:"allow_correction,omitempty"`
	// Services replaces the complete list. Nil omits the list; an empty non-nil
	// slice sends []. Business acceptance of an empty list is not guaranteed.
	Services *[]ServiceEntryRequest `json:"services,omitempty"`
	// RemoveLocalGovernmentRecipientAddress explicitly consents to the API's
	// destructive omission semantics. Required on UpdateWithRequest when no address is supplied.
	RemoveLocalGovernmentRecipientAddress bool `json:"-"`
}

// ServiceEntryRequest describes one complete target line item. Sending Services
// replaces every existing item; it does not patch individual items. Name and
// TaxSymbol are required by the API. Nil fields use API defaults, not old values.
type ServiceEntryRequest struct {
	Name                       *string  `json:"name,omitempty"`
	TaxSymbol                  *string  `json:"tax_symbol,omitempty"`
	Unit                       *string  `json:"unit,omitempty"`
	Quantity                   *float64 `json:"quantity,omitempty"`
	UnitNetPrice               *int     `json:"unit_net_price,omitempty"`
	NetPrice                   *int     `json:"net_price,omitempty"`
	GrossPrice                 *int     `json:"gross_price,omitempty"`
	TaxPrice                   *int     `json:"tax_price,omitempty"`
	Symbol                     *string  `json:"symbol,omitempty"`
	PKWiU                      *string  `json:"pkwiu,omitempty"`
	CN                         *string  `json:"cn,omitempty"`
	PKOB                       *string  `json:"pkob,omitempty"`
	GTIN                       *string  `json:"gtin,omitempty"`
	FlatRateTaxSymbol          *string  `json:"flat_rate_tax_symbol,omitempty"`
	Discount                   *float64 `json:"discount,omitempty"`
	UnitNetPriceBeforeDiscount *int     `json:"unit_net_price_before_discount,omitempty"`
	GtuID                      *int64   `json:"gtu_id,omitempty"`
	RelatedID                  *int64   `json:"related_id,omitempty"`
	UserCalculated             *bool    `json:"user_calculated,omitempty"`
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
// Get, Update, SendByEmail, MarkAsPaid, and GetPDF accept a UUID or a numeric
// ID encoded as a string. Endpoint paths are specific to each action.
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

// Get returns a single [Invoice] by UUID or numeric ID encoded as a string. Returns [ErrNotFound] (wrapped in
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
// Deprecated: use CreateWithRequest to send explicit zero values.
func (s *InvoiceService) Create(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	if invoice == nil {
		return nil, fmt.Errorf("infakt: invoice must not be nil")
	}
	if invoice.LocalGovernmentRecipientAddress != nil || invoice.LocalGovernmentSellerAddress != nil {
		return nil, fmt.Errorf("infakt: use CreateWithRequest for JST addresses; response and request address fields differ")
	}

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

// Update is retained for source compatibility and always returns a local
// migration error, because response models cannot express safe update intent.
//
// Deprecated: use UpdateWithRequest for explicit fields and address-removal intent.
func (s *InvoiceService) Update(ctx context.Context, uuid string, invoice *Invoice) (*Invoice, error) {
	return nil, fmt.Errorf("infakt: legacy Update cannot express JST recipient-address intent; migrate to UpdateWithRequest and explicitly supply LocalGovernmentRecipientAddress or set RemoveLocalGovernmentRecipientAddress")
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

// GetPDF returns the rendered invoice PDF as raw bytes. The inFakt endpoint
// (GET /v3/invoices/{uuid}/pdf.json) responds with the document directly
// (Content-Type: application/pdf), so the returned slice is the complete PDF,
// ready to be written to disk or streamed to a client.
//
// documentType selects the variant ("original", "duplicate", "copy"); use the
// [PDFDocumentTypeOriginal] / [PDFDocumentTypeDuplicate] / [PDFDocumentTypeCopy]
// constants (an empty string defaults to "original"). locale ("pl", "en", ...)
// is optional; pass an empty string to use the account default.
//
// Returns [ErrNotFound] (wrapped in an [*ErrorResponse]) if no invoice exists
// with that UUID.
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

// LocalGovernmentAddressRequest is a JST recipient or seller address payload. Required fields
// are always serialized, including their zero values. CountryID defaults to Poland.
type LocalGovernmentAddressRequest struct {
	CompanyName    string `json:"company_name"`
	Street         string `json:"street"`
	BuildingNumber string `json:"building_number"`
	City           string `json:"city"`
	TaxID          string `json:"tax_id,omitempty"`
	DoorNumber     string `json:"door_number,omitempty"`
	PostCode       string `json:"post_code,omitempty"`
	CountryID      *int64 `json:"country_id,omitempty"`
}

// TransactionTermsRequest describes contracts and orders associated with an invoice.
type TransactionTermsRequest struct {
	Contracts *[]TransactionReferenceRequest `json:"contracts,omitempty"`
	Orders    *[]TransactionReferenceRequest `json:"orders,omitempty"`
}

// TransactionReferenceRequest identifies a contract or order and its date.
type TransactionReferenceRequest struct {
	Number string `json:"number"`
	Date   string `json:"date"`
}

// ThirdPartyAddressRequest describes an invoice third-party address.
type ThirdPartyAddressRequest struct {
	ID                 *int64  `json:"id,omitempty"`
	Destroy            *bool   `json:"_destroy,omitempty"`
	CompanyName        *string `json:"company_name,omitempty"`
	TaxID              *string `json:"tax_id,omitempty"`
	Street             *string `json:"street,omitempty"`
	BuildingNumber     *string `json:"building_number,omitempty"`
	DoorNumber         *string `json:"door_number,omitempty"`
	PostCode           *string `json:"post_code,omitempty"`
	City               *string `json:"city,omitempty"`
	CountryCode        *string `json:"country_code,omitempty"`
	CountryID          *int64  `json:"country_id,omitempty"`
	RoleCode           *string `json:"role_code,omitempty"`
	InternalIdentifier *string `json:"internal_identifier,omitempty"`
}

// PaymentExtension contains the availability and link of an online payment.
type PaymentExtension struct {
	Link      *string `json:"link,omitempty"`
	Available bool    `json:"available"`
}

// ShareExtension contains a public sharing link and its validity limit.
type ShareExtension struct {
	Link       *string `json:"link,omitempty"`
	Available  bool    `json:"available"`
	ValidUntil *string `json:"valid_until,omitempty"`
}

// CreateWithRequest creates an invoice using explicit optional fields.
func (s *InvoiceService) CreateWithRequest(ctx context.Context, payload *InvoiceRequest) (*Invoice, error) {
	if payload == nil {
		return nil, fmt.Errorf("infakt: request must not be nil")
	}
	return s.writeRequest(ctx, http.MethodPost, fmt.Sprintf("/%s/invoices.json", apiVersion), payload)
}

// UpdateWithRequest updates an invoice by UUID or numeric ID encoded as a string.
// Supply the complete current/new JST recipient address or explicitly consent to
// its removal by setting RemoveLocalGovernmentRecipientAddress. No implicit GET
// or merge occurs; callers must coordinate concurrent edits themselves.
func (s *InvoiceService) UpdateWithRequest(ctx context.Context, id string, payload *InvoiceRequest) (*Invoice, error) {
	if payload == nil {
		return nil, fmt.Errorf("infakt: request must not be nil")
	}
	if payload.LocalGovernmentRecipientAddress == nil && !payload.RemoveLocalGovernmentRecipientAddress {
		return nil, fmt.Errorf("infakt: explicitly supply LocalGovernmentRecipientAddress or set RemoveLocalGovernmentRecipientAddress")
	}
	if payload.LocalGovernmentRecipientAddress != nil && payload.RemoveLocalGovernmentRecipientAddress {
		return nil, fmt.Errorf("infakt: recipient address and removal intent are mutually exclusive")
	}
	return s.writeRequest(ctx, http.MethodPut, fmt.Sprintf("/%s/invoices/%s.json", apiVersion, url.PathEscape(id)), payload)
}

func (s *InvoiceService) writeRequest(ctx context.Context, method, path string, payload *InvoiceRequest) (*Invoice, error) {
	if payload.Services != nil && *payload.Services == nil {
		return nil, fmt.Errorf("infakt: services must be an array, not null; omit it or supply a non-nil slice")
	}
	root := struct {
		Invoice *InvoiceRequest `json:"invoice"`
	}{payload}
	req, err := s.client.newRequest(ctx, method, path, root)
	if err != nil {
		return nil, err
	}
	var result Invoice
	if err := s.client.do(req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// LocalGovernmentAddress is a JST address returned by the API. Response field
// names differ from LocalGovernmentAddressRequest. For an update, copy the
// address into a request and resolve CountryID using the API country catalog.
type LocalGovernmentAddress struct {
	CompanyName  *string `json:"company_name,omitempty"`
	NIP          *string `json:"nip,omitempty"`
	Street       *string `json:"street,omitempty"`
	StreetNumber *string `json:"street_number,omitempty"`
	FlatNumber   *string `json:"flat_number,omitempty"`
	PostalCode   *string `json:"postal_code,omitempty"`
	City         *string `json:"city,omitempty"`
	Country      *string `json:"country,omitempty"`
}
