package inboxapp

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

type RecordMessagesRequest struct {
	Provider         string
	Messages         []MessageInput
	ExpandRecipients bool
	OutboxTable      string
	EventSource      string
	PrepareUnseen    func(context.Context, *mailboxv1.EmailInboxMessage) error
}
