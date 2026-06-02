package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/eventbus"
	"github.com/byte-v-forge/common-lib/eventcatalog"
	"github.com/byte-v-forge/common-lib/eventoutbox"
	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"google.golang.org/protobuf/proto"
)

const (
	mailboxPlatformEventSource  = "mailbox-api"
	mailboxPlatformEventVersion = eventcatalog.EventVersionV1
)

type mailboxPlatformEvents struct {
	publisher eventbus.Publisher
	source    string
}

func newMailboxPlatformEvents(publisher eventbus.Publisher) *mailboxPlatformEvents {
	if publisher == nil {
		return nil
	}
	return &mailboxPlatformEvents{publisher: publisher, source: mailboxPlatformEventSource}
}

func (p *mailboxPlatformEvents) Publish(ctx context.Context, message eventbus.Message) (eventbus.PublishAck, error) {
	if p == nil || p.publisher == nil {
		return eventbus.PublishAck{}, fmt.Errorf("mailbox platform event publisher is not configured")
	}
	return p.publisher.Publish(ctx, message)
}

func mailboxPlatformEventRecords(source string, messages []*mailboxv1.EmailInboxMessage) ([]eventoutbox.Record, error) {
	records := []eventoutbox.Record{}
	for _, message := range messages {
		if message == nil {
			continue
		}
		record, err := mailboxEmailReceivedEventRecord(source, message)
		if err != nil {
			return nil, fmt.Errorf("prepare mailbox platform event record: %w", err)
		}
		records = append(records, record)
	}
	return records, nil
}

func mailboxEmailReceivedEventRecord(source string, message *mailboxv1.EmailInboxMessage) (eventoutbox.Record, error) {
	eventCtx := mailboxPlatformEventContext(source, eventcatalog.MailboxEmailReceived.EventName, emailReceivedEventID(message), message)
	return eventoutbox.NewRecordFor(
		eventcatalog.MailboxEmailReceived,
		&mailboxv1.MailboxEmailReceivedEvent{
			Context: eventCtx,
			Message: proto.Clone(message).(*mailboxv1.EmailInboxMessage),
		},
		eventCtx,
		emailAttributes(message, nil),
	)
}

func mailboxPlatformEventContext(source string, eventName string, eventID string, message *mailboxv1.EmailInboxMessage) *commonv1.EventContext {
	occurredAt := time.Now()
	if message.GetReceivedAtUnix() > 0 {
		occurredAt = time.Unix(message.GetReceivedAtUnix(), 0)
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = mailboxPlatformEventSource
	}
	return eventbus.NewEventContext(eventbus.EventContextConfig{
		EventID:       eventID,
		EventName:     eventName,
		EventVersion:  mailboxPlatformEventVersion,
		OccurredAt:    occurredAt,
		SourceService: source,
		CorrelationID: message.GetMailboxEmail(),
	})
}

func emailReceivedEventID(message *mailboxv1.EmailInboxMessage) string {
	return eventbus.StableEventID("mailbox-email-",
		message.GetProviderKey(),
		message.GetMailboxEmail(),
		message.GetId(),
		fmt.Sprintf("%d", message.GetReceivedAtUnix()),
	)
}

func emailSignalEventID(message *mailboxv1.EmailInboxMessage, signal *mailboxv1.EmailSignal) string {
	return eventbus.StableEventID("mailbox-email-signal-",
		message.GetProviderKey(),
		message.GetMailboxEmail(),
		message.GetId(),
		signal.GetKind().String(),
		signal.GetProfile(),
		signal.GetParser(),
		signal.GetCode(),
		fmt.Sprintf("%d", message.GetReceivedAtUnix()),
	)
}

func emailAttributes(message *mailboxv1.EmailInboxMessage, signal *mailboxv1.EmailSignal) map[string]string {
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
