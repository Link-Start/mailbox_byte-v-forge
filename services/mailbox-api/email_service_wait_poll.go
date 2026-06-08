package main

import (
	"context"
	"strings"
	"time"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
)

func (s *EmailService) requestMailboxEmailPoll(ctx context.Context, request *mailboxv1.WaitForMailboxEmailRequest, deadline time.Time, issuedAfterUnix int64) error {
	email := request.GetEmailAddress()
	if s.providers.IsStoredInboxOnlyAddress(email) {
		return nil
	}
	if s.work == nil {
		logWarning("mailbox email poll dispatcher is not configured; polling locally email=%s", emailx.Redact(email))
		go s.pollMailboxEmailLocally(ctx, email)
		return nil
	}
	return s.work.PublishEmailPollRequested(ctx, &mailboxv1.MailboxEmailPollRequest{
		EmailAddress:    email,
		SubjectKeyword:  strings.TrimSpace(request.GetSubjectKeyword()),
		ParserProfile:   strings.TrimSpace(request.GetParserProfile()),
		SignalKind:      request.GetSignalKind(),
		IssuedAfterUnix: issuedAfterUnix,
		DeadlineUnix:    deadline.Unix(),
		Reason:          "wait_for_email",
	})
}

func (s *EmailService) pollMailboxEmailLocally(ctx context.Context, email string) {
	if s == nil || s.watcher == nil {
		return
	}
	if err := s.watcher.PollForEmail(context.WithoutCancel(ctx), email); err != nil {
		logWarning("local mailbox email poll failed email=%s: %s", emailx.Redact(email), safeMailboxError(err))
	}
}
