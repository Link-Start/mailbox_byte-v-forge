package main

import (
	"context"
	"fmt"
	"time"

	"github.com/byte-v-forge/common-lib/eventoutbox"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/jackc/pgx/v5"
)

const mailboxPlatformEventOutboxTable = "mailbox_platform_event_outbox"

func (s *MailboxStore) enqueueInboxOutboxEvents(ctx context.Context, tx pgx.Tx, messages []*mailboxv1.EmailInboxMessage) error {
	if len(messages) == 0 {
		return nil
	}
	records, err := mailboxPlatformEventRecords(mailboxPlatformEventSource, messages)
	if err != nil {
		return err
	}
	return enqueueMailboxOutboxEvents(ctx, tx, records)
}

func enqueueMailboxOutboxEvents(ctx context.Context, tx pgx.Tx, records []eventoutbox.Record) error {
	for _, record := range records {
		if err := eventoutbox.InsertRecordPgx(ctx, tx, mailboxPlatformEventOutboxTable, record, time.Now().Unix()); err != nil {
			return fmt.Errorf("insert mailbox outbox event: %w", err)
		}
	}
	return nil
}

func runMailboxPlatformEventOutboxWorker(ctx context.Context, store *MailboxStore, publisher *mailboxPlatformEvents) error {
	if store == nil || publisher == nil {
		return nil
	}
	return eventoutbox.RunPgxWorker(ctx, eventoutbox.PgxWorkerConfig{
		Name:      "mailbox platform event outbox",
		Beginner:  store.pool,
		Table:     mailboxPlatformEventOutboxTable,
		Publisher: publisher,
		Logf:      logWarning,
	})
}
