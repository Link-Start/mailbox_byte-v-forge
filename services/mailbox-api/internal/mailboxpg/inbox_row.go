package mailboxpg

import "mailboxapi/internal/inboxapp"

const inboxMessageSelectSQL = `
	SELECT message_id, mailbox_email, subject, from_address, body_preview,
		received_at, recipients_json, provider, source_mailbox_email, body_text,
		html_body, raw_size
	FROM mailbox_inbox_messages
`

func scanInboxMessageRow(scanner Scanner) (inboxapp.MessageRow, error) {
	var row inboxapp.MessageRow
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
