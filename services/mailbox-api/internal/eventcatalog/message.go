package eventcatalog

import (
	"strings"

	"google.golang.org/protobuf/proto"
	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
)

func (definition Definition) NewMessage(
	event proto.Message,
	metadata *commonv1.EventMetadata,
	attributes map[string]string,
) (eventbus.Message, error) {
	if err := definition.ValidateEvent(event); err != nil {
		return eventbus.Message{}, err
	}
	if err := definition.ValidateMetadata(metadata); err != nil {
		return eventbus.Message{}, err
	}
	return eventbus.Message{
		Subject:    strings.TrimSpace(definition.Subject),
		Event:      event,
		Metadata:   metadata,
		Extensions: attributes,
	}, nil
}
