package mailboxpg

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxprovider"
)

type PersistInboxMessage struct {
	Key            string
	ID             string
	MailboxEmail   string
	Subject        string
	FromAddress    string
	BodyPreview    string
	ReceivedAtUnix int64
	Recipients     []string
	Provider       string
	SourceEmail    string
	BodyText       string
	HTMLBody       string
	RawSize        int64
}

func InsertInboxMessage(ctx context.Context, tx pgx.Tx, msg PersistInboxMessage, now int64) error {
	recipientsJSON, err := json.Marshal(uniqueEmails(msg.Recipients))
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO mailbox_inbox_messages (
			provider, mailbox_email, message_key, message_id, subject, from_address,
			body_preview, body_text, html_body, raw_size, received_at, recipients_json,
			source_mailbox_email,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
		ON CONFLICT (provider, mailbox_email, message_key) DO UPDATE SET
			message_id = EXCLUDED.message_id,
			subject = EXCLUDED.subject,
			from_address = EXCLUDED.from_address,
			body_preview = EXCLUDED.body_preview,
			body_text = EXCLUDED.body_text,
			html_body = EXCLUDED.html_body,
			raw_size = EXCLUDED.raw_size,
			received_at = EXCLUDED.received_at,
			recipients_json = EXCLUDED.recipients_json,
			source_mailbox_email = EXCLUDED.source_mailbox_email,
			updated_at = EXCLUDED.updated_at
	`, mailboxprovider.NormalizeKey(msg.Provider), emailx.Normalize(msg.MailboxEmail), msg.Key, strings.TrimSpace(msg.ID),
		strings.TrimSpace(msg.Subject), emailx.Normalize(msg.FromAddress), strings.TrimSpace(msg.BodyPreview),
		strings.TrimSpace(msg.BodyText), strings.TrimSpace(msg.HTMLBody), msg.RawSize, msg.ReceivedAtUnix,
		string(recipientsJSON), emailx.Normalize(msg.SourceEmail), now)
	return err
}

func uniqueEmails(values []string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, value := range values {
		trimmed := emailx.Normalize(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
