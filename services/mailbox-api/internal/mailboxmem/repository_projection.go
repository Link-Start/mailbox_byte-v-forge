package mailboxmem

import (
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/internal/redactx"
)

const mailboxErrorSnippetLimit = 600

func (r *Repository) project(record *mailboxmodel.Record) *mailboxmodel.Record {
	record = cloneRecord(record)
	if record == nil {
		return nil
	}
	record.ProviderKey = r.providers.NormalizeProviderInput(record.GetProviderKey())
	record.Domain = domainForEmail(record.GetEmailAddress())
	r.providers.PrepareProjection(record)
	return record
}

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

func domainForEmail(email string) string {
	_, domain, ok := strings.Cut(emailx.Normalize(email), "@")
	if !ok {
		return ""
	}
	return domain
}

func safeText(value string) string {
	return redactx.TextSnippet(value, mailboxErrorSnippetLimit)
}
