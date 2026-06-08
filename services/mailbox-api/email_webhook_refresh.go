package main

import (
	"context"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/redisx"
)

func (h *emailWebhookHandler) triggerRefresh() {
	if h.refreshLock == nil {
		logWarning("Outlook webhook refresh lock is not configured; running refresh without distributed lock")
		go h.refreshMailboxes(nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	lock, err := h.refreshLock.Lock(ctx, "graph-webhook-refresh")
	cancel()
	if err != nil {
		logInfo("Outlook webhook refresh already running")
		return
	}
	go h.refreshMailboxes(lock)
}

func (h *emailWebhookHandler) refreshMailboxes(lock *redisx.Lock) {
	defer releaseGraphWebhookRefreshLock(lock)

	ctx, cancel := context.WithTimeout(context.Background(), h.config.outlookFetchTimeout)
	defer cancel()

	mailboxes, err := h.watcher.mailboxes.ListOAuthMailboxes(ctx, int32(h.config.outlookRefreshMaxMailbox))
	if err != nil {
		logWarning("list OAuth mailboxes for webhook refresh: %v", err)
		return
	}
	fetched := 0
	failed := 0
	messageLimit := int32(h.watcher.DefaultMessageLimit())
	for _, mailbox := range mailboxes {
		if _, err := h.watcher.FetchMailboxInbox(ctx, mailbox, messageLimit, 0); err != nil {
			failed++
			logWarning("webhook mailbox refresh failed for %s: %v", emailx.Redact(mailbox.GetEmailAddress()), err)
			continue
		}
		fetched++
	}
	logInfo("completed Outlook webhook refresh mailboxes=%d fetched=%d failed=%d", len(mailboxes), fetched, failed)
}

func releaseGraphWebhookRefreshLock(lock *redisx.Lock) {
	if lock == nil {
		return
	}
	unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := lock.Unlock(unlockCtx); err != nil {
		logWarning("release Outlook webhook refresh lock failed: %v", err)
	}
}
