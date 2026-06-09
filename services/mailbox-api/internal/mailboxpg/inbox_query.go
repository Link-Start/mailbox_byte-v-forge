package mailboxpg

import (
	"context"
	"errors"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func (r *Repository) ListInboxRows(ctx context.Context, email string, limit int, receivedAfterUnix int64) ([]inboxapp.MessageRow, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	rows, err := newInboxMessageQuery().
		WhereMailbox(email).
		WhereReceivedAfter(receivedAfterUnix).
		OrderByLatest().
		Limit(normalizeInboxRowLimit(limit)).
		Query(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	return scanInboxRows(rows)
}

func (r *Repository) LatestInboxRows(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, limit int) ([]inboxapp.MessageRow, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	rows, err := newInboxMessageQuery().
		WhereMailbox(email).
		WhereReceivedAtOrAfter(issuedAfterUnix).
		WhereKeyword(subjectKeyword).
		OrderByLatest().
		Limit(normalizeInboxRowLimit(limit)).
		Query(ctx, r.pool)
	if err != nil {
		return nil, err
	}
	return scanInboxRows(rows)
}
