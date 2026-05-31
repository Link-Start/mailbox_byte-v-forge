package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/eventbus"
	"github.com/byte-v-forge/common-lib/eventcatalog"
	"github.com/byte-v-forge/common-lib/eventoutbox"
	commonv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/common/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"

	"mailboxapi/pb"
)

type mailboxWorkDispatcher struct {
	db     *gorm.DB
	source string
}

func newMailboxWorkDispatcher(db *gorm.DB, source string) *mailboxWorkDispatcher {
	if db == nil {
		return nil
	}
	source = strings.TrimSpace(source)
	if source == "" {
		source = mailboxPlatformEventSource
	}
	return &mailboxWorkDispatcher{db: db, source: source}
}

func (d *mailboxWorkDispatcher) PublishRegistrationRequested(ctx context.Context, operationID string) error {
	return d.publishOperationRequested(ctx, eventcatalog.MailboxRegistrationRequested, "mailbox-registration-", operationID, &pb.MailboxRegistrationOperationRequest{OperationId: strings.TrimSpace(operationID)})
}

func (d *mailboxWorkDispatcher) PublishOAuthRequested(ctx context.Context, operationID string) error {
	return d.publishOperationRequested(ctx, eventcatalog.MailboxOAuthRequested, "mailbox-oauth-", operationID, &pb.MailboxOAuthOperationRequest{OperationId: strings.TrimSpace(operationID)})
}

func (d *mailboxWorkDispatcher) publishOperationRequested(ctx context.Context, definition eventcatalog.Definition, eventPrefix string, operationID string, request proto.Message) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("mailbox work dispatcher is not configured")
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return fmt.Errorf("operation_id is required")
	}
	eventCtx := d.context(definition.EventName, eventbus.StableEventID(eventPrefix, operationID), operationID)
	return d.enqueue(ctx, eventbus.Message{
		Subject:    definition.Subject,
		Event:      request,
		Context:    eventCtx,
		Attributes: eventbus.Attributes("operation_id", operationID),
	})
}

func (d *mailboxWorkDispatcher) PublishEmailPollRequested(ctx context.Context, request *mailboxv1.MailboxEmailPollRequest) error {
	if d == nil || d.db == nil || request == nil {
		return nil
	}
	request.EmailAddress = emailx.Normalize(request.GetEmailAddress())
	eventCtx := d.context(
		eventcatalog.MailboxEmailPollRequested.EventName,
		eventbus.StableEventID("mailbox-email-poll-", request.GetEmailAddress(), request.GetSubjectKeyword(), request.GetParserProfile(), request.GetSignalKind().String(), fmt.Sprintf("%d", request.GetIssuedAfterUnix()), fmt.Sprintf("%d", request.GetDeadlineUnix())),
		request.GetEmailAddress(),
	)
	return d.enqueue(ctx, eventbus.Message{
		Subject: eventcatalog.MailboxEmailPollRequested.Subject,
		Event:   request,
		Context: eventCtx,
		Attributes: eventbus.Attributes(
			"email_address", request.GetEmailAddress(),
			"signal_kind", request.GetSignalKind().String(),
			"reason", request.GetReason(),
		),
	})
}

func (d *mailboxWorkDispatcher) PublishInboxFetchRequested(ctx context.Context, operationID string, request *mailboxv1.FetchMailboxInboxesRequest) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("mailbox work dispatcher is not configured")
	}
	operationID = strings.TrimSpace(operationID)
	if operationID == "" {
		return fmt.Errorf("operation_id is required")
	}
	if request == nil {
		request = &mailboxv1.FetchMailboxInboxesRequest{}
	}
	eventCtx := d.context(eventcatalog.MailboxInboxFetchRequested.EventName, eventbus.StableEventID("mailbox-inbox-fetch-", operationID), operationID)
	return d.enqueue(ctx, eventbus.Message{
		Subject: eventcatalog.MailboxInboxFetchRequested.Subject,
		Event: &pb.MailboxInboxFetchRequest{
			OperationId: operationID,
			Request:     proto.Clone(request).(*mailboxv1.FetchMailboxInboxesRequest),
		},
		Context: eventCtx,
		Attributes: eventbus.Attributes(
			"operation_id", operationID,
			"email_address", emailx.Normalize(request.GetEmailAddress()),
		),
	})
}

func (d *mailboxWorkDispatcher) context(eventName string, eventID string, correlationID string) *commonv1.EventContext {
	return eventbus.NewEventContext(eventbus.EventContextConfig{
		EventID:       eventID,
		EventName:     eventName,
		EventVersion:  mailboxPlatformEventVersion,
		SourceService: d.source,
		CorrelationID: correlationID,
	})
}

func (d *mailboxWorkDispatcher) enqueue(ctx context.Context, message eventbus.Message) error {
	record, err := eventoutbox.NewRecord(message)
	if err != nil {
		return err
	}
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return eventoutbox.InsertRecordGORM(ctx, tx, mailboxPlatformEventOutboxTable, record, time.Now().Unix())
	})
}
