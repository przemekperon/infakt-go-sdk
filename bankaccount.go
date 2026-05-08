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

// BankAccountService manages bank accounts on the inFakt API. Bank
// accounts are read-only.
//
// Supported endpoints:
//   - List
//
// Access it through [Client.BankAccounts].
//
// See https://docs.infakt.pl for the corresponding API reference.
type BankAccountService struct {
	client *Client
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
