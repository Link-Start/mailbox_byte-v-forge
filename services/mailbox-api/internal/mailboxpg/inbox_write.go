package mailboxpg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/inboxapp"
)

func (r *Repository) RecordMessages(ctx context.Context, request inboxapp.RecordMessagesRequest) ([]*mailboxv1.EmailInboxMessage, error) {
	provider := r.providers.NormalizeProviderInput(request.Provider)
	if provider == "" {
		return nil, fmt.Errorf("email provider is required")
	}
	if len(request.Messages) == 0 {
		return []*mailboxv1.EmailInboxMessage{}, nil
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().Unix()
	write, err := r.persistInboxInputs(ctx, tx, provider, request, now)
	if err != nil {
		return nil, err
	}
	if err := UpdateInboxWatermarks(ctx, tx, write.watermarks, now); err != nil {
		return nil, err
	}
	if err := r.pruneInbound(ctx, tx, provider, write.retention); err != nil {
		return nil, err
	}
	if err := enqueueInboxOutboxRecords(ctx, tx, request.OutboxTable, request.EventSource, write.unseen, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return write.unseen, nil
}
