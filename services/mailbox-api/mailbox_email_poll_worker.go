package main

import (
	"context"
	"log"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/eventbus"
	"mailboxapi/internal/eventcatalog"
)

type mailboxEmailPollWorker struct {
	service *EmailService
}

func runMailboxEmailPollWorker(ctx context.Context, consumer eventbus.Consumer, service *EmailService) error {
	worker := &mailboxEmailPollWorker{service: service}
	return eventbus.RunTypedConsumerWorker(ctx, eventbus.TypedConsumerWorkerConfig[*mailboxv1.MailboxEmailPollRequest]{
		Name:           "mailbox email poll requests",
		Consumer:       consumer,
		Expected:       eventcatalog.MailboxEmailPollRequested.ExpectedMessage(),
		NewMessage:     func() *mailboxv1.MailboxEmailPollRequest { return &mailboxv1.MailboxEmailPollRequest{} },
		Handler:        worker.handle,
		MalformedLabel: "terminate malformed mailbox email poll request",
	})
}

func (w *mailboxEmailPollWorker) handle(ctx context.Context, request *mailboxv1.MailboxEmailPollRequest, _ eventbus.ReceivedMessage) eventbus.HandlerResult {
	email := emailx.Normalize(request.GetEmailAddress())
	if email == "" {
		return eventbus.TermResult("terminate mailbox email poll request without email")
	}
	if w.service.providers.IsStoredInboxOnlyAddress(email) {
		return eventbus.AckResult("ack stored-inbox-only mailbox email poll request")
	}
	if deadlineReached(request.GetDeadlineUnix()) {
		return eventbus.AckResult("ack expired mailbox email poll request")
	}
	if err := w.service.watcher.PollForEmail(ctx, email); err != nil {
		if isAuthError(err) {
			log.Printf("mailbox email poll auth failure email=%s: %s", emailx.Redact(email), safeMailboxError(err))
			return eventbus.TermResult("terminate auth-failed mailbox email poll request")
		}
		delay := mailboxPollRetryDelay(err, w.service.watcher.DefaultPollInterval())
		log.Printf("mailbox email poll failed email=%s: %s", emailx.Redact(email), safeMailboxError(err))
		return eventbus.NakResult(delay, "delay mailbox email poll retry")
	}
	if _, found, err := w.service.latestEmailResponse(ctx, &mailboxv1.WaitForMailboxEmailRequest{
		EmailAddress:    email,
		SubjectKeyword:  request.GetSubjectKeyword(),
		ParserProfile:   request.GetParserProfile(),
		SignalKind:      request.GetSignalKind(),
		IssuedAfterUnix: request.GetIssuedAfterUnix(),
	}, request.GetIssuedAfterUnix()); err != nil {
		log.Printf("mailbox email poll projection check failed email=%s: %s", emailx.Redact(email), safeMailboxError(err))
		return eventbus.NakResult(defaultMailboxWorkRetryDelay, "retry mailbox email poll projection")
	} else if found {
		return eventbus.AckResult("ack found mailbox email poll request")
	}
	if deadlineReached(request.GetDeadlineUnix()) {
		return eventbus.AckResult("ack timed-out mailbox email poll request")
	}
	return eventbus.NakResult(mailboxPollInterval(w.service.watcher.DefaultPollInterval()), "delay mailbox email poll")
}
