package natseventbus

import (
	"fmt"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
)

func deadLetterMetadata(envelope *commonv1.EventEnvelope, durable string, attempt int32, original deadLetterSource) *commonv1.EventMetadata {
	eventID := eventbus.StableEventID("dead-letter-", envelope.GetSubject(), original.id, durable, fmt.Sprintf("%d", attempt))
	return eventbus.NewEventMetadata(eventbus.EventMetadataConfig{
		EventID:       eventID,
		EventName:     "mailbox.dead_letter",
		EventVersion:  eventcatalog.EventVersionV1,
		SourceService: "mailbox-eventbus",
		Subject:       eventcatalog.DeadLetter.Subject,
		CorrelationID: original.correlationID,
		TraceID:       original.traceID,
	})
}

func deadLetterEvent(envelope *commonv1.EventEnvelope, durable string, attempt int32, reason string, metadata *commonv1.EventMetadata, original deadLetterSource) *commonv1.DeadLetterEvent {
	return &commonv1.DeadLetterEvent{
		Metadata:             metadata,
		OriginalSubject:      envelope.GetSubject(),
		OriginalEventId:      original.id,
		OriginalEventType:    original.name,
		OriginalEventVersion: original.version,
		OriginalSource:       original.source,
		ConsumerDurable:      durable,
		DeliveryAttempt:      attempt,
		ErrorCode:            "terminated",
		ErrorMessage:         reason,
		CorrelationId:        original.correlationID,
	}
}

func deadLetterAttributes(envelope *commonv1.EventEnvelope, durable string, attempt int32, original deadLetterSource) map[string]string {
	return eventbus.Attributes(
		"original_subject", envelope.GetSubject(),
		"original_event_id", original.id,
		"consumer_durable", durable,
		"delivery_attempt", fmt.Sprintf("%d", attempt),
	)
}
