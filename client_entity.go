package infakt

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// ClientEntity represents a client (kontrahent) in the inFakt system.
type ClientEntity struct {
	ID                int64  `json:"id,omitempty"`
	CompanyName       string `json:"company_name,omitempty"`
	Street            string `json:"street,omitempty"`
	City              string `json:"city,omitempty"`
	Country           string `json:"country,omitempty"`
	PostalCode        string `json:"postal_code,omitempty"`
	NIP               string `json:"nip,omitempty"`
	PhoneNumber       string `json:"phone_number,omitempty"`
	Email             string `json:"email,omitempty"`
	Note              string `json:"note,omitempty"`
	InvoiceNote       string `json:"invoice_note,omitempty"`
	PaymentDays       int    `json:"payment_days,omitempty"`
	PersonName        string `json:"person_name,omitempty"`
	BankAccount       string `json:"bank_account,omitempty"`
	TaxPayer          bool   `json:"tax_payer,omitempty"`
	ReceivingMethod   string `json:"receiving_method,omitempty"`
	PaymentMethod     string `json:"payment_method,omitempty"`
	SameForwardAddress bool  `json:"same_forward_address,omitempty"`
}

// ClientEntityRequest is used for creating and updating client entities.
// Pointer fields allow distinguishing between zero values and unset fields.
type ClientEntityRequest struct {
	CompanyName       *string `json:"company_name,omitempty"`
	Street            *string `json:"street,omitempty"`
	City              *string `json:"city,omitempty"`
	Country           *string `json:"country,omitempty"`
	PostalCode        *string `json:"postal_code,omitempty"`
	NIP               *string `json:"nip,omitempty"`
	PhoneNumber       *string `json:"phone_number,omitempty"`
	Email             *string `json:"email,omitempty"`
	Note              *string `json:"note,omitempty"`
	InvoiceNote       *string `json:"invoice_note,omitempty"`
	PaymentDays       *int    `json:"payment_days,omitempty"`
	PersonName        *string `json:"person_name,omitempty"`
	BankAccount       *string `json:"bank_account,omitempty"`
	TaxPayer          *bool   `json:"tax_payer,omitempty"`
	ReceivingMethod   *string `json:"receiving_method,omitempty"`
	PaymentMethod     *string `json:"payment_method,omitempty"`
	SameForwardAddress *bool  `json:"same_forward_address,omitempty"`
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
	_, err = s.client.do(req, &root)
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
	_, err = s.client.do(req, &entity)
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
	_, err = s.client.do(req, &created)
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
	_, err = s.client.do(req, &updated)
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

	_, err = s.client.do(req, nil)
	return err
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
