package mailboxmem

import (
	"context"
	"errors"
	"strings"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func (r *Repository) GetInboxRow(ctx context.Context, email string, messageID string, provider string) (inboxapp.MessageRow, bool, error) {
	if err := ctx.Err(); err != nil {
		return inboxapp.MessageRow{}, false, err
	}
	email = emailx.Normalize(email)
	if email == "" {
		return inboxapp.MessageRow{}, false, errors.New("email_address is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return inboxapp.MessageRow{}, false, errors.New("message_id is required")
	}
	provider = r.providers.NormalizeProviderInput(provider)
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, message := range r.messages {
		if message.row.MailboxEmail != email || message.row.ID != messageID {
			continue
		}
		if provider != "" && message.row.Provider != provider {
			continue
		}
		return message.row, true, nil
	}
	return inboxapp.MessageRow{}, false, nil
}
