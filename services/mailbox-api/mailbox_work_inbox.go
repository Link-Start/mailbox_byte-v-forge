package main

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventoutbox"

	"mailboxapi/pb"
)

func (d *mailboxWorkDispatcher) PublishInboxFetchRequested(ctx context.Context, operationID string, request *mailboxv1.FetchMailboxInboxesRequest) error {
	if d == nil || d.beginner == nil {
		return fmt.Errorf("mailbox work dispatcher is not configured")
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return fmt.Errorf("operation_id is required")
	}
	if request == nil {
		request = &mailboxv1.FetchMailboxInboxesRequest{}
	}
	metadata := d.metadata(mailboxInboxFetchRequested.EventName, mailboxInboxFetchRequested.Subject, eventbus.StableEventID("mailbox-inbox-fetch-", operationID), operationID)
	record, err := eventoutbox.NewRecordFor(
		mailboxInboxFetchRequested,
		&pb.MailboxInboxFetchRequest{
			OperationId: operationID,
			Request:     proto.Clone(request).(*mailboxv1.FetchMailboxInboxesRequest),
		},
		metadata,
		eventbus.Attributes(
			"operation_id", operationID,
			"email_address", emailx.Normalize(request.GetEmailAddress()),
		),
	)
	if err != nil {
		return err
	}
	return d.enqueue(ctx, record)
}
