package mailboxpg

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
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
