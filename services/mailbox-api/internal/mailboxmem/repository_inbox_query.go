package mailboxmem

import (
	"context"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/pagex"
)

func (r *Repository) ListInboxRows(ctx context.Context, email string, limit int, receivedAfterUnix int64) ([]inboxapp.MessageRow, error) {
	return r.filterInboxRows(ctx, email, "", receivedAfterUnix, false, limit)
}

func (r *Repository) ListInboxPageRows(ctx context.Context, email string, limit int, cursor string, keyword string) (inboxapp.MessageRowPage, error) {
	limit = normalizeInboxRowLimit(limit)
	offset, err := pagex.ParseOffsetToken(cursor)
	if err != nil {
		return inboxapp.MessageRowPage{}, err
	}
	rows, err := r.filterInboxRows(ctx, email, keyword, 0, false, limit+offset+1)
	if err != nil {
		return inboxapp.MessageRowPage{}, err
	}
	if offset >= len(rows) {
		return inboxapp.MessageRowPage{}, nil
	}
	rows, hasMore := pagex.TrimLimit(rows[offset:], limit)
	nextCursor := ""
	if hasMore {
		nextCursor = pagex.OffsetToken(offset + limit)
	}
	return inboxapp.MessageRowPage{Rows: rows, NextCursor: nextCursor}, nil
}

func (r *Repository) LatestInboxRows(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, limit int) ([]inboxapp.MessageRow, error) {
	return r.filterInboxRows(ctx, email, subjectKeyword, issuedAfterUnix, true, limit)
}
