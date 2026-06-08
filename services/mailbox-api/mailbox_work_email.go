package main

import (
	"context"
	"fmt"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
	"mailboxapi/internal/eventoutbox"
)

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
