package mailboxpg

import (
	"context"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/mailboxprovider"
)

func MarkInboxMessageSeen(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, key string, now int64) (bool, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO mailbox_inbox_seen (provider, mailbox_email, message_key, seen_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, mailbox_email, message_key) DO NOTHING
	`, mailboxprovider.NormalizeKey(provider), emailx.Normalize(mailboxEmail), key, now)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
