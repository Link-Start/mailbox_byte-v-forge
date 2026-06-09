package eventoutbox

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
)

func NewRecord(message eventbus.Message) (Record, error) {
	envelope, err := eventbus.NewEnvelope(message)
	if err != nil {
		return Record{}, err
	}
	metadata := envelope.GetMetadata()
	if metadata == nil || strings.TrimSpace(metadata.GetId()) == "" {
		return Record{}, ErrMissingEventID
	}
	payload, err := proto.Marshal(envelope)
	if err != nil {
		return Record{}, fmt.Errorf("marshal event outbox envelope: %w", err)
	}
	return Record{
		EventID:        metadata.GetId(),
		Subject:        envelope.GetSubject(),
		EventName:      metadata.GetType(),
		IdempotencyKey: metadata.GetIdempotencyKey(),
		Envelope:       payload,
	}, nil
}

func NewRecordFor(
	definition eventcatalog.Definition,
	event proto.Message,
	metadata *commonv1.EventMetadata,
	attributes map[string]string,
) (Record, error) {
	message, err := definition.NewMessage(event, metadata, attributes)
	if err != nil {
		return Record{}, err
	}
	return NewRecord(message)
}
