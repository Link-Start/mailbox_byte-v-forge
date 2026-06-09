package eventcatalog

var (
	MailboxEmailPollRequested = Definition{
		Subject:          "mailbox.email.poll.requested",
		EventName:        "mailbox.email.poll_requested",
		EventVersion:     EventVersionV1,
		Kind:             KindCommand,
		PayloadType:      "byte.v.forge.contracts.mailbox.v1.MailboxEmailPollRequest",
		OwnerService:     "mailbox-api",
		ConsumerDurable:  "mailbox-email-poll",
		Retryable:        true,
		MaxDeliveries:    20,
		RetryDelaySecond: 5,
	}

	MailboxEmailReceived = Definition{
		Subject:      "mailbox.email.received",
		EventName:    "mailbox.email.received",
		EventVersion: EventVersionV1,
		Kind:         KindFact,
		PayloadType:  "byte.v.forge.contracts.mailbox.v1.MailboxEmailReceivedEvent",
		OwnerService: "mailbox-api",
	}
	MailboxEmailSignalReceived = Definition{
		Subject:      "mailbox.email.signal.received",
		EventName:    "mailbox.email.signal.received",
		EventVersion: EventVersionV1,
		Kind:         KindFact,
		PayloadType:  "byte.v.forge.contracts.mailbox.v1.MailboxEmailSignalReceivedEvent",
		OwnerService: "mailbox-api",
	}
	DeadLetter = Definition{
		Subject:      DeadLetterTopic,
		EventName:    "mailbox.dead_letter",
		EventVersion: EventVersionV1,
		Kind:         KindFact,
		PayloadType:  "byte.v.forge.contracts.common.v1.DeadLetterEvent",
		OwnerService: "mailbox-api",
	}
)

func All() []Definition {
	return []Definition{
		MailboxEmailPollRequested,
		MailboxEmailReceived,
		MailboxEmailSignalReceived,
		DeadLetter,
	}
}
