package main

import (
	"context"
	"fmt"
	"time"

	"github.com/byte-v-forge/common-lib/eventbus"
	"github.com/byte-v-forge/common-lib/eventoutbox"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/jackc/pgx/v5"
)

const mailboxPlatformEventOutboxTable = "mailbox_platform_event_outbox"

func (s *MailboxStore) enqueueInboxOutboxEvents(ctx context.Context, tx pgx.Tx, messages []*mailboxv1.EmailInboxMessage) error {
	if len(messages) == 0 {
		return nil
	}
	return enqueueMailboxOutboxEvents(ctx, tx, mailboxPlatformEventMessages(mailboxPlatformEventSource, messages))
}

func enqueueMailboxOutboxEvents(ctx context.Context, tx pgx.Tx, messages []eventbus.Message) error {
	for _, message := range messages {
		record, err := eventoutbox.NewRecord(message)
		if err != nil {
			return fmt.Errorf("prepare mailbox outbox event: %w", err)
		}
		if err := eventoutbox.InsertRecordPgx(ctx, tx, mailboxPlatformEventOutboxTable, record, time.Now().Unix()); err != nil {
			return err
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
