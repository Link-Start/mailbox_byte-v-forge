package mailboxpg

import (
	"context"
	"errors"
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
