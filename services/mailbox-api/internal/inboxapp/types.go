package inboxapp

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

type MessageRow struct {
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

type MessageInput struct {
	Message  *mailboxv1.EmailInboxMessage
	BodyText string
	HTMLBody string
}
