package main

import (
	"encoding/json"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
)

const inboxMessageSelectSQL = `
	SELECT message_id, mailbox_email, subject, from_address, body_preview,
		received_at, recipients_json, provider, source_mailbox_email, body_text,
		html_body, raw_size
	FROM mailbox_inbox_messages
`

func scanInboxMessageRow(scanner rowScanner) (inboxMessageRow, error) {
	var row inboxMessageRow
	err := scanner.Scan(
		&row.ID,
		&row.MailboxEmail,
		&row.Subject,
		&row.FromAddress,
		&row.BodyPreview,
		&row.ReceivedAtUnix,
		&row.RecipientsJSON,
		&row.Provider,
		&row.SourceEmail,
		&row.BodyText,
		&row.HTMLBody,
		&row.RawSize,
	)
	return row, err
}

func inboxMessageToProtoLenient(row inboxMessageRow) *mailboxv1.EmailInboxMessage {
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
		BodyText:           row.BodyText,
		HtmlBody:           row.HTMLBody,
		RawSize:            row.RawSize,
	}, "")
}
