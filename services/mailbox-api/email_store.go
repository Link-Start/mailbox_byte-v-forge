package main

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/internal/mailboxpg"
	"mailboxapi/internal/mailboxprovider"
)

type MailboxStore struct {
	pool      *pgxpool.Pool
	mailboxes *mailboxpg.Repository
}

func NewMailboxStore(ctx context.Context, dsn string, providers *mailboxprovider.Registry) (*MailboxStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("PG_DSN is required")
	}
	if providers == nil {
		return nil, errors.New("mailbox providers are required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	mailboxes, err := mailboxpg.NewRepository(pool, providers)
	if err != nil {
		pool.Close()
		return nil, err
	}
	store := &MailboxStore{pool: pool, mailboxes: mailboxes}
	if err := mailboxes.EnsureSchema(ctx, mailboxPlatformEventOutboxTable); err != nil {
		pool.Close()
		return nil, err
	}
	return store, nil
}

func (s *MailboxStore) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
