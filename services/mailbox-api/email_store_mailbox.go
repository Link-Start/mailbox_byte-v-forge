package main

import (
	"context"
	"errors"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/randx"
	"github.com/jackc/pgx/v5"

	"mailboxapi/pb"
)

func (s *MailboxStore) UpsertMailbox(ctx context.Context, mailbox *pb.EmailMailbox) (*pb.EmailMailbox, error) {
	if mailbox == nil {
		return nil, errors.New("mailbox is required")
	}
	email := emailx.Normalize(mailbox.GetEmailAddress())
	if email == "" {
		return nil, errors.New("email_address is required")
	}
	requestedProvider := normalizeEmailProvider(mailbox.GetProviderKey())
	insertProvider := requestedProvider
	if insertProvider == "" {
		insertProvider = defaultMailboxProvider()
	}
	rowID, err := randx.Hex(16)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
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
	if err := mailboxProviderUpsert(ctx, tx, persistedProvider, mailbox, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return s.FindMailbox(ctx, email)
}
