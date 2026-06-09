package inboxapp

import (
	"strings"
	"time"

	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/eventbus"
)

func EventMetadata(source string, eventName string, subject string, eventID string, message *mailboxv1.EmailInboxMessage) *commonv1.EventMetadata {
	occurredAt := time.Now()
	if message.GetReceivedAtUnix() > 0 {
		occurredAt = time.Unix(message.GetReceivedAtUnix(), 0)
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = "mailbox-api"
	}
	return eventbus.NewEventMetadata(eventbus.EventMetadataConfig{
		EventID:       eventID,
		EventName:     eventName,
		EventVersion:  EventVersion,
		OccurredAt:    occurredAt,
		SourceService: source,
		Subject:       subject,
		CorrelationID: message.GetMailboxEmail(),
	})
}
