package mailboxpg

import (
	"fmt"
	"strings"
	"time"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

type providerTokenUpdate struct {
	table       string
	emailColumn string
	assignments []string
	args        []any
}

func (r *Repository) updateMailboxTokenStatement(provider string, refreshToken string, accessToken string) (providerTokenUpdate, error) {
	fields, err := r.providerTokenFields(provider)
	if err != nil {
		return providerTokenUpdate{}, err
	}
	statement, err := newProviderTokenUpdate(fields, refreshToken, accessToken)
	if err != nil {
		return providerTokenUpdate{}, err
	}
	if err := statement.addOptional(fields.AuthStatusColumn, mailboxmodel.AuthStatusAuthorized); err != nil {
		return providerTokenUpdate{}, err
	}
	if err := statement.addOptional(fields.LastErrorColumn, ""); err != nil {
		return providerTokenUpdate{}, err
	}
	if err := statement.addOptional(fields.UpdatedAtColumn, time.Now().Unix()); err != nil {
		return providerTokenUpdate{}, err
	}
	return statement, nil
}

func (r *Repository) providerTokenFields(provider string) (mailboxprovider.TokenFields, error) {
	definition := r.providers.StorageByKey(provider)
	if definition == nil {
		return mailboxprovider.TokenFields{}, fmt.Errorf("mailbox provider has no token storage: %s", provider)
	}
	fields, ok := definition.TokenFields()
	if !ok {
		return mailboxprovider.TokenFields{}, fmt.Errorf("mailbox provider has no token storage: %s", provider)
	}
	return fields, nil
}

func (u providerTokenUpdate) assignmentsSQL() string {
	return strings.Join(u.assignments, ", ")
}
