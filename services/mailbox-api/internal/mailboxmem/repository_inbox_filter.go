package mailboxmem

import (
	"context"
	"errors"
	"sort"
	"strings"

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
		if message.row.MailboxEmail != email {
			continue
		}
		if includeEqual {
			if message.row.ReceivedAtUnix < timestamp {
				continue
			}
			if !messageMatchesKeyword(message.row, keyword) {
				continue
			}
		} else if message.row.ReceivedAtUnix <= timestamp {
			continue
		}
		rows = append(rows, message)
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

func messageMatchesKeyword(row inboxapp.MessageRow, keyword string) bool {
	keyword = strings.ToLower(strings.TrimSpace(keyword))
	if keyword == "" {
		return true
	}
	return strings.Contains(strings.ToLower(row.Subject), keyword) ||
		strings.Contains(strings.ToLower(row.BodyPreview), keyword) ||
		strings.Contains(strings.ToLower(row.BodyText), keyword)
}

func sortStoredMessages(rows []storedMessage) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].row.ReceivedAtUnix == rows[j].row.ReceivedAtUnix {
			if rows[i].updatedAt == rows[j].updatedAt {
				return rows[i].key > rows[j].key
			}
			return rows[i].updatedAt > rows[j].updatedAt
		}
		return rows[i].row.ReceivedAtUnix > rows[j].row.ReceivedAtUnix
	})
}

func normalizeInboxRowLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
