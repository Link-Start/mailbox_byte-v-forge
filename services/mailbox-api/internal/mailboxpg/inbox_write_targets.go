package mailboxpg

import (
	"mailboxapi/internal/inboxapp"
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

func trackInboxRetention(write *inboxWriteResult, mailboxEmail string) {
	write.retention.TouchedMailboxes[mailboxEmail] = struct{}{}
	if domain := DomainForEmail(mailboxEmail); domain != "" {
		write.retention.TouchedDomains[domain] = struct{}{}
	}
}
