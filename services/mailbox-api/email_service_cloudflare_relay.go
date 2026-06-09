package main

import (
	"context"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"

	"mailboxapi/pb"
)

func (s *EmailService) pullCloudflareRelayPendingForInbox(ctx context.Context, email string) {
	if s == nil || s.cloudflareRelay == nil || !s.isCloudflareStoredInbox(email) {
		return
	}
	events, err := s.cloudflareRelay.PullPending(ctx, email)
	if err != nil {
		logWarning("pull Cloudflare relay pending email=%s: %v", emailx.Redact(email), err)
		return
	}
	if len(events) == 0 {
		return
	}
	messages, ackIDs := s.recordCloudflareRelayEvents(ctx, events)
	if len(messages) > 0 {
		s.watcher.DispatchMailboxEvents(ctx, messages)
	}
	if err := s.cloudflareRelay.Ack(ctx, ackIDs); err != nil {
		logWarning("ack Cloudflare relay pending email=%s count=%d: %v", emailx.Redact(email), len(ackIDs), err)
		return
	}
	logInfo("pulled Cloudflare relay pending email=%s events=%d messages=%d", emailx.Redact(email), len(ackIDs), len(messages))
}

func (s *EmailService) recordCloudflareRelayEvents(ctx context.Context, events []*pb.InboundEmailWebhook) ([]*mailboxv1.EmailInboxMessage, []string) {
	messages := []*mailboxv1.EmailInboxMessage{}
	ackIDs := []string{}
	for _, event := range events {
		if event == nil {
			continue
		}
		event.ProviderKey = emailProviderCloudflare
		recorded, err := s.inbox.RecordInboundEmail(ctx, event)
		if err != nil {
			logWarning("record Cloudflare relay pending event_id=%s: %v", safeMailboxText(event.GetEventId()), err)
			continue
		}
		messages = append(messages, recorded...)
		if eventID := strings.TrimSpace(event.GetEventId()); eventID != "" {
			ackIDs = append(ackIDs, eventID)
		}
	}
	return messages, ackIDs
}

func (s *EmailService) isCloudflareStoredInbox(email string) bool {
	mailbox, ok := s.providers.StoredInboxOnlyMailbox(email)
	return ok && s.providers.normalizeProviderInput(mailbox.GetProviderKey()) == emailProviderCloudflare
}
