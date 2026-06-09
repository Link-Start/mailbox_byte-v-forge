package mailboxpg

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/stringx"
)

func inboxMessageBodies(input inboxapp.MessageInput, message *mailboxv1.EmailInboxMessage) (string, string) {
	bodyText := strings.TrimSpace(input.BodyText)
	if bodyText == "" {
		bodyText = strings.TrimSpace(message.GetBodyPreview())
	}
	return bodyText, strings.TrimSpace(input.HTMLBody)
}

func preparedInboxMessageKey(provider string, mailboxEmail string, message *mailboxv1.EmailInboxMessage) string {
	return inboxapp.StableMessageKey(provider, mailboxEmail, stringx.FirstNonEmpty(message.GetId(), message.GetSubject(), message.GetBodyPreview()))
}

func inboxInsertInput(provider string, mailboxEmail string, key string, bodyText string, htmlBody string, persisted *mailboxv1.EmailInboxMessage) PersistInboxMessage {
	return PersistInboxMessage{
		Key:            key,
		ID:             persisted.GetId(),
		MailboxEmail:   mailboxEmail,
		Subject:        persisted.GetSubject(),
		FromAddress:    persisted.GetFromAddress(),
		BodyPreview:    persisted.GetBodyPreview(),
		ReceivedAtUnix: persisted.GetReceivedAtUnix(),
		Recipients:     persisted.GetRecipients(),
		Provider:       provider,
		SourceEmail:    persisted.GetSourceMailboxEmail(),
		BodyText:       bodyText,
		HTMLBody:       htmlBody,
		RawSize:        persisted.GetRawSize(),
	}
}
