package mailboxpg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"

	"mailboxapi/internal/inboxapp"
)

func (r *Repository) InboxWatermark(ctx context.Context, email string) (int64, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return 0, errors.New("email_address is required")
	}
	var watermark int64
	err := r.pool.QueryRow(ctx, "SELECT last_inbox_received_at_ns FROM mailboxes WHERE email = $1", email).Scan(&watermark)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	return watermark, err
}

func (r *Repository) HasInboxMessages(ctx context.Context, email string) (bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM mailbox_inbox_messages WHERE mailbox_email = $1
		)
	`, email).Scan(&exists)
	return exists, err
}

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

func scanInboxRows(rows pgx.Rows) ([]inboxapp.MessageRow, error) {
	defer rows.Close()

	out := []inboxapp.MessageRow{}
	for rows.Next() {
		row, err := scanInboxMessageRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func normalizeInboxRowLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
