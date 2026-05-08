package infakt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CostInvoice mirrors the JSON shape returned by
// `GET /v3/documents/costs.json` and `GET /v3/documents/costs/{uuid}.json`
// — the read-only portion of the inFakt cost-invoice (faktury kosztowe)
// resource.
//
// Per https://docs.infakt.pl ("Koszty"), per-resource operations address a
// cost invoice by [CostInvoice.UUID]; the SDK only exposes the read-only
// surface (List + Get) at this time. Monetary fields ([CostInvoice.NetPrice],
// [CostInvoice.GrossPrice], [CostInvoice.TaxPrice]) are expressed in grosze
// (1/100 PLN) when non-nil. Date fields are formatted as YYYY-MM-DD strings;
// [CostInvoice.CreatedAt] follows the API's "YYYY-MM-DD HH:MM:SS ±HHMM"
// format.
//
// Pointer fields ([CostInvoice.AccountedAt], [CostInvoice.SellerName],
// [CostInvoice.SellerTaxCode], [CostInvoice.SellerBankAccount],
// [CostInvoice.Description], [CostInvoice.Category],
// [CostInvoice.ReconciliationID]) are nullable in the API and may be
// JSON `null`; treat a nil pointer as "absent".
type CostInvoice struct {
	UUID     string `json:"uuid,omitempty"`
	Number   string `json:"number,omitempty"`
	Currency string `json:"currency,omitempty"`
	// NetPrice is the net amount, in grosze (1/100 PLN).
	NetPrice *int `json:"net_price,omitempty"`
	// GrossPrice is the gross amount (net + VAT), in grosze (1/100 PLN).
	GrossPrice *int `json:"gross_price,omitempty"`
	// TaxPrice is the VAT amount, in grosze (1/100 PLN).
	TaxPrice *int `json:"tax_price,omitempty"`
	// IssueDate is the document's issue date (YYYY-MM-DD).
	IssueDate string `json:"issue_date,omitempty"`
	// ReceivedDate is the date the document was received (YYYY-MM-DD).
	ReceivedDate string `json:"received_date,omitempty"`
	// DueDate is the payment due date (YYYY-MM-DD).
	DueDate string `json:"due_date,omitempty"`
	// AccountedAt is the date the cost was booked into the ledger
	// (YYYY-MM-DD); nil while the cost is still unaccounted.
	AccountedAt *string `json:"accounted_at,omitempty"`
	// SellerName is the supplier's name; nil when the cost was uploaded
	// without a parsed seller.
	SellerName *string `json:"seller_name,omitempty"`
	// SellerTaxCode is the supplier's NIP / VAT number; nil when not
	// known.
	SellerTaxCode *string `json:"seller_tax_code,omitempty"`
	// SellerBankAccount is the supplier's bank account number; nil when
	// not provided.
	SellerBankAccount *string `json:"seller_bank_account,omitempty"`
	// Description is a free-form note on the cost; nil when empty.
	Description *string `json:"description,omitempty"`
	// AddedBy is the user-readable identifier of who added the cost.
	AddedBy string `json:"added_by,omitempty"`
	// Category is the cost-category label (e.g. "Wydatki przedsiębiorcy");
	// nil when uncategorised.
	Category *string `json:"category,omitempty"`
	// Kind classifies the cost (e.g. "document_scan", "ksef").
	Kind string `json:"kind,omitempty"`
	// NotesCount is the number of notes attached to the cost.
	NotesCount int `json:"notes_count,omitempty"`
	// Duplicated, when true, marks the cost as a detected duplicate.
	Duplicated bool `json:"duplicated,omitempty"`
	// CreatedAt is the document creation timestamp in the API format
	// "YYYY-MM-DD HH:MM:SS ±HHMM".
	CreatedAt string `json:"created_at,omitempty"`
	// Source identifies how the cost entered the system (e.g. "ksef",
	// "upload").
	Source string `json:"source,omitempty"`
	// ReconciliationID references the reconciliation record this cost
	// belongs to; nil when not yet reconciled.
	ReconciliationID *int64 `json:"reconciliation_id,omitempty"`
	// Statuses is the set of workflow statuses currently applied to the
	// cost (e.g. accounting / payment status groups).
	Statuses []CostInvoiceStatus `json:"statuses,omitempty"`
}

// CostInvoiceStatus is a single entry in [CostInvoice.Statuses].
type CostInvoiceStatus struct {
	// Symbol is the machine identifier (e.g. "cost_not_accounted").
	Symbol string `json:"symbol,omitempty"`
	// Name is the human-readable label.
	Name string `json:"name,omitempty"`
	// Group is the status group (e.g. "cost_accounting").
	Group string `json:"group,omitempty"`
}

// CostInvoiceListOptions specifies the optional parameters to the
// [CostInvoiceService.List] method.
//
// Filters are translated into the inFakt query syntax q[field_op]=value.
// See https://docs.infakt.pl ("Filtrowanie"). Use [ListOptions.Filters] to
// express predicates not exposed here.
type CostInvoiceListOptions struct {
	ListOptions

	// DateFrom narrows the listing to costs issued on or after this date
	// (q[issue_date_gteq]). Format: YYYY-MM-DD.
	DateFrom string
	// DateTo narrows the listing to costs issued on or before this date
	// (q[issue_date_lteq]). Format: YYYY-MM-DD.
	DateTo string

	// SellerTaxCode filters by supplier NIP / VAT number
	// (q[seller_tax_code_eq]).
	SellerTaxCode string
	// Currency filters by ISO 4217 currency code (q[currency_eq]).
	Currency string
}

// CostInvoiceService provides read-only access to cost invoices
// (faktury kosztowe) on the inFakt API.
//
// Supported endpoints:
//   - List (`GET /v3/documents/costs.json`)
//   - Get  (`GET /v3/documents/costs/{uuid}.json`)
//
// The mutating endpoints documented under https://docs.infakt.pl ("Koszty")
// — upload, paid_many, destroy_many, etc. — are not yet exposed by the SDK.
//
// Access it through [Client.CostInvoices].
type CostInvoiceService struct {
	client *Client
}

type costInvoiceListRoot struct {
	MetaInfo MetaInfo      `json:"metainfo"`
	Entities []CostInvoice `json:"entities"`
}

// List returns cost invoices, paginated and filtered via
// [CostInvoiceListOptions]. The returned [MetaInfo] reports the total
// count and pagination cursors. On API errors a typed [*ErrorResponse]
// is returned.
func (s *CostInvoiceService) List(ctx context.Context, opts *CostInvoiceListOptions) ([]CostInvoice, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/documents/costs.json", apiVersion)

	if opts != nil {
		var err error
		path, err = addOptions(path, &opts.ListOptions)
		if err != nil {
			return nil, nil, err
		}
		path = addCostInvoiceFilters(path, opts)
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var root costInvoiceListRoot
	if err := s.client.do(req, &root); err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single [CostInvoice] by UUID. Returns [ErrNotFound]
// (wrapped in an [*ErrorResponse]) if no cost invoice exists with that
// UUID.
func (s *CostInvoiceService) Get(ctx context.Context, uuid string) (*CostInvoice, error) {
	path := fmt.Sprintf("/%s/documents/costs/%s.json", apiVersion, url.PathEscape(uuid))

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var cost CostInvoice
	if err := s.client.do(req, &cost); err != nil {
		return nil, err
	}

	return &cost, nil
}

func addCostInvoiceFilters(path string, opts *CostInvoiceListOptions) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}

	q := u.Query()
	if opts.DateFrom != "" {
		q.Set("q[issue_date_gteq]", opts.DateFrom)
	}
	if opts.DateTo != "" {
		q.Set("q[issue_date_lteq]", opts.DateTo)
	}
	if opts.SellerTaxCode != "" {
		q.Set("q[seller_tax_code_eq]", opts.SellerTaxCode)
	}
	if opts.Currency != "" {
		q.Set("q[currency_eq]", opts.Currency)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
