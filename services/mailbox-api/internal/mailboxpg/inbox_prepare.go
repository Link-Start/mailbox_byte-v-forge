package mailboxpg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/inboxapp"
)

func persistInboxMessage(ctx context.Context, tx pgx.Tx, provider string, mailboxEmail string, input inboxapp.MessageInput, now int64) (*mailboxv1.EmailInboxMessage, string, error) {
	message := input.Message
	mailboxEmail = emailx.Normalize(mailboxEmail)
	if mailboxEmail == "" {
		return nil, "", fmt.Errorf("mailbox_email is required")
	}
	if message == nil {
		return nil, "", fmt.Errorf("message is required")
	}
	bodyText, htmlBody := inboxMessageBodies(input, message)
	key := preparedInboxMessageKey(provider, mailboxEmail, message)
	persisted := persistedInboxMessage(provider, mailboxEmail, key, bodyText, htmlBody, message, now)
	if err := InsertInboxMessage(ctx, tx, inboxInsertInput(provider, mailboxEmail, key, bodyText, htmlBody, persisted), now); err != nil {
		return nil, "", err
	}
	return persisted, key, nil
}
