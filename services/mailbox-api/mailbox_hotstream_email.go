package main

import (
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
	"mailboxapi/internal/hotstream"
	"mailboxapi/internal/inboxapp"
)

func mailboxEmailReceivedEvent(message *mailboxv1.EmailInboxMessage) *observabilityv1.HotStreamEvent {
	return hotstream.NewEvent(hotstream.EventConfig{
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
	})
}

func mailboxSignalReceivedEvent(message *mailboxv1.EmailInboxMessage, signal *mailboxv1.EmailSignal) *observabilityv1.HotStreamEvent {
	return hotstream.NewEvent(hotstream.EventConfig{
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
	})
}

func emailOccurredAt(message *mailboxv1.EmailInboxMessage) time.Time {
	if message.GetReceivedAtUnix() > 0 {
		return time.Unix(message.GetReceivedAtUnix(), 0)
	}
	return time.Now()
}
