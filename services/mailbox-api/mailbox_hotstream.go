package main

import (
	"context"
	"fmt"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/hotstream"

	"mailboxapi/internal/inboxapp"
)

const (
	mailboxHotStreamSource       = "mailbox-api"
	mailboxResourceEmail         = "mailbox.email"
	mailboxResourceOperation     = "mailbox.operation"
	mailboxEventEmailReceived    = "mailbox.email.received"
	mailboxEventSignalReceived   = "mailbox.email.signal_received"
	mailboxEventOperationUpdated = "mailbox.operation.updated"
)

type mailboxHotStream struct {
	publisher hotstream.Publisher
}

func newMailboxHotStream(publisher hotstream.Publisher) *mailboxHotStream {
	if publisher == nil {
		return nil
	}
	return &mailboxHotStream{publisher: publisher}
}

func (p *mailboxHotStream) PublishEmailMessages(ctx context.Context, messages []*mailboxv1.EmailInboxMessage) {
	if p == nil || p.publisher == nil || len(messages) == 0 {
		return
	}
	for _, message := range messages {
		if message == nil {
			continue
		}
		p.publish(ctx, hotstream.NewEvent(hotstream.EventConfig{
			EventID:       inboxapp.EmailReceivedEventID(message),
			EventType:     mailboxEventEmailReceived,
			SourceService: mailboxHotStreamSource,
			ResourceType:  mailboxResourceEmail,
			ResourceID:    message.GetMailboxEmail(),
			Scope:         message.GetProviderKey(),
			OccurredAt:    emailOccurredAt(message),
			CorrelationID: message.GetMailboxEmail(),
			Attributes: map[string]string{
				"mailbox_email": message.GetMailboxEmail(),
				"provider_key":  message.GetProviderKey(),
				"message_id":    message.GetId(),
			},
		}))
		for _, signal := range message.GetSignals() {
			if signal == nil || signal.GetKind() == mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_UNSPECIFIED {
				continue
			}
			p.publish(ctx, hotstream.NewEvent(hotstream.EventConfig{
				EventID:       inboxapp.EmailSignalEventID(message, signal),
				EventType:     mailboxEventSignalReceived,
				SourceService: mailboxHotStreamSource,
				ResourceType:  mailboxResourceEmail,
				ResourceID:    message.GetMailboxEmail(),
				Scope:         message.GetProviderKey(),
				OccurredAt:    emailOccurredAt(message),
				CorrelationID: message.GetMailboxEmail(),
				Attributes: map[string]string{
					"mailbox_email":  message.GetMailboxEmail(),
					"provider_key":   message.GetProviderKey(),
					"message_id":     message.GetId(),
					"signal_kind":    signal.GetKind().String(),
					"signal_profile": signal.GetProfile(),
				},
			}))
		}
	}
}

func (p *mailboxHotStream) PublishOperation(ctx context.Context, operation *mailboxv1.MailboxOperation) {
	if p == nil || p.publisher == nil || operation == nil {
		return
	}
	action := operationActionValue(operation.GetAction())
	status := operationStatusValue(operation.GetStatus())
	p.publish(ctx, hotstream.NewEvent(hotstream.EventConfig{
		EventID:       eventbus.StableEventID("mailbox-operation-", operation.GetOperationId(), status, fmt.Sprintf("%d", operation.GetUpdatedAt())),
		EventType:     mailboxEventOperationUpdated,
		SourceService: mailboxHotStreamSource,
		ResourceType:  mailboxResourceOperation,
		ResourceID:    operation.GetOperationId(),
		Scope:         action,
		OccurredAt:    time.Unix(operation.GetUpdatedAt(), 0),
		CorrelationID: operation.GetOperationId(),
		Attributes: map[string]string{
			"operation_id":  operation.GetOperationId(),
			"action":        action,
			"status":        status,
			"email_address": operation.GetEmailAddress(),
		},
	}))
}

func (p *mailboxHotStream) publish(ctx context.Context, event *observabilityv1.HotStreamEvent) {
	if err := p.publisher.Publish(context.WithoutCancel(ctx), event); err != nil {
		logWarning("publish mailbox hotstream event failed type=%s resource=%s: %v", event.GetMetadata().GetType(), event.GetResourceId(), err)
	}
}

func emailOccurredAt(message *mailboxv1.EmailInboxMessage) time.Time {
	if message.GetReceivedAtUnix() > 0 {
		return time.Unix(message.GetReceivedAtUnix(), 0)
	}
	return time.Now()
}
