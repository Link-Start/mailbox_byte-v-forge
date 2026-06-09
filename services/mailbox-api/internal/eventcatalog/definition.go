package eventcatalog

import commonv1 "mailboxapi/internal/contracts/commonv1"

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
