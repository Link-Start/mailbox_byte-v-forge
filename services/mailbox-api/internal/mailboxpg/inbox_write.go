package mailboxpg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/stringx"

	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxprovider"
)

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
	for _, input := range request.Messages {
		if input.Message == nil {
			continue
		}
		for _, mailboxEmail := range persistTargetMailboxes(input, request.ExpandRecipients) {
			touchedMailboxes[mailboxEmail] = struct{}{}
			if domain := DomainForEmail(mailboxEmail); domain != "" {
				touchedDomains[domain] = struct{}{}
			}
			persisted, key, err := persistInboxMessage(ctx, tx, provider, mailboxEmail, input, now)
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

func persistTargetMailboxes(input inboxapp.MessageInput, expandRecipients bool) []string {
	message := input.Message
	if message == nil {
		return []string{}
	}
	if !expandRecipients {
		return inboxapp.UniqueEmails([]string{message.GetMailboxEmail()})
	}
	return inboxapp.MessageMailboxEmails(message.GetMailboxEmail(), message.GetRecipients())
}

func persistInboxMessage(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, input inboxapp.MessageInput, now int64) (*mailboxv1.EmailInboxMessage, string, error) {
	message := input.Message
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if mailboxEmail == "" {
		return nil, "", fmt.Errorf("mailbox_email is required")
	}
	if message == nil {
		return nil, "", fmt.Errorf("message is required")
	}
	receivedAt := message.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = now
	}
	bodyText := strings.TrimSpace(input.BodyText)
	if bodyText == "" {
		bodyText = strings.TrimSpace(message.GetBodyPreview())
	}
	htmlBody := strings.TrimSpace(input.HTMLBody)
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
		BodyArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "body_text", int64(len(bodyText)), mailboxprovider.NormalizeKey),
		HtmlArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "html_body", int64(len(htmlBody)), mailboxprovider.NormalizeKey),
		RawSize:            message.GetRawSize(),
	}
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
		HTMLBody:       htmlBody,
		RawSize:        persisted.GetRawSize(),
	}, now); err != nil {
		return nil, "", err
	}
	return persisted, key, nil
}
