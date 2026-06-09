package natseventbus

import (
	"context"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventcatalog"
)

func publishDeadLetter(ctx context.Context, bus *Bus, durable string, envelope *commonv1.EventEnvelope, attempt int32, reason string) error {
	if bus == nil || envelope == nil {
		return nil
	}
	original := deadLetterOriginal(envelope)
	metadata := deadLetterMetadata(envelope, durable, attempt, original)
	message, err := eventcatalog.DeadLetter.NewMessage(deadLetterEvent(envelope, durable, attempt, reason, metadata, original), metadata, deadLetterAttributes(envelope, durable, attempt, original))
	if err != nil {
		return err
	}
	_, err = bus.Publish(ctx, message)
	return err
}
