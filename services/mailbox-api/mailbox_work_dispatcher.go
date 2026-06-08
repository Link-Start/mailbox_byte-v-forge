package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
	"mailboxapi/internal/eventoutbox"

	"mailboxapi/pb"
)

type mailboxWorkDispatcher struct {
	beginner eventoutbox.PgxBeginner
	source   string
}

func newMailboxWorkDispatcher(beginner eventoutbox.PgxBeginner, source string) *mailboxWorkDispatcher {
	if beginner == nil {
		return nil
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = mailboxEventSource
	}
	return &mailboxWorkDispatcher{beginner: beginner, source: source}
}

func (d *mailboxWorkDispatcher) PublishRegistrationRequested(ctx context.Context, operationID string) error {
	return d.publishOperationRequested(ctx, mailboxRegistrationRequested, "mailbox-registration-", operationID, &pb.MailboxRegistrationOperationRequest{OperationId: strings.TrimSpace(operationID)})
}

func (d *mailboxWorkDispatcher) PublishOAuthRequested(ctx context.Context, operationID string) error {
	return d.publishOperationRequested(ctx, mailboxOAuthRequested, "mailbox-oauth-", operationID, &pb.MailboxOAuthOperationRequest{OperationId: strings.TrimSpace(operationID)})
}

func (d *mailboxWorkDispatcher) publishOperationRequested(ctx context.Context, definition eventcatalog.Definition, eventPrefix string, operationID string, request proto.Message) error {
	if d == nil || d.beginner == nil {
		return fmt.Errorf("mailbox work dispatcher is not configured")
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return fmt.Errorf("operation_id is required")
	}
	metadata := d.metadata(definition.EventName, definition.Subject, eventbus.StableEventID(eventPrefix, operationID), operationID)
	record, err := eventoutbox.NewRecordFor(definition, request, metadata, eventbus.Attributes("operation_id", operationID))
	if err != nil {
		return err
	}
	return d.enqueue(ctx, record)
}
