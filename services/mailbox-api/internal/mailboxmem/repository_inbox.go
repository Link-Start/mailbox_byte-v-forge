package mailboxmem

import (
	"context"
	"errors"
	"strings"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxprovider"
)

func (r *Repository) RecordMessages(ctx context.Context, request inboxapp.RecordMessagesRequest) ([]*mailboxv1.EmailInboxMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	provider := r.providers.NormalizeProviderInput(request.Provider)
	if provider == "" {
		return nil, errors.New("email provider is required")
	}
	if len(request.Messages) == 0 {
		return []*mailboxv1.EmailInboxMessage{}, nil
	}
	now := time.Now().Unix()
	unseen := []*mailboxv1.EmailInboxMessage{}
	touchedMailboxes := map[string]struct{}{}
	touchedDomains := map[string]struct{}{}

	r.mu.Lock()
	for _, input := range request.Messages {
		if input.Message == nil {
			continue
		}
		for _, mailboxEmail := range persistTargetMailboxes(input, request.ExpandRecipients) {
			touchedMailboxes[mailboxEmail] = struct{}{}
			if domain := domainForEmail(mailboxEmail); domain != "" {
				touchedDomains[domain] = struct{}{}
			}
			persisted, key, row, err := r.prepareInboxMessage(provider, mailboxEmail, input, now)
			if err != nil {
				r.mu.Unlock()
				return nil, err
			}
			storageKey := messageStorageKey(provider, mailboxEmail, key)
			_, existed := r.messages[storageKey]
			r.messages[storageKey] = storedMessage{key: key, createdAt: now, updatedAt: now, row: row}
			r.trackInboxWatermarkLocked(mailboxEmail, persisted.GetReceivedAtUnix(), now)
			if !existed {
				if request.PrepareUnseen != nil {
					if err := request.PrepareUnseen(ctx, persisted); err != nil {
						r.mu.Unlock()
						return nil, err
					}
				}
				unseen = append(unseen, persisted)
			}
		}
	}
	r.pruneInboundLocked(provider, mailboxprovider.InboxRetention{
		TouchedMailboxes: touchedMailboxes,
		TouchedDomains:   touchedDomains,
	})
	r.mu.Unlock()
	return unseen, nil
}

func (r *Repository) trackInboxWatermarkLocked(mailboxEmail string, receivedAtUnix int64, now int64) {
	if receivedAtUnix <= 0 {
		return
	}
	entry, ok := r.mailboxes[emailx.Normalize(mailboxEmail)]
	if !ok || entry.record == nil {
		return
	}
	watermark := time.Unix(receivedAtUnix, 0).UnixNano()
	if watermark > entry.inboxWatermark {
		entry.inboxWatermark = watermark
		entry.record.UpdatedAt = now
		r.mailboxes[emailx.Normalize(mailboxEmail)] = entry
	}
}

func (r *Repository) pruneInboundLocked(provider string, retention mailboxprovider.InboxRetention) {
	definition := r.providers.RetentionByKey(provider)
	if definition == nil {
		return
	}
	policy, ok := definition.RetentionPolicy()
	if !ok {
		return
	}
	switch policy.Scope {
	case mailboxprovider.RetentionScopeDomain:
		for domain := range retention.TouchedDomains {
			r.pruneMessagesLocked(func(message storedMessage) bool {
				return message.row.Provider == provider && domainForEmail(message.row.MailboxEmail) == domain
			}, policy.MaxMessages)
		}
	case mailboxprovider.RetentionScopeMailbox:
		for mailboxEmail := range retention.TouchedMailboxes {
			r.pruneMessagesLocked(func(message storedMessage) bool {
				return message.row.Provider == provider && message.row.MailboxEmail == mailboxEmail
			}, policy.MaxMessages)
		}
	}
}

func (r *Repository) pruneMessagesLocked(match func(storedMessage) bool, keep int) {
	if keep <= 0 {
		return
	}
	matches := []storedMessage{}
	for _, message := range r.messages {
		if match(message) {
			matches = append(matches, message)
		}
	}
	sortStoredMessages(matches)
	for index := keep; index < len(matches); index++ {
		delete(r.messages, messageStorageKey(matches[index].row.Provider, matches[index].row.MailboxEmail, matches[index].key))
	}
}

func (r *Repository) deleteInboxLocked(email string) bool {
	deleted := false
	for key, message := range r.messages {
		if message.row.MailboxEmail == email {
			delete(r.messages, key)
			deleted = true
		}
	}
	return deleted
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

func messageStorageKey(provider string, mailboxEmail string, key string) string {
	return strings.Join([]string{
		mailboxprovider.NormalizeKey(provider),
		emailx.Normalize(mailboxEmail),
		strings.TrimSpace(key),
	}, "\x00")
}
