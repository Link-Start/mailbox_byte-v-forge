package main

import (
	"context"

	"github.com/byte-v-forge/common-lib/eventoutbox"
)

const mailboxPlatformEventOutboxTable = "mailbox_platform_event_outbox"

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
