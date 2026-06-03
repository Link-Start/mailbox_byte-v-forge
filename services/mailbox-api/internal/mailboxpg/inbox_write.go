package mailboxpg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/eventoutbox"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/inboxapp"
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

type inboxMessageKey struct {
	provider     string
	mailboxEmail string
	messageKey   string
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
	touchedMailboxes := map[string]struct{}{}
	touchedDomains := map[string]struct{}{}
	watermarks := map[string]int64{}
	unseen := []*mailboxv1.EmailInboxMessage{}
	for _, message := range request.Messages {
		for _, mailboxEmail := range persistTargetMailboxes(message, request.ExpandRecipients) {
			touchedMailboxes[mailboxEmail] = struct{}{}
			if domain := DomainForEmail(mailboxEmail); domain != "" {
				touchedDomains[domain] = struct{}{}
			}
			persisted, key, err := persistInboxMessage(ctx, tx, provider, mailboxEmail, message, now)
			if err != nil {
				return nil, err
			}
			trackInboxWatermark(watermarks, mailboxEmail, persisted.GetReceivedAtUnix())
			inserted, err := MarkInboxMessageSeen(ctx, tx, provider, mailboxEmail, key, now)
			if err != nil {
				return nil, err
			}
			if inserted {
				if request.PrepareUnseen != nil {
					if err := request.PrepareUnseen(ctx, persisted); err != nil {
						return nil, err
					}
				}
				unseen = append(unseen, persisted)
			}
		}
	}
	if err := UpdateInboxWatermarks(ctx, tx, watermarks, now); err != nil {
		return nil, err
	}
	if err := r.pruneInbound(ctx, tx, provider, mailboxprovider.InboxRetention{
		TouchedMailboxes: touchedMailboxes,
		TouchedDomains:   touchedDomains,
	}); err != nil {
		return nil, err
	}
	if err := enqueueInboxOutboxRecords(ctx, tx, request.OutboxTable, request.EventSource, unseen, now); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return unseen, nil
}

func persistTargetMailboxes(message *mailboxv1.EmailInboxMessage, expandRecipients bool) []string {
	if !expandRecipients {
		return inboxapp.UniqueEmails([]string{message.GetMailboxEmail()})
	}
	return inboxapp.MessageMailboxEmails(message.GetMailboxEmail(), message.GetRecipients())
}

func persistInboxMessage(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, message *mailboxv1.EmailInboxMessage, now int64) (*mailboxv1.EmailInboxMessage, string, error) {
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if mailboxEmail == "" {
		return nil, "", fmt.Errorf("mailbox_email is required")
	}
	receivedAt := message.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = now
	}
	sourceEmail := emailx.Normalize(stringx.FirstNonEmpty(message.GetSourceMailboxEmail(), message.GetMailboxEmail(), mailboxEmail))
	key := inboxapp.StableMessageKey(provider, mailboxEmail, stringx.FirstNonEmpty(message.GetId(), message.GetSubject(), message.GetBodyPreview()))
	messageID := stringx.FirstNonEmpty(message.GetId(), key)
	persisted := &mailboxv1.EmailInboxMessage{
		Id:                 messageID,
		MailboxEmail:       mailboxEmail,
		Subject:            strings.TrimSpace(message.GetSubject()),
		FromAddress:        emailx.Normalize(message.GetFromAddress()),
		BodyPreview:        inboxapp.CompactMessageText(message.GetBodyPreview(), 500),
		ReceivedAtUnix:     receivedAt,
		Recipients:         inboxapp.UniqueEmails(message.GetRecipients()),
		ProviderKey:        provider,
		SourceMailboxEmail: sourceEmail,
		BodyArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "body_text", int64(len(message.GetBodyPreview())), mailboxprovider.NormalizeKey),
		RawSize:            message.GetRawSize(),
	}
	bodyText := strings.TrimSpace(message.GetBodyPreview())
	if err := InsertInboxMessage(ctx, tx, PersistInboxMessage{
		Key:            key,
		ID:             messageID,
		MailboxEmail:   mailboxEmail,
		Subject:        persisted.GetSubject(),
		FromAddress:    persisted.GetFromAddress(),
		BodyPreview:    persisted.GetBodyPreview(),
		ReceivedAtUnix: receivedAt,
		Recipients:     persisted.GetRecipients(),
		Provider:       provider,
		SourceEmail:    sourceEmail,
		BodyText:       bodyText,
		HTMLBody:       "",
		RawSize:        persisted.GetRawSize(),
	}, now); err != nil {
		return nil, "", err
	}
	return persisted, key, nil
}

func trackInboxWatermark(watermarks map[string]int64, mailboxEmail string, receivedAtUnix int64) {
	if receivedAtUnix <= 0 {
		return
	}
	watermark := time.Unix(receivedAtUnix, 0).UnixNano()
	if watermarks[mailboxEmail] < watermark {
		watermarks[mailboxEmail] = watermark
	}
}

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

func UpdateInboxWatermarks(ctx context.Context, tx pgx.Tx, watermarks map[string]int64, now int64) error {
	for mailboxEmail, watermark := range watermarks {
		if watermark <= 0 {
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE mailboxes
			SET last_inbox_received_at_ns = GREATEST(last_inbox_received_at_ns, $1), updated_at = $2
			WHERE email = $3
		`, watermark, now, emailx.Normalize(mailboxEmail)); err != nil {
			return err
		}
	}
	return nil
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
