package inboxapp

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/stringx"

	"mailboxapi/pb"
)

func inboundEmailInputs(event *pb.InboundEmailWebhook, provider string, recipients []string, receivedAt int64) []MessageInput {
	body := inboundTextBody(event)
	inputs := make([]MessageInput, 0, len(recipients))
	for _, recipient := range recipients {
		key := StableMessageKey(provider, recipient, stringx.FirstNonEmpty(event.GetEventId(), event.GetMessageId(), event.GetSubject()))
		messageID := stringx.FirstNonEmpty(event.GetMessageId(), event.GetEventId(), key)
		inputs = append(inputs, MessageInput{Message: &mailboxv1.EmailInboxMessage{
			Id:                 messageID,
			MailboxEmail:       recipient,
			Subject:            strings.TrimSpace(event.GetSubject()),
			FromAddress:        emailx.Normalize(event.GetFromAddress()),
			BodyPreview:        CompactMessageText(body, 500),
			ReceivedAtUnix:     receivedAt,
			Recipients:         recipients,
			ProviderKey:        provider,
			SourceMailboxEmail: recipient,
			RawSize:            event.GetRawSize(),
		}, BodyText: body, HTMLBody: strings.TrimSpace(event.GetHtmlBody())})
	}
	return inputs
}

func inboundTextBody(event *pb.InboundEmailWebhook) string {
	body := strings.TrimSpace(event.GetTextBody())
	if body == "" {
		body = CompactMessageText(event.GetHtmlBody(), 5000)
	}
	return body
}
