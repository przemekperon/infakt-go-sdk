package infakt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ClientEntity represents a client (kontrahent) in the inFakt system.
type ClientEntity struct {
	ID                   int64  `json:"id,omitempty"`
	UUID                 string `json:"uuid,omitempty"`
	CompanyName          string `json:"company_name,omitempty"`
	Street               string `json:"street,omitempty"`
	StreetNumber         string `json:"street_number,omitempty"`
	FlatNumber           string `json:"flat_number,omitempty"`
	City                 string `json:"city,omitempty"`
	Country              string `json:"country,omitempty"`
	CountryFullName      string `json:"country_full_name,omitempty"`
	PostalCode           string `json:"postal_code,omitempty"`
	NIP                  string `json:"nip,omitempty"`
	PhoneNumber          string `json:"phone_number,omitempty"`
	WebSite              string `json:"web_site,omitempty"`
	Email                string `json:"email,omitempty"`
	Note                 string `json:"note,omitempty"`
	Receiver             string `json:"receiver,omitempty"`
	MailingCompanyName   string `json:"mailing_company_name,omitempty"`
	MailingStreet        string `json:"mailing_street,omitempty"`
	MailingCity          string `json:"mailing_city,omitempty"`
	MailingPostalCode    string `json:"mailing_postal_code,omitempty"`
	DaysToPayment        string `json:"days_to_payment,omitempty"`
	PaymentMethod        string `json:"payment_method,omitempty"`
	InvoiceNote          string `json:"invoice_note,omitempty"`
	SameForwardAddress   bool   `json:"same_forward_address,omitempty"`
	FirstName            string `json:"first_name,omitempty"`
	LastName             string `json:"last_name,omitempty"`
	BusinessActivityKind string `json:"business_activity_kind,omitempty"`
}

// ClientEntityRequest is used for creating and updating client entities.
// Pointer fields allow distinguishing between zero values and unset fields.
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
	DaysToPayment        *string `json:"days_to_payment,omitempty"`
	PaymentMethod        *string `json:"payment_method,omitempty"`
	InvoiceNote          *string `json:"invoice_note,omitempty"`
	SameForwardAddress   *bool   `json:"same_forward_address,omitempty"`
	FirstName            *string `json:"first_name,omitempty"`
	LastName             *string `json:"last_name,omitempty"`
	BusinessActivityKind *string `json:"business_activity_kind,omitempty"`
}

// ClientEntityListOptions specifies the optional parameters to the
// ClientEntityService.List method.
type ClientEntityListOptions struct {
	ListOptions

	// Filter fields using inFakt query format (q[field_predicate]=value)
	CompanyName string // q[company_name_cont]
	NIP         string // q[nip_eq]
}

// ClientEntityService handles communication with the client entity
// related methods of the inFakt API.
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

// List returns a list of client entities.
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

// Get returns a single client entity by ID.
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

// Create creates a new client entity.
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

// Update updates an existing client entity.
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

// Delete deletes a client entity.
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
