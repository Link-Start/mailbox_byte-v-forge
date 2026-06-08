package mailboxmem

import (
	"sort"
	"strings"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
	"mailboxapi/internal/pagex"
)

func (r *Repository) listStoredMailboxesLocked(query mailboxprovider.ListQuery) []*mailboxmodel.Record {
	rows := []*mailboxmodel.Record{}
	for _, entry := range r.mailboxes {
		record := cloneRecord(entry.record)
		if !matchesMailboxQuery(record, query) {
			continue
		}
		rows = append(rows, r.project(record))
	}
	return rows
}

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

func matchesMailboxQuery(record *mailboxmodel.Record, query mailboxprovider.ListQuery) bool {
	if record == nil {
		return false
	}
	if query.AuthStatus != "" && strings.TrimSpace(record.GetAuthStatus()) != query.AuthStatus {
		return false
	}
	if query.Provider != "" && mailboxprovider.NormalizeKey(record.GetProviderKey()) != query.Provider {
		return false
	}
	if query.EmailAddress != "" && emailx.Normalize(record.GetEmailAddress()) != query.EmailAddress {
		return false
	}
	if query.HasCursor() {
		updatedAt := record.GetUpdatedAt()
		cursorAt := query.Cursor.UpdatedAt.Unix()
		if updatedAt > cursorAt || (updatedAt == cursorAt && record.GetEmailAddress() >= query.Cursor.ID) {
			return false
		}
	}
	return true
}

func mailboxPageFromRows(rows []*mailboxmodel.Record, limit int) mailboxmodel.ListPage {
	rows = uniqueMailboxRows(rows)
	sortMailboxRows(rows)
	page := pagex.NewKeysetPage(rows, limit, func(mailbox *mailboxmodel.Record) pagex.KeysetCursor {
		return pagex.KeysetCursorValue(time.Unix(mailbox.GetUpdatedAt(), 0).UTC(), mailbox.GetEmailAddress())
	})
	return mailboxmodel.ListPage{Mailboxes: page.Items, NextCursor: page.NextCursor}
}

func uniqueMailboxRows(rows []*mailboxmodel.Record) []*mailboxmodel.Record {
	seen := map[string]struct{}{}
	out := make([]*mailboxmodel.Record, 0, len(rows))
	for _, row := range rows {
		email := emailx.Normalize(row.GetEmailAddress())
		if email == "" {
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, row)
	}
	return out
}

func sortMailboxRows(rows []*mailboxmodel.Record) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].GetUpdatedAt() == rows[j].GetUpdatedAt() {
			return rows[i].GetEmailAddress() > rows[j].GetEmailAddress()
		}
		return rows[i].GetUpdatedAt() > rows[j].GetUpdatedAt()
	})
}
