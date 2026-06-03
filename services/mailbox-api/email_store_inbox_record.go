package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/byte-v-forge/common-lib/stringx"
	"github.com/jackc/pgx/v5"

	"mailboxapi/pb"
)

func (s *MailboxStore) InboxWatermark(ctx context.Context, email string) (int64, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return 0, errors.New("email_address is required")
	}
	var watermark int64
	err := s.pool.QueryRow(ctx, "SELECT last_inbox_received_at_ns FROM mailboxes WHERE email = $1", email).Scan(&watermark)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("mailbox not found: %s", emailx.Redact(email))
	}
	return watermark, err
}

func (s *MailboxStore) HasInboxMessages(ctx context.Context, email string) (bool, error) {
	email = emailx.Normalize(email)
	if email == "" {
		return false, errors.New("email_address is required")
	}
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM mailbox_inbox_messages WHERE mailbox_email = $1
		)
	`, email).Scan(&exists)
	return exists, err
}

func (s *MailboxStore) RecordInboundEmail(ctx context.Context, event *pb.InboundEmailWebhook) ([]*mailboxv1.EmailInboxMessage, error) {
	if event == nil {
		return nil, errors.New("email event is required")
	}
	provider := normalizeEmailProvider(event.GetProviderKey())
	if provider == "" {
		return nil, errors.New("email event provider is required")
	}
	recipients := uniqueStrings(event.GetRecipients())
	if len(recipients) == 0 {
		return nil, errors.New("email event recipients are required")
	}
	receivedAt := event.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = time.Now().Unix()
	}
	body := strings.TrimSpace(event.GetTextBody())
	if body == "" {
		body = compactMessageText(event.GetHtmlBody(), 5000)
	}

	messages := make([]*mailboxv1.EmailInboxMessage, 0, len(recipients))
	for _, recipient := range recipients {
		key := stableMessageKey(provider, recipient, stringx.FirstNonEmpty(event.GetEventId(), event.GetMessageId(), event.GetSubject()))
		messageID := stringx.FirstNonEmpty(event.GetMessageId(), event.GetEventId(), key)
		messages = append(messages, &mailboxv1.EmailInboxMessage{
			Id:                 messageID,
			MailboxEmail:       recipient,
			Subject:            strings.TrimSpace(event.GetSubject()),
			FromAddress:        emailx.Normalize(event.GetFromAddress()),
			BodyPreview:        compactMessageText(body, 500),
			ReceivedAtUnix:     receivedAt,
			Recipients:         recipients,
			ProviderKey:        provider,
			SourceMailboxEmail: recipient,
			RawSize:            event.GetRawSize(),
		})
	}
	return s.recordInboxMessages(ctx, provider, messages, false)
}

func (s *MailboxStore) RecordInboxMessages(ctx context.Context, sourceEmail string, messages []graphMessage) ([]*mailboxv1.EmailInboxMessage, error) {
	sourceEmail = emailx.Normalize(sourceEmail)
	if sourceEmail == "" {
		return nil, errors.New("email_address is required")
	}
	return s.recordInboxMessages(ctx, emailProviderOutlook, inboxMessages(sourceEmail, messages), true)
}

func (s *MailboxStore) recordInboxMessages(ctx context.Context, provider string, messages []*mailboxv1.EmailInboxMessage, expandRecipients bool) ([]*mailboxv1.EmailInboxMessage, error) {
	provider = normalizeEmailProvider(provider)
	if provider == "" {
		return nil, errors.New("email provider is required")
	}
	if len(messages) == 0 {
		return []*mailboxv1.EmailInboxMessage{}, nil
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	now := time.Now().Unix()
	touchedMailboxes := map[string]struct{}{}
	touchedDomains := map[string]struct{}{}
	watermarks := map[string]int64{}
	unseen := []*mailboxv1.EmailInboxMessage{}
	for _, message := range messages {
		for _, mailboxEmail := range persistTargetMailboxes(message, expandRecipients) {
			touchedMailboxes[mailboxEmail] = struct{}{}
			if domain := domainForEmail(mailboxEmail); domain != "" {
				touchedDomains[domain] = struct{}{}
			}
			persisted, key, err := persistInboxMessage(ctx, tx, provider, mailboxEmail, message, now)
			if err != nil {
				return nil, err
			}
			trackInboxWatermark(watermarks, mailboxEmail, persisted.GetReceivedAtUnix())
			inserted, err := markInboxMessageSeen(ctx, tx, provider, mailboxEmail, key, now)
			if err != nil {
				return nil, err
			}
			if inserted {
				persisted = emailMessageWithSignals(persisted, "")
				if err := s.attachEmailSignalSecrets(ctx, persisted); err != nil {
					return nil, err
				}
				unseen = append(unseen, persisted)
			}
		}
	}
	if err := updateInboxWatermarks(ctx, tx, watermarks, now); err != nil {
		return nil, err
	}
	if err := mailboxProviderPruneInbound(ctx, tx, provider, mailboxInboxRetention{
		touchedMailboxes: touchedMailboxes,
		touchedDomains:   touchedDomains,
	}); err != nil {
		return nil, err
	}
	if err := s.enqueueInboxOutboxEvents(ctx, tx, unseen); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	s.recordRecentInboxMessages(ctx, unseen)
	return unseen, nil
}

func persistTargetMailboxes(message *mailboxv1.EmailInboxMessage, expandRecipients bool) []string {
	if !expandRecipients {
		return uniqueStrings([]string{message.GetMailboxEmail()})
	}
	return messageMailboxEmails(message.GetMailboxEmail(), message.GetRecipients())
}

func persistInboxMessage(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, message *mailboxv1.EmailInboxMessage, now int64) (*mailboxv1.EmailInboxMessage, string, error) {
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if mailboxEmail == "" {
		return nil, "", errors.New("mailbox_email is required")
	}
	receivedAt := message.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = now
	}
	sourceEmail := emailx.Normalize(stringx.FirstNonEmpty(message.GetSourceMailboxEmail(), message.GetMailboxEmail(), mailboxEmail))
	key := stableMessageKey(provider, mailboxEmail, stringx.FirstNonEmpty(message.GetId(), message.GetSubject(), message.GetBodyPreview()))
	messageID := stringx.FirstNonEmpty(message.GetId(), key)
	persisted := &mailboxv1.EmailInboxMessage{
		Id:                 messageID,
		MailboxEmail:       mailboxEmail,
		Subject:            strings.TrimSpace(message.GetSubject()),
		FromAddress:        emailx.Normalize(message.GetFromAddress()),
		BodyPreview:        compactMessageText(message.GetBodyPreview(), 500),
		ReceivedAtUnix:     receivedAt,
		Recipients:         uniqueStrings(message.GetRecipients()),
		ProviderKey:        provider,
		SourceMailboxEmail: sourceEmail,
		BodyArtifactRef:    inboxArtifactRef(provider, mailboxEmail, messageID, "body_text", int64(len(message.GetBodyPreview()))),
		RawSize:            message.GetRawSize(),
	}
	bodyText := strings.TrimSpace(message.GetBodyPreview())
	if err := insertInboxMessage(ctx, tx, inboxPersistMessage{
		key:            key,
		id:             messageID,
		mailboxEmail:   mailboxEmail,
		subject:        persisted.GetSubject(),
		fromAddress:    persisted.GetFromAddress(),
		bodyPreview:    persisted.GetBodyPreview(),
		receivedAtUnix: receivedAt,
		recipients:     persisted.GetRecipients(),
		provider:       provider,
		sourceEmail:    sourceEmail,
		bodyText:       bodyText,
		htmlBody:       "",
		rawSize:        persisted.GetRawSize(),
	}, now); err != nil {
		return nil, "", err
	}
	return persisted, key, nil
}

func markInboxMessageSeen(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, key string, now int64) (bool, error) {
	tag, err := tx.Exec(ctx, `
		INSERT INTO mailbox_inbox_seen (provider, mailbox_email, message_key, seen_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (provider, mailbox_email, message_key) DO NOTHING
	`, provider, mailboxEmail, key, now)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
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

func updateInboxWatermarks(ctx context.Context, tx pgx.Tx, watermarks map[string]int64, now int64) error {
	for mailboxEmail, watermark := range watermarks {
		if watermark <= 0 {
			continue
		}
		if _, err := tx.Exec(ctx, `
			UPDATE mailboxes
			SET last_inbox_received_at_ns = GREATEST(last_inbox_received_at_ns, $1), updated_at = $2
			WHERE email = $3
		`, watermark, now, mailboxEmail); err != nil {
			return err
		}
	}
	return nil
}
