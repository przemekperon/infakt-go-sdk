package infakt

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// MaxPageLimit is the upper bound the inFakt API enforces for the per-page
// `limit` query parameter. Per https://docs.infakt.pl (sekcja "Stronicowanie"
// i "Limity"), values above 100 are clamped server-side; the SDK clamps them
// before the request to keep client- and server-side behavior identical.
const MaxPageLimit = 100

// ListOptions specifies the optional pagination, sorting, field-selection
// and filtering parameters accepted by the various List methods on this
// package's services. It is embedded into the resource-specific *ListOptions
// types (e.g. [InvoiceListOptions], [ClientEntityListOptions],
// [ProductListOptions]) so callers can set them alongside resource-specific
// filters.
type ListOptions struct {
	// Offset is the zero-based index of the first record to return.
	Offset int `json:"offset,omitempty"`
	// Limit is the number of records per page; values are clamped to
	// [MaxPageLimit] (100) before the request.
	Limit int `json:"limit,omitempty"`
	// Order is an optional sort expression in the inFakt format
	// "field direction", e.g. "name asc" or "id desc".
	Order string `json:"-"`
	// Fields restricts the response to a subset of attributes. Each entry
	// is a top-level field name (e.g. "name") or a nested expression
	// (e.g. "services(name,tax_symbol)"). Joined with commas in the
	// "fields" query parameter.
	Fields []string `json:"-"`
	// Filters carries Ransack-style query predicates. Each map entry is
	// rendered as q[<key>]=<value>, so {"number_eq": "1/2026"} produces
	// "q[number_eq]=1/2026". Use this to express filters not exposed by
	// resource-specific *ListOptions fields.
	Filters map[string]string `json:"-"`
}

// MetaInfo contains pagination metadata returned in the JSON envelope of
// every List response. It is paired with the entity slice; together they
// allow callers to drive offset/limit pagination loops by consulting
// [MetaInfo.TotalCount] and the cursor URLs in [MetaInfo.Next] /
// [MetaInfo.Previous]. See also [ListOptions].
type MetaInfo struct {
	// Count is the number of entities returned in the current page.
	Count int `json:"count"`
	// TotalCount is the total number of entities matching the query
	// across all pages.
	TotalCount int `json:"total_count"`
	// Next is an opaque cursor URL for the next page, or the empty
	// string when there is no next page. Treat this value as opaque —
	// the SDK does not parse it, and callers typically advance pages
	// by incrementing [ListOptions.Offset] instead.
	Next string `json:"next"`
	// Previous is an opaque cursor URL for the previous page, or the
	// empty string when on the first page. As with [MetaInfo.Next],
	// treat the value as opaque.
	Previous string `json:"previous"`
}

// addOptions adds the parameters in opts as URL query parameters to s.
func addOptions(s string, opts *ListOptions) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, fmt.Errorf("infakt: invalid URL %q: %w", s, err)
	}

	if opts == nil {
		return s, nil
	}

	q := u.Query()
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}
	limit := opts.Limit
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if opts.Order != "" {
		q.Set("order", opts.Order)
	}
	if len(opts.Fields) > 0 {
		q.Set("fields", strings.Join(opts.Fields, ","))
	}
	for k, v := range opts.Filters {
		q.Set("q["+k+"]", v)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
