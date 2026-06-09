package eventcatalog

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	commonv1 "mailboxapi/internal/contracts/commonv1"
	"mailboxapi/internal/eventbus"
)

func (definition Definition) ValidateEvent(event proto.Message) error {
	if strings.TrimSpace(definition.Subject) == "" {
		return ErrEmptyDefinitionSubject
	}
	if strings.TrimSpace(definition.PayloadType) == "" {
		return ErrEmptyDefinitionPayloadType
	}
	if event == nil {
		return eventbus.ErrEmptyEvent
	}
	actualType := string(event.ProtoReflect().Descriptor().FullName())
	if actualType != strings.TrimSpace(definition.PayloadType) {
		return fmt.Errorf("%w: expected %s, got %s", ErrMismatchedPayloadType, definition.PayloadType, actualType)
	}
	return nil
}

func (definition Definition) ValidateMetadata(metadata *commonv1.EventMetadata) error {
	if strings.TrimSpace(definition.EventName) == "" {
		return ErrEmptyDefinitionEventName
	}
	if strings.TrimSpace(definition.EventVersion) == "" {
		return ErrEmptyDefinitionEventVersion
	}
	if err := eventbus.ValidateMetadata(metadata); err != nil {
		return err
	}
	if strings.TrimSpace(metadata.GetType()) != strings.TrimSpace(definition.EventName) {
		return fmt.Errorf("%w: expected %s, got %s", ErrMismatchedEventName, definition.EventName, metadata.GetType())
	}
	if strings.TrimSpace(metadata.GetVersion()) != strings.TrimSpace(definition.EventVersion) {
		return fmt.Errorf("%w: expected %s, got %s", ErrMismatchedEventVersion, definition.EventVersion, metadata.GetVersion())
	}
	return nil
}
