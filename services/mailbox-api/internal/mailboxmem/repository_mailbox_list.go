package mailboxmem

import (
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
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
