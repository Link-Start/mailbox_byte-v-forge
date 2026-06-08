package main

import (
	"fmt"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	observabilityv1 "mailboxapi/internal/contracts/observabilityv1"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/hotstream"
)

func mailboxOperationUpdatedEvent(operation *mailboxv1.MailboxOperation) *observabilityv1.HotStreamEvent {
	action := operationActionValue(operation.GetAction())
	status := operationStatusValue(operation.GetStatus())
	return hotstream.NewEvent(hotstream.EventConfig{
		EventID:       eventbus.StableEventID("mailbox-operation-", operation.GetOperationId(), status, fmt.Sprintf("%d", operation.GetUpdatedAt())),
		EventType:     mailboxEventOperationUpdated,
		SourceService: mailboxHotStreamSource,
		ResourceType:  mailboxResourceOperation,
		ResourceID:    operation.GetOperationId(),
		Scope:         action,
		OccurredAt:    time.Unix(operation.GetUpdatedAt(), 0),
		CorrelationID: operation.GetOperationId(),
		Attributes: map[string]string{
			"operation_id":  operation.GetOperationId(),
			"action":        action,
			"status":        status,
			"email_address": operation.GetEmailAddress(),
		},
	})
}
