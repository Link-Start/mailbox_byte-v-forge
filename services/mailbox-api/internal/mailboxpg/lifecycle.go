package mailboxpg

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventoutbox"

	"mailboxapi/internal/mailboxprovider"
)

func OpenRepository(ctx context.Context, dsn string, providers *mailboxprovider.Registry, outboxTable string) (*Repository, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("mailbox postgres DSN is required")
	}
	if providers == nil {
		return nil, errors.New("mailbox providers are required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	repo, err := NewRepository(pool, providers)
	if err != nil {
		pool.Close()
		return nil, err
	}
	if err := repo.EnsureSchema(ctx, outboxTable); err != nil {
		pool.Close()
		return nil, err
	}
	return repo, nil
}

func (r *Repository) Close() {
	if r != nil && r.pool != nil {
		r.pool.Close()
	}
}

func (r *Repository) RunOutboxWorker(ctx context.Context, table string, publisher eventbus.Publisher, logf func(string, ...any)) error {
	if r == nil || r.pool == nil || publisher == nil {
		return nil
	}
	return eventoutbox.RunPgxWorker(ctx, eventoutbox.PgxWorkerConfig{
		Name:      "mailbox platform event outbox",
		Beginner:  r.pool,
		Table:     table,
		Publisher: publisher,
		Logf:      logf,
	})
}
