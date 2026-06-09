package mailboxmem

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/stringx"
)

func (r *Repository) persistedInboxMessage(provider string, mailboxEmail string, sourceEmail string, key string, bodyText string, htmlBody string, message *mailboxv1.EmailInboxMessage, now int64) *mailboxv1.EmailInboxMessage {
	receivedAt := message.GetReceivedAtUnix()
	if receivedAt <= 0 {
		receivedAt = now
	}
	messageID := stringx.FirstNonEmpty(message.GetId(), key)
	return &mailboxv1.EmailInboxMessage{
		Id:                 messageID,
		MailboxEmail:       mailboxEmail,
		Subject:            strings.TrimSpace(message.GetSubject()),
		FromAddress:        emailx.Normalize(message.GetFromAddress()),
		BodyPreview:        inboxapp.CompactMessageText(message.GetBodyPreview(), 500),
		ReceivedAtUnix:     receivedAt,
		Recipients:         inboxapp.UniqueEmails(message.GetRecipients()),
		ProviderKey:        provider,
		SourceMailboxEmail: sourceEmail,
		BodyArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "body_text", int64(len(bodyText)), r.providers.NormalizeProviderInput),
		HtmlArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "html_body", int64(len(htmlBody)), r.providers.NormalizeProviderInput),
		RawSize:            message.GetRawSize(),
	}
}
