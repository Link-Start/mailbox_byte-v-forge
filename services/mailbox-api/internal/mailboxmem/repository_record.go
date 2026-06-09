package mailboxmem

import (
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func cloneRecord(record *mailboxmodel.Record) *mailboxmodel.Record {
	if record == nil {
		return nil
	}
	return &mailboxmodel.Record{
		EmailAddress: record.EmailAddress,
		Password:     record.Password,
		RefreshToken: record.RefreshToken,
		AccessToken:  record.AccessToken,
		LastError:    record.LastError,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
		AuthStatus:   record.AuthStatus,
		ProviderKey:  record.ProviderKey,
		LatestSignal: record.LatestSignal,
		Domain:       record.Domain,
	}
}

func providerRecord(record *mailboxmodel.Record) mailboxprovider.MailboxRecord {
	if record == nil {
		return mailboxprovider.MailboxRecord{}
	}
	return mailboxprovider.MailboxRecord{
		Email:        record.EmailAddress,
		Provider:     record.ProviderKey,
		RefreshToken: record.RefreshToken,
		AuthStatus:   record.AuthStatus,
	}
}
