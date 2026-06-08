package inboxapp

import (
	"encoding/json"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func MessageFromRow(row MessageRow, profile string, normalizeProvider func(string) string) (*mailboxv1.EmailInboxMessage, error) {
	recipients := []string{}
	if strings.TrimSpace(row.RecipientsJSON) != "" {
		if err := json.Unmarshal([]byte(row.RecipientsJSON), &recipients); err != nil {
			return nil, err
		}
	}
	return messageFromRow(row, profile, recipients, normalizeProvider), nil
}

func MessageFromRowLenient(row MessageRow, normalizeProvider func(string) string) *mailboxv1.EmailInboxMessage {
	recipients := []string{}
	if err := json.Unmarshal([]byte(row.RecipientsJSON), &recipients); err != nil {
		recipients = []string{}
	}
	return messageFromRow(row, "", recipients, normalizeProvider)
}

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

func messageFromRow(row MessageRow, profile string, recipients []string, normalizeProvider func(string) string) *mailboxv1.EmailInboxMessage {
	if normalizeProvider == nil {
		normalizeProvider = NormalizeProviderKey
	}
	return MessageWithSignals(&mailboxv1.EmailInboxMessage{
		Id:                 row.ID,
		MailboxEmail:       emailx.Normalize(row.MailboxEmail),
		Subject:            row.Subject,
		FromAddress:        row.FromAddress,
		BodyPreview:        row.BodyPreview,
		ReceivedAtUnix:     row.ReceivedAtUnix,
		Recipients:         UniqueEmails(recipients),
		ProviderKey:        normalizeProvider(row.Provider),
		SourceMailboxEmail: emailx.Normalize(row.SourceEmail),
		BodyArtifactRef:    ArtifactRef(row.Provider, row.MailboxEmail, row.ID, "body_text", int64(len(row.BodyText)), normalizeProvider),
		HtmlArtifactRef:    ArtifactRef(row.Provider, row.MailboxEmail, row.ID, "html_body", int64(len(row.HTMLBody)), normalizeProvider),
		RawSize:            row.RawSize,
	}, profile)
}
