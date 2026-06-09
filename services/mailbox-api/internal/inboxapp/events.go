package inboxapp

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/eventcatalog"
	"mailboxapi/internal/eventoutbox"
)

const EventVersion = eventcatalog.EventVersionV1

func EventRecords(source string, messages []*mailboxv1.EmailInboxMessage) ([]eventoutbox.Record, error) {
	records := []eventoutbox.Record{}
	for _, message := range messages {
		if message == nil {
			continue
		}
		record, err := EmailReceivedEventRecord(source, message)
		if err != nil {
			return nil, fmt.Errorf("prepare mailbox event record: %w", err)
		}
		records = append(records, record)
		for _, signal := range message.GetSignals() {
			if signal == nil || signal.GetKind() == mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_UNSPECIFIED {
				continue
			}
			record, err := EmailSignalReceivedEventRecord(source, message, signal)
			if err != nil {
				return nil, fmt.Errorf("prepare mailbox signal event record: %w", err)
			}
			records = append(records, record)
		}
	}
	return records, nil
}

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
