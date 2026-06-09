package eventcatalog

import commonv1 "mailboxapi/internal/contracts/commonv1"

const (
	StreamName      = "MAILBOX_EVENTS"
	StreamSubject   = "mailbox.>"
	DeadLetterTopic = "mailbox.dead_letter"
	EventVersionV1  = "v1"
)

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

func Subjects() []string {
	return []string{StreamSubject}
}
