package mailboxmem

import (
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

func (r *Repository) listVirtualMailboxesLocked(query mailboxprovider.ListQuery) []*mailboxmodel.Record {
	rows := []*mailboxmodel.Record{}
	for _, source := range r.providers.VirtualMailboxSources() {
		if query.Provider != "" && query.Provider != source.Key() {
			continue
		}
		if !source.StoredInboxOnly() || !source.IncludeVirtual(query.AuthStatus) {
			continue
		}
		rows = append(rows, r.virtualMailboxesForProviderLocked(source.Key(), query)...)
	}
	return rows
}

func (r *Repository) virtualMailboxesForProviderLocked(provider string, query mailboxprovider.ListQuery) []*mailboxmodel.Record {
	byEmail := map[string]*mailboxmodel.Record{}
	for _, message := range r.messages {
		if message.row.Provider != provider {
			continue
		}
		email := emailx.Normalize(message.row.MailboxEmail)
		if email == "" || r.mailboxes[email].record != nil {
			continue
		}
		record := byEmail[email]
		if record == nil {
			record = &mailboxmodel.Record{
				EmailAddress: email,
				ProviderKey:  provider,
				CreatedAt:    message.createdAt,
				UpdatedAt:    message.updatedAt,
				Domain:       domainForEmail(email),
			}
			byEmail[email] = record
		}
		if message.createdAt < record.CreatedAt {
			record.CreatedAt = message.createdAt
		}
		if message.updatedAt > record.UpdatedAt {
			record.UpdatedAt = message.updatedAt
		}
	}
	rows := []*mailboxmodel.Record{}
	for _, record := range byEmail {
		if matchesMailboxQuery(record, query) {
			rows = append(rows, r.project(record))
		}
	}
	return rows
}
