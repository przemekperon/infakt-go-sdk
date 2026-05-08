package infakt

import (
	"context"
	"fmt"
	"net/http"
)

// BankAccount represents a bank account in the inFakt system.
type BankAccount struct {
	ID       int64  `json:"id,omitempty"`
	BankName string `json:"bank_name,omitempty"`
	// AccountNumber is the bank account number. For Polish accounts it is
	// in IBAN format (26 digits, no spaces).
	AccountNumber string `json:"account_number,omitempty"`
	// Swift is the BIC/SWIFT code used for international transfers.
	Swift string `json:"swift,omitempty"`
	// Currency is the ISO 4217 currency code of the account
	// (e.g., "PLN", "EUR").
	Currency string `json:"currency,omitempty"`
	// Default, when true, marks this account as the default to use on new
	// invoices.
	Default bool `json:"default,omitempty"`
	// CustomName is an optional user-provided alias for the account.
	CustomName string `json:"custom_name,omitempty"`
}

// BankAccountService manages bank accounts on the inFakt API.
//
// Supported endpoints:
//   - List, Get, Create, Update, Delete
//
// Note: a newly created account must be verified in the inFakt web panel
// before it can be referenced from an invoice. Mutating endpoints require
// the api:sensitive:bank_accounts:write scope on the API key.
//
// Access it through [Client.BankAccounts].
//
// See https://docs.infakt.pl for the corresponding API reference.
type BankAccountService struct {
	client *Client
}

type bankAccountRoot struct {
	BankAccount BankAccount `json:"bank_account"`
}

type bankAccountListRoot struct {
	MetaInfo MetaInfo      `json:"metainfo"`
	Entities []BankAccount `json:"entities"`
}

// List returns bank accounts, paginated via [ListOptions]. The returned
// [MetaInfo] reports the total count and pagination cursors.
func (s *BankAccountService) List(ctx context.Context, opts *ListOptions) ([]BankAccount, *MetaInfo, error) {
	path := fmt.Sprintf("/%s/bank_accounts.json", apiVersion)

	if opts != nil {
		var err error
		path, err = addOptions(path, opts)
		if err != nil {
			return nil, nil, err
		}
	}

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, nil, err
	}

	var root bankAccountListRoot
	err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}

// Get returns a single [BankAccount] by ID. Returns [ErrNotFound] (wrapped
// in an [*ErrorResponse]) if no account exists with that ID.
func (s *BankAccountService) Get(ctx context.Context, id int64) (*BankAccount, error) {
	path := fmt.Sprintf("/%s/bank_accounts/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var account BankAccount
	err = s.client.do(req, &account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// Create creates a new [BankAccount] from the supplied prototype and returns
// the server-assigned record. The account must be verified in the inFakt
// panel before it can be used on an invoice.
func (s *BankAccountService) Create(ctx context.Context, account *BankAccount) (*BankAccount, error) {
	path := fmt.Sprintf("/%s/bank_accounts.json", apiVersion)

	root := &bankAccountRoot{BankAccount: *account}
	req, err := s.client.newRequest(ctx, http.MethodPost, path, root)
	if err != nil {
		return nil, err
	}

	var created BankAccount
	err = s.client.do(req, &created)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

// Update updates an existing [BankAccount]. Returns [ErrNotFound] (wrapped
// in an [*ErrorResponse]) if no account exists with that ID.
func (s *BankAccountService) Update(ctx context.Context, id int64, account *BankAccount) (*BankAccount, error) {
	path := fmt.Sprintf("/%s/bank_accounts/%d.json", apiVersion, id)

	root := &bankAccountRoot{BankAccount: *account}
	req, err := s.client.newRequest(ctx, http.MethodPut, path, root)
	if err != nil {
		return nil, err
	}

	var updated BankAccount
	err = s.client.do(req, &updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// Delete removes a [BankAccount] by ID. Returns [ErrNotFound] (wrapped in
// an [*ErrorResponse]) if no account exists with that ID.
func (s *BankAccountService) Delete(ctx context.Context, id int64) error {
	path := fmt.Sprintf("/%s/bank_accounts/%d.json", apiVersion, id)

	req, err := s.client.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return s.client.do(req, nil)
}
