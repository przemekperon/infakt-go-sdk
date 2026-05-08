package infakt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ClientEntity represents a client (kontrahent) in the inFakt system.
//
// Address fields ([ClientEntity.Street], [ClientEntity.StreetNumber],
// [ClientEntity.FlatNumber], [ClientEntity.City], [ClientEntity.PostalCode],
// [ClientEntity.Country]) describe the client's primary address. The
// MailingCompanyName, MailingStreet, MailingCity and MailingPostalCode
// fields describe a separate correspondence address; they are populated
// when [ClientEntity.SameForwardAddress] is false.
type ClientEntity struct {
	ID          int64  `json:"id,omitempty"`
	UUID        string `json:"uuid,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
	// Street is the primary street name.
	Street string `json:"street,omitempty"`
	// StreetNumber is the building number on the primary address.
	StreetNumber string `json:"street_number,omitempty"`
	// FlatNumber is the apartment or suite number on the primary address.
	FlatNumber string `json:"flat_number,omitempty"`
	// City is the primary address city.
	City string `json:"city,omitempty"`
	// Country is the ISO country code of the primary address.
	Country         string `json:"country,omitempty"`
	CountryFullName string `json:"country_full_name,omitempty"`
	// PostalCode is the primary address postal code.
	PostalCode string `json:"postal_code,omitempty"`
	// NIP is the Polish tax identifier (Numer Identyfikacji Podatkowej),
	// 10 digits.
	NIP         string `json:"nip,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	WebSite     string `json:"web_site,omitempty"`
	Email       string `json:"email,omitempty"`
	Note        string `json:"note,omitempty"`
	// Receiver is an optional contact person at the client, printed on
	// invoices.
	Receiver string `json:"receiver,omitempty"`
	// MailingCompanyName is the company name on the correspondence address;
	// used when [ClientEntity.SameForwardAddress] is false.
	MailingCompanyName string `json:"mailing_company_name,omitempty"`
	// MailingStreet is the street on the correspondence address; used when
	// [ClientEntity.SameForwardAddress] is false.
	MailingStreet string `json:"mailing_street,omitempty"`
	// MailingCity is the city on the correspondence address; used when
	// [ClientEntity.SameForwardAddress] is false.
	MailingCity string `json:"mailing_city,omitempty"`
	// MailingPostalCode is the postal code on the correspondence address;
	// used when [ClientEntity.SameForwardAddress] is false.
	MailingPostalCode string `json:"mailing_postal_code,omitempty"`
	// DaysToPayment is the default payment term, in days. Per
	// https://docs.infakt.pl ("Klienci"), this field is an integer.
	DaysToPayment int `json:"days_to_payment,omitempty"`
	// PaymentMethod is the default payment method (see [Invoice.PaymentMethod]).
	PaymentMethod string `json:"payment_method,omitempty"`
	// InvoiceNote is a default note appended to invoices issued to this
	// client.
	InvoiceNote string `json:"invoice_note,omitempty"`
	// SameForwardAddress, when true, indicates the mailing address mirrors
	// the primary address; the Mailing* fields can be ignored.
	SameForwardAddress bool   `json:"same_forward_address,omitempty"`
	FirstName          string `json:"first_name,omitempty"`
	LastName           string `json:"last_name,omitempty"`
	// BusinessActivityKind classifies the client (e.g., "self_employed",
	// "private_person", "other_business"). See https://docs.infakt.pl.
	BusinessActivityKind string `json:"business_activity_kind,omitempty"`
}

// ClientEntityRequest is the partial-update payload for creating and
// updating client entities. Pointer fields let callers distinguish unset
// from zero-value: leaving a field as nil means "do not change", while
// passing [String]("") sets it to the empty string. Use the helpers
// [String] and [Bool] from helpers.go to build values.
type ClientEntityRequest struct {
	CompanyName          *string `json:"company_name,omitempty"`
	Street               *string `json:"street,omitempty"`
	StreetNumber         *string `json:"street_number,omitempty"`
	FlatNumber           *string `json:"flat_number,omitempty"`
	City                 *string `json:"city,omitempty"`
	Country              *string `json:"country,omitempty"`
	PostalCode           *string `json:"postal_code,omitempty"`
	NIP                  *string `json:"nip,omitempty"`
	PhoneNumber          *string `json:"phone_number,omitempty"`
	WebSite              *string `json:"web_site,omitempty"`
	Email                *string `json:"email,omitempty"`
	Note                 *string `json:"note,omitempty"`
	Receiver             *string `json:"receiver,omitempty"`
	MailingCompanyName   *string `json:"mailing_company_name,omitempty"`
	MailingStreet        *string `json:"mailing_street,omitempty"`
	MailingCity          *string `json:"mailing_city,omitempty"`
	MailingPostalCode    *string `json:"mailing_postal_code,omitempty"`
	DaysToPayment        *int    `json:"days_to_payment,omitempty"`
	PaymentMethod        *string `json:"payment_method,omitempty"`
	InvoiceNote          *string `json:"invoice_note,omitempty"`
	SameForwardAddress   *bool   `json:"same_forward_address,omitempty"`
	FirstName            *string `json:"first_name,omitempty"`
	LastName             *string `json:"last_name,omitempty"`
	BusinessActivityKind *string `json:"business_activity_kind,omitempty"`
}

// ClientEntityListOptions specifies the optional parameters to the
// [ClientEntityService.List] method.
//
// Filters are translated into the inFakt query syntax q[field_op]=value.
// Common operators include cont (substring match), eq (exact match),
// gteq (>=), and lteq (<=). The Go fields below map to specific operators
// internally; see the comments next to each field.
type ClientEntityListOptions struct {
	ListOptions

	// Filter fields using inFakt query format (q[field_predicate]=value)
	CompanyName string // q[company_name_cont]
	NIP         string // q[nip_eq]
}

// ClientEntityService manages client entities (kontrahenci) on the
// inFakt API.
//
// Supported endpoints:
//   - List, Get, Create, Update, Delete
//
// Access it through [Client.Clients].
//
// See https://docs.infakt.pl for the corresponding API reference.
type ClientEntityService struct {
	client *Client
}

type clientEntityRoot struct {
	ClientEntity ClientEntity `json:"client"`
}

type clientEntityListRoot struct {
	MetaInfo MetaInfo       `json:"metainfo"`
	Entities []ClientEntity `json:"entities"`
}

// List returns client entities, paginated and filtered via
// [ClientEntityListOptions]. The returned [MetaInfo] reports the total
// count and pagination cursors.
func (s *ClientEntityService) List(ctx context.Context, opts *ClientEntityListOptions) ([]ClientEntity, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/clients.json", apiVersion)

	if opts != nil {
		var err error
		path, err = addOptions(path, &opts.ListOptions)
		if err != nil {
			return nil, nil, err
		}
		path = addClientEntityFilters(path, opts)
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var root clientEntityListRoot
	err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single [ClientEntity] by ID. Returns [ErrNotFound]
// (wrapped in an [*ErrorResponse]) if no client exists with that ID.
func (s *ClientEntityService) Get(ctx context.Context, id int64) (*ClientEntity, error) {
	path := fmt.Sprintf("/%s/clients/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var entity ClientEntity
	err = s.client.do(req, &entity)
	if err != nil {
		return nil, err
	}

	return &entity, nil
}

// Create creates a new [ClientEntity] from the supplied prototype and
// returns the server-assigned record (with ID and UUID populated).
func (s *ClientEntityService) Create(ctx context.Context, entity *ClientEntity) (*ClientEntity, error) {
	path := fmt.Sprintf("/%s/clients.json", apiVersion)

	root := &clientEntityRoot{ClientEntity: *entity}
	req, err := s.client.newRequest(ctx, http.MethodPost, path, root)
	if err != nil {
		return nil, err
	}

	var created ClientEntity
	err = s.client.do(req, &created)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// Update updates an existing [ClientEntity]. Returns [ErrNotFound]
// (wrapped in an [*ErrorResponse]) if no client exists with that ID.
func (s *ClientEntityService) Update(ctx context.Context, id int64, entity *ClientEntity) (*ClientEntity, error) {
	path := fmt.Sprintf("/%s/clients/%d.json", apiVersion, id)

	root := &clientEntityRoot{ClientEntity: *entity}
	req, err := s.client.newRequest(ctx, http.MethodPut, path, root)
	if err != nil {
		return nil, err
	}

	var updated ClientEntity
	err = s.client.do(req, &updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// Delete deletes a [ClientEntity] by ID. Returns [ErrNotFound] (wrapped
// in an [*ErrorResponse]) if no client exists with that ID.
func (s *ClientEntityService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/%s/clients/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}

func addClientEntityFilters(path string, opts *ClientEntityListOptions) string {
	u, err := url.Parse(path)
	if err != nil {
		return path
	}

	q := u.Query()
	if opts.CompanyName != "" {
		q.Set("q[company_name_cont]", opts.CompanyName)
	}
	if opts.NIP != "" {
		q.Set("q[nip_eq]", opts.NIP)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
