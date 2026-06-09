package mailboxpg

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxprovider"
)

func InsertInboxMessage(ctx context.Context, tx pgx.Tx, msg PersistInboxMessage, now int64) error {
	recipientsJSON, err := json.Marshal(inboxapp.UniqueEmails(msg.Recipients))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, inboxMessageInsertSQL,
		mailboxprovider.NormalizeKey(msg.Provider), emailx.Normalize(msg.MailboxEmail), msg.Key, strings.TrimSpace(msg.ID),
		strings.TrimSpace(msg.Subject), emailx.Normalize(msg.FromAddress), strings.TrimSpace(msg.BodyPreview),
		strings.TrimSpace(msg.BodyText), strings.TrimSpace(msg.HTMLBody), msg.RawSize, msg.ReceivedAtUnix,
		string(recipientsJSON), emailx.Normalize(msg.SourceEmail), now)
	return err
}
