package mailboxmem

import (
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/mailboxprovider"
)

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

func trackInboxRetention(retention *mailboxprovider.InboxRetention, mailboxEmail string) {
	retention.TouchedMailboxes[mailboxEmail] = struct{}{}
	if domain := domainForEmail(mailboxEmail); domain != "" {
		retention.TouchedDomains[domain] = struct{}{}
	}
}

func messageStorageKey(provider string, mailboxEmail string, key string) string {
	return strings.Join([]string{
		mailboxprovider.NormalizeKey(provider),
		emailx.Normalize(mailboxEmail),
		strings.TrimSpace(key),
	}, "\x00")
}
