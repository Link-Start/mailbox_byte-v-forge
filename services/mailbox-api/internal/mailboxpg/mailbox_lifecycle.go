package mailboxpg

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/randx"

	"mailboxapi/internal/mailboxmodel"
)

func (r *Repository) UpsertMailbox(ctx context.Context, mailbox *mailboxmodel.Record) (*mailboxmodel.Record, error) {
	if mailbox == nil {
		return nil, errors.New("mailbox is required")
	}
	email := emailx.Normalize(mailbox.GetEmailAddress())
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	requestedProvider := r.providers.NormalizeProviderInput(mailbox.GetProviderKey())
	insertProvider := requestedProvider
	if insertProvider == "" {
		insertProvider = r.providers.DefaultKey()
	}
	rowID, err := randx.Hex(16)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var persistedProvider string
	if err := tx.QueryRow(ctx, `
		INSERT INTO mailboxes (id, email, provider, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$4)
		ON CONFLICT (email) DO UPDATE SET
			provider = CASE WHEN $5 <> '' THEN EXCLUDED.provider ELSE mailboxes.provider END,
			updated_at = EXCLUDED.updated_at
		RETURNING provider
	`, rowID, email, insertProvider, now, requestedProvider).Scan(&persistedProvider); err != nil {
		return nil, err
	}
	if err := r.upsertProviderMailboxData(ctx, tx, persistedProvider, mailbox, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.FindMailbox(ctx, email)
}

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

func (r *Repository) UpdateMailboxTokens(ctx context.Context, email string, refreshToken string, accessToken string) error {
	email = emailx.Normalize(email)
	row, err := r.newMailboxSelectQuery().WhereEmail(email).ScanOne(ctx, r.pool)
	if err != nil {
		return err
	}
	if err := r.updateMailboxTokens(ctx, row.Provider, email, refreshToken, accessToken); err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, "UPDATE mailboxes SET updated_at = $1 WHERE email = $2", time.Now().Unix(), email)
	return err
}

func (r *Repository) DeleteMailbox(ctx context.Context, email string) (bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = lockStoredMailbox(ctx, tx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		deleted, deleteErr := deleteMailboxInbox(ctx, tx, []string{email})
		if deleteErr != nil {
			return false, deleteErr
		}
		if err := tx.Commit(ctx); err != nil {
			return false, err
		}
		return deleted, nil
	}
	if err != nil {
		return false, err
	}

	deleteEmails := []string{email}
	if _, err := deleteMailboxInbox(ctx, tx, deleteEmails); err != nil {
		return false, err
	}
	args, inClause := sqlInArgs(deleteEmails)
	tag, err := tx.Exec(ctx, "DELETE FROM mailboxes WHERE email IN ("+inClause+")", args...)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
