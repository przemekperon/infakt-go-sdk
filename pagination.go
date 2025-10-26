package infakt

import (
	"fmt"
	"net/url"
)

// ListOptions specifies the optional parameters to various List methods.
type ListOptions struct {
	Offset int `json:"offset,omitempty"`
	Limit  int `json:"limit,omitempty"`
}

// MetaInfo contains pagination metadata from the API response.
type MetaInfo struct {
	Count      int    `json:"count"`
	TotalCount int    `json:"total_count"`
	Next       string `json:"next"`
	Previous   string `json:"previous"`
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
