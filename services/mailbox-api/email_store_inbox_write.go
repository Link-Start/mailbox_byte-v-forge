package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/hashx"
	"github.com/jackc/pgx/v5"
)

func insertInboxMessage(ctx context.Context, tx pgx.Tx, msg inboxPersistMessage, now int64) error {
	recipientsJSON, err := json.Marshal(uniqueStrings(msg.recipients))
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
	`, normalizeEmailProvider(msg.provider), emailx.Normalize(msg.mailboxEmail), msg.key, strings.TrimSpace(msg.id),
		strings.TrimSpace(msg.subject), emailx.Normalize(msg.fromAddress), strings.TrimSpace(msg.bodyPreview),
		strings.TrimSpace(msg.bodyText), strings.TrimSpace(msg.htmlBody), msg.rawSize, msg.receivedAtUnix,
		string(recipientsJSON), emailx.Normalize(msg.sourceEmail), now)
	return err
}

func pruneMailboxMessages(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, limit int) error {
	provider = normalizeEmailProvider(provider)
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if provider == "" || mailboxEmail == "" || limit <= 0 {
		return nil
	}
	keys, err := expiredInboxKeys(ctx, tx, `
		SELECT provider, mailbox_email, message_key
		FROM mailbox_inbox_messages
		WHERE provider = $1 AND mailbox_email = $2
		ORDER BY received_at DESC, updated_at DESC, message_key DESC
		OFFSET $3
	`, provider, mailboxEmail, limit)
	if err != nil {
		return err
	}
	return deleteInboxKeys(ctx, tx, keys)
}

func pruneDomainMessages(ctx context.Context, tx pgx.Tx, provider string, domain string, limit int) error {
	provider = normalizeEmailProvider(provider)
	domain = strings.Trim(strings.ToLower(strings.TrimSpace(domain)), ".")
	if provider == "" || domain == "" || limit <= 0 {
		return nil
	}
	keys, err := expiredInboxKeys(ctx, tx, `
		SELECT provider, mailbox_email, message_key
		FROM mailbox_inbox_messages
		WHERE provider = $1 AND split_part(mailbox_email, '@', 2) = $2
		ORDER BY received_at DESC, updated_at DESC, message_key DESC
		OFFSET $3
	`, provider, domain, limit)
	if err != nil {
		return err
	}
	return deleteInboxKeys(ctx, tx, keys)
}

func expiredInboxKeys(ctx context.Context, tx pgx.Tx, query string, args ...any) ([]inboxMessageKey, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := []inboxMessageKey{}
	for rows.Next() {
		var key inboxMessageKey
		if err := rows.Scan(&key.provider, &key.mailboxEmail, &key.messageKey); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

func deleteInboxKeys(ctx context.Context, tx pgx.Tx, keys []inboxMessageKey) error {
	if len(keys) == 0 {
		return nil
	}
	providers := make([]string, 0, len(keys))
	mailboxEmails := make([]string, 0, len(keys))
	messageKeys := make([]string, 0, len(keys))
	for _, key := range keys {
		providers = append(providers, key.provider)
		mailboxEmails = append(mailboxEmails, key.mailboxEmail)
		messageKeys = append(messageKeys, key.messageKey)
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM mailbox_inbox_messages msg
		USING unnest($1::text[], $2::text[], $3::text[]) expired(provider, mailbox_email, message_key)
		WHERE msg.provider = expired.provider
		  AND msg.mailbox_email = expired.mailbox_email
		  AND msg.message_key = expired.message_key
	`, providers, mailboxEmails, messageKeys); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		DELETE FROM mailbox_inbox_seen seen
		USING unnest($1::text[], $2::text[], $3::text[]) expired(provider, mailbox_email, message_key)
		WHERE seen.provider = expired.provider
		  AND seen.mailbox_email = expired.mailbox_email
		  AND seen.message_key = expired.message_key
	`, providers, mailboxEmails, messageKeys)
	return err
}

func messageMailboxEmails(accountEmail string, recipients []string) []string {
	items := []string{emailx.Normalize(accountEmail)}
	for _, recipient := range recipients {
		if email := emailx.Normalize(recipient); email != "" {
			items = append(items, email)
		}
	}
	return uniqueStrings(items)
}

func stableMessageKey(provider string, mailboxEmail string, value string) string {
	return hashx.StableParts(normalizeEmailProvider(provider), emailx.Normalize(mailboxEmail), strings.TrimSpace(value))
}
