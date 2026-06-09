package inboxapp

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func MessageInputs(messages []*mailboxv1.EmailInboxMessage) []MessageInput {
	out := make([]MessageInput, 0, len(messages))
	for _, message := range messages {
		if message == nil {
			continue
		}
		out = append(out, MessageInput{
			Message:  message,
			BodyText: strings.TrimSpace(message.GetBodyPreview()),
		})
	}
	return out
}
