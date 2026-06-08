package eventcatalog

import commonv1 "mailboxapi/internal/contracts/commonv1"

const (
	StreamName      = "BYTE_V_FORGE_EVENTS"
	StreamSubject   = "byte.v.forge.>"
	DeadLetterTopic = "byte.v.forge.platform.dead_letter"
	EventVersionV1  = "v1"
)

type Kind string

const (
	KindFact    Kind = "fact"
	KindCommand Kind = "command"
)

type Definition struct {
	Subject          string
	EventName        string
	EventVersion     string
	Kind             Kind
	PayloadType      string
	OwnerService     string
	ConsumerDurable  string
	Retryable        bool
	MaxDeliveries    int
	RetryDelaySecond int
}

func (definition Definition) Proto() *commonv1.EventDefinition {
	return &commonv1.EventDefinition{
		Subject:           definition.Subject,
		EventName:         definition.EventName,
		EventVersion:      definition.EventVersion,
		Kind:              protoKind(definition.Kind),
		PayloadType:       definition.PayloadType,
		OwnerService:      definition.OwnerService,
		ConsumerDurable:   definition.ConsumerDurable,
		Retryable:         definition.Retryable,
		MaxDeliveries:     int32(definition.MaxDeliveries),
		RetryDelaySeconds: int32(definition.RetryDelaySecond),
	}
}

func Catalog() *commonv1.EventCatalog {
	definitions := All()
	out := make([]*commonv1.EventDefinition, 0, len(definitions))
	for _, definition := range definitions {
		out = append(out, definition.Proto())
	}
	return &commonv1.EventCatalog{
		StreamName:    StreamName,
		StreamSubject: StreamSubject,
		Definitions:   out,
	}
}

func protoKind(kind Kind) commonv1.EventKind {
	switch kind {
	case KindFact:
		return commonv1.EventKind_EVENT_KIND_FACT
	case KindCommand:
		return commonv1.EventKind_EVENT_KIND_COMMAND
	default:
		return commonv1.EventKind_EVENT_KIND_UNSPECIFIED
	}
}

var (
	MailboxEmailPollRequested = Definition{
		Subject:          "byte.v.forge.mailbox.email.poll.requested",
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
		Subject:      "byte.v.forge.mailbox.email.received",
		EventName:    "mailbox.email.received",
		EventVersion: EventVersionV1,
		Kind:         KindFact,
		PayloadType:  "byte.v.forge.contracts.mailbox.v1.MailboxEmailReceivedEvent",
		OwnerService: "mailbox-api",
	}
	MailboxEmailSignalReceived = Definition{
		Subject:      "byte.v.forge.mailbox.email.signal.received",
		EventName:    "mailbox.email.signal.received",
		EventVersion: EventVersionV1,
		Kind:         KindFact,
		PayloadType:  "byte.v.forge.contracts.mailbox.v1.MailboxEmailSignalReceivedEvent",
		OwnerService: "mailbox-api",
	}
	DeadLetter = Definition{
		Subject:      DeadLetterTopic,
		EventName:    "platform.dead_letter",
		EventVersion: EventVersionV1,
		Kind:         KindFact,
		PayloadType:  "byte.v.forge.contracts.common.v1.DeadLetterEvent",
		OwnerService: "platform",
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

func Subjects() []string {
	return []string{StreamSubject}
}
