package mailboxpg

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

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
