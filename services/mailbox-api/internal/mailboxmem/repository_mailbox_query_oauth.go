package mailboxmem

import (
	"context"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Repository) ListOAuthMailboxes(ctx context.Context, limit int32) ([]*mailboxmodel.Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := oauthMailboxLimit(limit)
	r.mu.RLock()
	rows := make([]*mailboxmodel.Record, 0, len(r.mailboxes))
	for _, entry := range r.mailboxes {
		record := cloneRecord(entry.record)
		if record == nil || record.AuthStatus != mailboxmodel.AuthStatusAuthorized {
			continue
		}
		if r.providers.ValidatePoll(providerRecord(record)) != nil {
			continue
		}
		rows = append(rows, r.project(record))
	}
	r.mu.RUnlock()
	sortMailboxRows(rows)
	if len(rows) > n {
		rows = rows[:n]
	}
	return rows, nil
}

func oauthMailboxLimit(limit int32) int {
	n := int(limit)
	if n <= 0 {
		n = 100
	}
	if n > 500 {
		n = 500
	}
	return n
}
