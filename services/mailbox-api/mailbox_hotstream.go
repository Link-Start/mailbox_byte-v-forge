package main

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
	"mailboxapi/internal/hotstream"
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
		p.publish(ctx, mailboxEmailReceivedEvent(message))
		for _, signal := range message.GetSignals() {
			if signal == nil || signal.GetKind() == mailboxv1.EmailSignalKind_EMAIL_SIGNAL_KIND_UNSPECIFIED {
				continue
			}
			p.publish(ctx, mailboxSignalReceivedEvent(message, signal))
		}
	}
}

func (p *mailboxHotStream) PublishOperation(ctx context.Context, operation *mailboxv1.MailboxOperation) {
	if p == nil || p.publisher == nil || operation == nil {
		return
	}
	p.publish(ctx, mailboxOperationUpdatedEvent(operation))
}

func (p *mailboxHotStream) publish(ctx context.Context, event *observabilityv1.HotStreamEvent) {
	if err := p.publisher.Publish(context.WithoutCancel(ctx), event); err != nil {
		logWarning("publish mailbox hotstream event failed type=%s resource=%s: %v", event.GetMetadata().GetType(), event.GetResourceId(), err)
	}
}
