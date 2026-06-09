package eventcatalog

import (
	"strings"

	"mailboxapi/internal/eventbus"
)

func (definition Definition) ExpectedMessage() eventbus.ExpectedMessage {
	return eventbus.ExpectedMessage{
		Subject:      strings.TrimSpace(definition.Subject),
		EventName:    strings.TrimSpace(definition.EventName),
		EventVersion: strings.TrimSpace(definition.EventVersion),
		PayloadType:  strings.TrimSpace(definition.PayloadType),
	}
}
