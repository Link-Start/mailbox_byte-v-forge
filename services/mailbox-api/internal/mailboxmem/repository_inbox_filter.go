package mailboxmem

import (
	"context"
	"errors"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func (r *Repository) filterInboxRows(ctx context.Context, email string, keyword string, timestamp int64, includeEqual bool, limit int) ([]inboxapp.MessageRow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	r.mu.RLock()
	rows := []storedMessage{}
	for _, message := range r.messages {
		if inboxFilterMatches(message.row, email, keyword, timestamp, includeEqual) {
			rows = append(rows, message)
		}
	}
	r.mu.RUnlock()
	sortStoredMessages(rows)
	limit = normalizeInboxRowLimit(limit)
	if len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]inboxapp.MessageRow, 0, len(rows))
	for _, message := range rows {
		out = append(out, message.row)
	}
	return out, nil
}

func inboxFilterMatches(row inboxapp.MessageRow, email string, keyword string, timestamp int64, includeEqual bool) bool {
	if row.MailboxEmail != email {
		return false
	}
	if includeEqual {
		return row.ReceivedAtUnix >= timestamp && messageMatchesKeyword(row, keyword)
	}
	return row.ReceivedAtUnix > timestamp && messageMatchesKeyword(row, keyword)
}
