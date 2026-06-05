package mailboxpg

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/mailboxprovider"
)

type inboxMessageKey struct {
	provider     string
	mailboxEmail string
	messageKey   string
}

func pruneMailboxMessages(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, limit int) error {
	provider = mailboxprovider.NormalizeKey(provider)
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
	provider = mailboxprovider.NormalizeKey(provider)
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

func (r *Repository) pruneInbound(ctx context.Context, tx pgx.Tx, provider string, retention mailboxprovider.InboxRetention) error {
	definition := r.providers.RetentionByKey(provider)
	if definition == nil {
		return nil
	}
	policy, ok := definition.RetentionPolicy()
	if !ok {
		return nil
	}
	switch policy.Scope {
	case mailboxprovider.RetentionScopeDomain:
		for domain := range retention.TouchedDomains {
			if err := pruneDomainMessages(ctx, tx, provider, domain, policy.MaxMessages); err != nil {
				return err
			}
		}
	case mailboxprovider.RetentionScopeMailbox:
		for mailboxEmail := range retention.TouchedMailboxes {
			if err := pruneMailboxMessages(ctx, tx, provider, mailboxEmail, policy.MaxMessages); err != nil {
				return err
			}
		}
	}
	return nil
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
