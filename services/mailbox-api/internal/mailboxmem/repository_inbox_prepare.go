package mailboxmem

import (
	"encoding/json"
	"errors"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
	"mailboxapi/internal/stringx"
)

func (r *Repository) prepareInboxMessage(provider string, mailboxEmail string, input inboxapp.MessageInput, now int64) (*mailboxv1.EmailInboxMessage, string, inboxapp.MessageRow, error) {
	message := input.Message
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if mailboxEmail == "" {
		return nil, "", inboxapp.MessageRow{}, errors.New("mailbox_email is required")
	}
	if message == nil {
		return nil, "", inboxapp.MessageRow{}, errors.New("message is required")
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
	bodyPreview := inboxapp.CompactMessageText(message.GetBodyPreview(), 500)
	recipients := inboxapp.UniqueEmails(message.GetRecipients())
	recipientsJSON, err := json.Marshal(recipients)
	if err != nil {
		return nil, "", inboxapp.MessageRow{}, err
	}
	persisted := &mailboxv1.EmailInboxMessage{
		Id:                 messageID,
		MailboxEmail:       mailboxEmail,
		Subject:            strings.TrimSpace(message.GetSubject()),
		FromAddress:        emailx.Normalize(message.GetFromAddress()),
		BodyPreview:        bodyPreview,
		ReceivedAtUnix:     receivedAt,
		Recipients:         recipients,
		ProviderKey:        provider,
		SourceMailboxEmail: sourceEmail,
		BodyArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "body_text", int64(len(bodyText)), r.providers.NormalizeProviderInput),
		HtmlArtifactRef:    inboxapp.ArtifactRef(provider, mailboxEmail, messageID, "html_body", int64(len(htmlBody)), r.providers.NormalizeProviderInput),
		RawSize:            message.GetRawSize(),
	}
	row := inboxapp.MessageRow{
		ID:             persisted.GetId(),
		MailboxEmail:   mailboxEmail,
		Subject:        persisted.GetSubject(),
		FromAddress:    persisted.GetFromAddress(),
		BodyPreview:    persisted.GetBodyPreview(),
		ReceivedAtUnix: receivedAt,
		RecipientsJSON: string(recipientsJSON),
		Provider:       provider,
		SourceEmail:    sourceEmail,
		BodyText:       bodyText,
		HTMLBody:       htmlBody,
		RawSize:        persisted.GetRawSize(),
	}
	return persisted, key, row, nil
}
