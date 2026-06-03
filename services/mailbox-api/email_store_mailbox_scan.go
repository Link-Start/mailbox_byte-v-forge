package main

import (
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxpg"
)

func (s *MailboxStore) recordFromMailboxRow(row *mailboxpg.MailboxRow) *mailboxmodel.Record {
	return row.ToRecord(s.providers.NormalizeProviderInput, s.providers.PrepareProjection)
}

func normalizeEmailProvider(provider string) string {
	return normalizeMailboxProviderInput(provider)
}

func domainForEmail(email string) string {
	return mailboxpg.DomainForEmail(email)
}
