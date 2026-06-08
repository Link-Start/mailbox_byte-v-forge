package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	commonv1 "mailboxapi/internal/contracts/commonv1"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
	"mailboxapi/internal/eventoutbox"

	"mailboxapi/internal/inboxapp"
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

func (d *mailboxWorkDispatcher) PublishEmailPollRequested(ctx context.Context, request *mailboxv1.MailboxEmailPollRequest) error {
	if d == nil || d.beginner == nil || request == nil {
		return nil
	}
	request.EmailAddress = emailx.Normalize(request.GetEmailAddress())
	metadata := d.metadata(
		eventcatalog.MailboxEmailPollRequested.EventName,
		eventcatalog.MailboxEmailPollRequested.Subject,
		eventbus.StableEventID("mailbox-email-poll-", request.GetEmailAddress(), request.GetSubjectKeyword(), request.GetParserProfile(), request.GetSignalKind().String(), fmt.Sprintf("%d", request.GetIssuedAfterUnix()), fmt.Sprintf("%d", request.GetDeadlineUnix())),
		request.GetEmailAddress(),
	)
	record, err := eventoutbox.NewRecordFor(
		eventcatalog.MailboxEmailPollRequested,
		request,
		metadata,
		eventbus.Attributes(
			"email_address", request.GetEmailAddress(),
			"signal_kind", request.GetSignalKind().String(),
			"reason", request.GetReason(),
		),
	)
	if err != nil {
		return err
	}
	return d.enqueue(ctx, record)
}

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

func (d *mailboxWorkDispatcher) metadata(eventName string, subject string, eventID string, correlationID string) *commonv1.EventMetadata {
	return eventbus.NewEventMetadata(eventbus.EventMetadataConfig{
		EventID:       eventID,
		EventName:     eventName,
		EventVersion:  inboxapp.EventVersion,
		SourceService: d.source,
		Subject:       subject,
		CorrelationID: correlationID,
	})
}

func (d *mailboxWorkDispatcher) enqueue(ctx context.Context, record eventoutbox.Record) error {
	tx, err := d.beginner.Begin(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()
	if err := eventoutbox.InsertRecordPgx(ctx, tx, mailboxEventOutboxTable, record, time.Now().Unix()); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true
	return nil
}
