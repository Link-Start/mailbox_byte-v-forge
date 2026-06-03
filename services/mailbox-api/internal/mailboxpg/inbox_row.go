package mailboxpg

const inboxMessageSelectSQL = `
	SELECT message_id, mailbox_email, subject, from_address, body_preview,
		received_at, recipients_json, provider, source_mailbox_email, body_text,
		html_body, raw_size
	FROM mailbox_inbox_messages
`

type InboxMessageRow struct {
	ID             string
	MailboxEmail   string
	Subject        string
	FromAddress    string
	BodyPreview    string
	ReceivedAtUnix int64
	RecipientsJSON string
	Provider       string
	SourceEmail    string
	BodyText       string
	HTMLBody       string
	RawSize        int64
}

func scanInboxMessageRow(scanner Scanner) (InboxMessageRow, error) {
	var row InboxMessageRow
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
