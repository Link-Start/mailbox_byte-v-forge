package mailboxmem

import (
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxprovider"
)

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
