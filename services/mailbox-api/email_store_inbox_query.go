package main

import (
	"encoding/json"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxpg"
)

func inboxMessageToProtoLenient(row mailboxpg.InboxMessageRow) *mailboxv1.EmailInboxMessage {
	recipients := []string{}
	if err := json.Unmarshal([]byte(row.RecipientsJSON), &recipients); err != nil {
		recipients = []string{}
	}
	return emailMessageWithSignals(&mailboxv1.EmailInboxMessage{
		Id:                 row.ID,
		MailboxEmail:       emailx.Normalize(row.MailboxEmail),
		Subject:            row.Subject,
		FromAddress:        row.FromAddress,
		BodyPreview:        row.BodyPreview,
		ReceivedAtUnix:     row.ReceivedAtUnix,
		Recipients:         uniqueStrings(recipients),
		ProviderKey:        normalizeEmailProvider(row.Provider),
		SourceMailboxEmail: emailx.Normalize(row.SourceEmail),
		BodyArtifactRef:    inboxArtifactRef(row.Provider, row.MailboxEmail, row.ID, "body_text", int64(len(row.BodyText))),
		HtmlArtifactRef:    inboxArtifactRef(row.Provider, row.MailboxEmail, row.ID, "html_body", int64(len(row.HTMLBody))),
		RawSize:            row.RawSize,
	}, "")
}
