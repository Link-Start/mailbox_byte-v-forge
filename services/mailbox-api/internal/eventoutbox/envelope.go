package eventoutbox

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
)

func MessageFromEnvelope(payload []byte) (eventbus.Message, error) {
	envelope := &commonv1.EventEnvelope{}
	if err := proto.Unmarshal(payload, envelope); err != nil {
		return eventbus.Message{}, fmt.Errorf("decode event outbox envelope: %w", err)
	}
	messageType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(envelope.GetPayloadType()))
	if err != nil {
		return eventbus.Message{}, fmt.Errorf("resolve event outbox payload type %s: %w", envelope.GetPayloadType(), err)
	}
	message := messageType.New().Interface()
	if err := proto.Unmarshal(envelope.GetPayload(), message); err != nil {
		return eventbus.Message{}, fmt.Errorf("decode event outbox payload %s: %w", envelope.GetPayloadType(), err)
	}
	return eventbus.Message{
		Subject:    envelope.GetSubject(),
		Event:      message,
		Metadata:   envelope.GetMetadata(),
		Extensions: envelope.GetExtensions(),
	}, nil
}
