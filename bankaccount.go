package infakt

import (
	"context"
	"fmt"
	"net/http"
)

// BankAccount represents a bank account in the inFakt system.
type BankAccount struct {
	ID            int64  `json:"id,omitempty"`
	BankName      string `json:"bank_name,omitempty"`
	AccountNumber string `json:"account_number,omitempty"`
	Swift         string `json:"swift,omitempty"`
	Currency      string `json:"currency,omitempty"`
	IsDefault     bool   `json:"is_default,omitempty"`
}

// BankAccountService handles communication with the bank account related
// methods of the inFakt API. Bank accounts are read-only.
type BankAccountService struct {
	client *Client
}

type bankAccountListRoot struct {
	MetaInfo MetaInfo      `json:"metainfo"`
	Entities []BankAccount `json:"entities"`
}

// List returns a list of bank accounts.
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
	_, err = s.client.do(req, &root)
	if err != nil {
		return nil, nil, err
	}

	return root.Entities, &root.MetaInfo, nil
}
