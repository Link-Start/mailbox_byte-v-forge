package mailboxmem

import (
	"sort"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/pagex"
)

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
