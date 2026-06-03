package mailboxpg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"
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

func (r *Repository) ListInboxRows(ctx context.Context, email string, limit int, receivedAfterUnix int64) ([]InboxMessageRow, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	n := normalizeInboxRowLimit(limit)
	args := []any{email}
	query := inboxMessageSelectSQL + "WHERE mailbox_email = $1"
	if receivedAfterUnix > 0 {
		args = append(args, receivedAfterUnix)
		query += fmt.Sprintf(" AND received_at > $%d", len(args))
	}
	args = append(args, n)
	query += fmt.Sprintf(`
		ORDER BY received_at DESC, updated_at DESC, message_key DESC
		LIMIT $%d
	`, len(args))
	return r.queryInboxRows(ctx, query, args...)
}

func (r *Repository) LatestInboxRows(ctx context.Context, email string, subjectKeyword string, issuedAfterUnix int64, limit int) ([]InboxMessageRow, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	args := []any{email}
	query := inboxMessageSelectSQL + "WHERE mailbox_email = $1"
	if issuedAfterUnix > 0 {
		args = append(args, issuedAfterUnix)
		query += fmt.Sprintf(" AND received_at >= $%d", len(args))
	}
	if keyword := strings.TrimSpace(subjectKeyword); keyword != "" {
		args = append(args, "%"+keyword+"%")
		query += fmt.Sprintf(" AND (subject ILIKE $%d OR body_preview ILIKE $%d OR body_text ILIKE $%d)", len(args), len(args), len(args))
	}
	args = append(args, normalizeInboxRowLimit(limit))
	query += fmt.Sprintf(" ORDER BY received_at DESC, updated_at DESC, message_key DESC LIMIT $%d", len(args))
	return r.queryInboxRows(ctx, query, args...)
}

func (r *Repository) queryInboxRows(ctx context.Context, query string, args ...any) ([]InboxMessageRow, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []InboxMessageRow{}
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
