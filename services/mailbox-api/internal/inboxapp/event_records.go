package inboxapp

import (
	"google.golang.org/protobuf/proto"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/eventcatalog"
	"mailboxapi/internal/eventoutbox"
)

func EmailReceivedEventRecord(source string, message *mailboxv1.EmailInboxMessage) (eventoutbox.Record, error) {
	metadata := EventMetadata(source, eventcatalog.MailboxEmailReceived.EventName, eventcatalog.MailboxEmailReceived.Subject, EmailReceivedEventID(message), message)
	return eventoutbox.NewRecordFor(
		eventcatalog.MailboxEmailReceived,
		&mailboxv1.MailboxEmailReceivedEvent{
			Metadata: metadata,
			Message:  proto.Clone(message).(*mailboxv1.EmailInboxMessage),
		},
		metadata,
		EmailAttributes(message, nil),
	)
}

func EmailSignalReceivedEventRecord(source string, message *mailboxv1.EmailInboxMessage, signal *mailboxv1.EmailSignal) (eventoutbox.Record, error) {
	metadata := EventMetadata(source, eventcatalog.MailboxEmailSignalReceived.EventName, eventcatalog.MailboxEmailSignalReceived.Subject, EmailSignalEventID(message, signal), message)
	return eventoutbox.NewRecordFor(
		eventcatalog.MailboxEmailSignalReceived,
		&mailboxv1.MailboxEmailSignalReceivedEvent{
			Metadata: metadata,
			Message:  proto.Clone(message).(*mailboxv1.EmailInboxMessage),
			Signal:   proto.Clone(signal).(*mailboxv1.EmailSignal),
		},
		metadata,
		EmailAttributes(message, signal),
	)
}
