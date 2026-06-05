package mailboxpg

import (
	"context"
	"fmt"

	"github.com/byte-v-forge/common-lib/eventoutbox"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/inboxapp"
)

func enqueueInboxOutboxRecords(ctx context.Context, tx pgx.Tx, table string, source string, messages []*mailboxv1.EmailInboxMessage, now int64) error {
	if table == "" || len(messages) == 0 {
		return nil
	}
	records, err := inboxapp.EventRecords(source, messages)
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := eventoutbox.InsertRecordPgx(ctx, tx, table, record, now); err != nil {
			return fmt.Errorf("insert mailbox outbox event: %w", err)
		}
	}
	return nil
}
