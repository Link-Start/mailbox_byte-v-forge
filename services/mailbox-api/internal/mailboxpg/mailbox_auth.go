package mailboxpg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Repository) MarkEmailAuthStatus(ctx context.Context, email string, authStatus string, lastError string) (*mailboxmodel.Record, error) {
	email = emailx.Normalize(email)
	authStatus = strings.TrimSpace(authStatus)
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	if authStatus == "" {
		return nil, errors.New("auth_status is required")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	provider, err := lockStoredMailbox(ctx, tx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	if err := r.updateProviderAuth(ctx, tx, provider, email, authStatus, safeText(lastError), now); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, "UPDATE mailboxes SET updated_at = $1 WHERE email = $2", now, email); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindMailbox(ctx, email)
}
