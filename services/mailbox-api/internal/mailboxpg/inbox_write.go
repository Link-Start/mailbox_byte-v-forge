package mailboxpg

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
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
