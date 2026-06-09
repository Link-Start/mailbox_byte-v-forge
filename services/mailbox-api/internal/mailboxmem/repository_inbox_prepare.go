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
	bodyText, htmlBody := inboxBodies(input, message)
	sourceEmail := emailx.Normalize(stringx.FirstNonEmpty(message.GetSourceMailboxEmail(), message.GetMailboxEmail(), mailboxEmail))
	key := inboxapp.StableMessageKey(provider, mailboxEmail, stringx.FirstNonEmpty(message.GetId(), message.GetSubject(), message.GetBodyPreview()))
	persisted := r.persistedInboxMessage(provider, mailboxEmail, sourceEmail, key, bodyText, htmlBody, message, now)
	row, err := persistedInboxRow(provider, mailboxEmail, sourceEmail, bodyText, htmlBody, persisted)
	if err != nil {
		return nil, "", inboxapp.MessageRow{}, err
	}
	return persisted, key, row, nil
}

func inboxBodies(input inboxapp.MessageInput, message *mailboxv1.EmailInboxMessage) (string, string) {
	bodyText := strings.TrimSpace(input.BodyText)
	if bodyText == "" {
		bodyText = strings.TrimSpace(message.GetBodyPreview())
	}
	return bodyText, strings.TrimSpace(input.HTMLBody)
}

func persistedInboxRow(provider string, mailboxEmail string, sourceEmail string, bodyText string, htmlBody string, persisted *mailboxv1.EmailInboxMessage) (inboxapp.MessageRow, error) {
	recipientsJSON, err := json.Marshal(persisted.GetRecipients())
	if err != nil {
		return inboxapp.MessageRow{}, err
	}
	return inboxapp.MessageRow{
		ID:             persisted.GetId(),
		MailboxEmail:   mailboxEmail,
		Subject:        persisted.GetSubject(),
		FromAddress:    persisted.GetFromAddress(),
		BodyPreview:    persisted.GetBodyPreview(),
		ReceivedAtUnix: persisted.GetReceivedAtUnix(),
		RecipientsJSON: string(recipientsJSON),
		Provider:       provider,
		SourceEmail:    sourceEmail,
		BodyText:       bodyText,
		HTMLBody:       htmlBody,
		RawSize:        persisted.GetRawSize(),
	}, nil
}
