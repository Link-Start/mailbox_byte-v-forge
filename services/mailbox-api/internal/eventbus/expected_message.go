package eventbus

import (
	"strings"

	"google.golang.org/protobuf/proto"
)

type ExpectedMessage struct {
	Subject      string
	EventName    string
	EventVersion string
	PayloadType  string
}

func (expected ExpectedMessage) IsZero() bool {
	return strings.TrimSpace(expected.Subject) == "" &&
		strings.TrimSpace(expected.EventName) == "" &&
		strings.TrimSpace(expected.EventVersion) == "" &&
		strings.TrimSpace(expected.PayloadType) == ""
}

func (expected ExpectedMessage) ValidateReceived(received ReceivedMessage) error {
	if expected.IsZero() {
		return nil
	}
	if received.Envelope == nil {
		return ErrEmptyEnvelope
	}
	if err := expected.validateSubject(received); err != nil {
		return err
	}
	if err := expected.validateMetadata(received); err != nil {
		return err
	}
	return expected.validatePayloadType(received.Envelope.GetPayloadType())
}

func (expected ExpectedMessage) ValidateEvent(event proto.Message) error {
	return expected.validateEventPayloadType(event)
}
