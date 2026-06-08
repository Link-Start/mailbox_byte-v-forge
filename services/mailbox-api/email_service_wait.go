package main

import (
	"context"
	"errors"
	"strings"
	"time"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mailboxapi/internal/emailx"
)

func (s *EmailService) WaitForEmail(ctx context.Context, request *mailboxv1.WaitForMailboxEmailRequest) (*mailboxv1.WaitForMailboxEmailResponse, error) {
	timeoutSeconds := request.GetTimeoutSeconds()
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300
	}
	email := request.GetEmailAddress()
	issuedAfterUnix := request.GetIssuedAfterUnix()
	logInfo("waiting for email message email=%s timeout_seconds=%d issued_after_unix=%d", emailx.Redact(email), timeoutSeconds, issuedAfterUnix)
	if resp, ok, err := s.latestEmailResponse(ctx, request, issuedAfterUnix); err != nil {
		return nil, waitError(ctx, err)
	} else if ok {
		return resp, nil
	}
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	if !s.providers.IsStoredInboxOnlyAddress(email) {
		if s.work == nil {
			logWarning("mailbox email poll dispatcher is not configured; polling locally email=%s", emailx.Redact(email))
			go s.pollMailboxEmailLocally(ctx, email)
		} else if err := s.work.PublishEmailPollRequested(ctx, &mailboxv1.MailboxEmailPollRequest{
			EmailAddress:    email,
			SubjectKeyword:  strings.TrimSpace(request.GetSubjectKeyword()),
			ParserProfile:   strings.TrimSpace(request.GetParserProfile()),
			SignalKind:      request.GetSignalKind(),
			IssuedAfterUnix: issuedAfterUnix,
			DeadlineUnix:    deadline.Unix(),
			Reason:          "wait_for_email",
		}); err != nil {
			return nil, waitError(ctx, err)
		}
	}
	return s.waitForPersistedEmail(ctx, request, timeoutSeconds, issuedAfterUnix)
}

func (s *EmailService) pollMailboxEmailLocally(ctx context.Context, email string) {
	if s == nil || s.watcher == nil {
		return
	}
	if err := s.watcher.PollForEmail(context.WithoutCancel(ctx), email); err != nil {
		logWarning("local mailbox email poll failed email=%s: %s", emailx.Redact(email), safeMailboxError(err))
	}
}

func (s *EmailService) waitForPersistedEmail(ctx context.Context, request *mailboxv1.WaitForMailboxEmailRequest, timeoutSeconds int32, issuedAfterUnix int64) (*mailboxv1.WaitForMailboxEmailResponse, error) {
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	for time.Now().Before(deadline) {
		sleepFor := time.Duration(s.watcher.DefaultPollInterval()) * time.Second
		if remaining := time.Until(deadline); remaining < sleepFor {
			sleepFor = remaining
		}
		if sleepFor > 0 {
			timer := time.NewTimer(sleepFor)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, status.Error(codes.Canceled, "request cancelled")
			case <-timer.C:
			}
		}
		if resp, ok, err := s.latestEmailResponse(ctx, request, issuedAfterUnix); err != nil {
			return nil, waitError(ctx, err)
		} else if ok {
			return resp, nil
		}
	}
	logInfo("webhook-backed email message not found email=%s timeout_seconds=%d issued_after_unix=%d", emailx.Redact(request.GetEmailAddress()), timeoutSeconds, issuedAfterUnix)
	return &mailboxv1.WaitForMailboxEmailResponse{Found: false}, nil
}

func (s *EmailService) latestEmailResponse(ctx context.Context, request *mailboxv1.WaitForMailboxEmailRequest, issuedAfterUnix int64) (*mailboxv1.WaitForMailboxEmailResponse, bool, error) {
	message, ok, err := s.inbox.LatestMessageWithSignal(ctx, request.GetEmailAddress(), request.GetSubjectKeyword(), issuedAfterUnix, request.GetParserProfile(), request.GetSignalKind())
	if err != nil || !ok {
		return nil, false, err
	}
	logInfo("served persisted email for %s provider=%s received_at_unix=%d", emailx.Redact(request.GetEmailAddress()), message.GetProviderKey(), message.GetReceivedAtUnix())
	return &mailboxv1.WaitForMailboxEmailResponse{Found: true, Message: message}, true, nil
}

func waitError(ctx context.Context, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled) {
		return status.Error(codes.Canceled, "request cancelled")
	}
	return status.Error(codes.Internal, safeMailboxError(err))
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "not authorized") ||
		strings.Contains(msg, "no refresh token") ||
		strings.Contains(msg, "AUTH_FAILED")
}
