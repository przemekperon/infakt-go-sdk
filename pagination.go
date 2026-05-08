package infakt

import (
	"fmt"
	"net/url"
)

// ListOptions specifies the optional pagination parameters accepted by the
// various List methods on this package's services. It is embedded into the
// resource-specific *ListOptions types (e.g. [InvoiceListOptions],
// [ClientEntityListOptions], [ProductListOptions]) so callers can set offset
// and limit alongside resource-specific filters.
type ListOptions struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
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
		q.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}
	if opts.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
