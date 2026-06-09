package mailboxmem

import (
	"encoding/json"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/inboxapp"
)

func persistedInboxRow(provider string, mailboxEmail string, sourceEmail string, key string, bodyText string, htmlBody string, persisted *mailboxv1.EmailInboxMessage) (inboxapp.MessageRow, error) {
	recipientsJSON, err := json.Marshal(persisted.GetRecipients())
	if err != nil {
		return inboxapp.MessageRow{}, err
	}
	return inboxapp.MessageRow{
		Key:            key,
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
