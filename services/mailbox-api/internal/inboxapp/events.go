package inboxapp

import (
	"fmt"
	"strings"
	"time"

	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"google.golang.org/protobuf/proto"
	"mailboxapi/internal/eventbus"
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
			return nil, fmt.Errorf("prepare mailbox platform event record: %w", err)
		}
		records = append(records, record)
		for _, signal := range message.GetSignals() {
			if signal == nil || signal.GetKind() == mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_UNSPECIFIED {
				continue
			}
			record, err := EmailSignalReceivedEventRecord(source, message, signal)
			if err != nil {
				return nil, fmt.Errorf("prepare mailbox platform signal event record: %w", err)
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

func EventMetadata(source string, eventName string, subject string, eventID string, message *mailboxv1.EmailInboxMessage) *commonv1.EventMetadata {
	occurredAt := time.Now()
	if message.GetReceivedAtUnix() > 0 {
		occurredAt = time.Unix(message.GetReceivedAtUnix(), 0)
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = "mailbox-api"
	}
	return eventbus.NewEventMetadata(eventbus.EventMetadataConfig{
		EventID:       eventID,
		EventName:     eventName,
		EventVersion:  EventVersion,
		OccurredAt:    occurredAt,
		SourceService: source,
		Subject:       subject,
		CorrelationID: message.GetMailboxEmail(),
	})
}

func EmailReceivedEventID(message *mailboxv1.EmailInboxMessage) string {
	return eventbus.StableEventID("mailbox-email-",
		message.GetProviderKey(),
		message.GetMailboxEmail(),
		message.GetId(),
		fmt.Sprintf("%d", message.GetReceivedAtUnix()),
	)
}

func EmailSignalEventID(message *mailboxv1.EmailInboxMessage, signal *mailboxv1.EmailSignal) string {
	return eventbus.StableEventID("mailbox-email-signal-",
		message.GetProviderKey(),
		message.GetMailboxEmail(),
		message.GetId(),
		signal.GetKind().String(),
		signal.GetProfile(),
		signal.GetParser(),
		signal.GetSecretRef().GetSecretId(),
		fmt.Sprintf("%d", message.GetReceivedAtUnix()),
	)
}

func EmailAttributes(message *mailboxv1.EmailInboxMessage, signal *mailboxv1.EmailSignal) map[string]string {
	attrs := eventbus.Attributes(
		"mailbox_email", message.GetMailboxEmail(),
		"provider_key", message.GetProviderKey(),
		"message_id", message.GetId(),
	)
	if signal != nil {
		attrs = eventbus.WithAttribute(attrs, "signal_kind", signal.GetKind().String())
		attrs = eventbus.WithAttribute(attrs, "signal_profile", signal.GetProfile())
	}
	return attrs
}
