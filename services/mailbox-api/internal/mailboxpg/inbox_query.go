package mailboxpg

import (
	"context"
	"errors"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/pagex"
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

func (r *Repository) ListInboxPageRows(ctx context.Context, email string, limit int, cursor string, keyword string) (inboxapp.MessageRowPage, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return inboxapp.MessageRowPage{}, errors.New("email_address is required")
	}
	offset, err := pagex.ParseOffsetToken(cursor)
	if err != nil {
		return inboxapp.MessageRowPage{}, err
	}
	limit = normalizeInboxRowLimit(limit)
	rows, err := newInboxMessageQuery().
		WhereMailbox(email).
		WhereKeyword(keyword).
		OrderByLatest().
		Limit(limit+1).
		Offset(offset).
		Query(ctx, r.pool)
	if err != nil {
		return inboxapp.MessageRowPage{}, err
	}
	items, err := scanInboxRows(rows)
	if err != nil {
		return inboxapp.MessageRowPage{}, err
	}
	items, hasMore := pagex.TrimLimit(items, limit)
	nextCursor := ""
	if hasMore {
		nextCursor = pagex.OffsetToken(offset + limit)
	}
	return inboxapp.MessageRowPage{Rows: items, NextCursor: nextCursor}, nil
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
