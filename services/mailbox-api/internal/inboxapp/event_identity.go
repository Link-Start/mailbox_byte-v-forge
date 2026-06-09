package inboxapp

import (
	"fmt"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/eventbus"
)

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
