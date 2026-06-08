package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"mailboxapi/internal/emailx"
	"mailboxapi/internal/redisx"
)

func (h *emailWebhookHandler) handleGraphNotification(w http.ResponseWriter, r *http.Request) {
	if token := r.URL.Query().Get("validationToken"); token != "" {
		if !h.validGraphWebhookRequestToken(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(token))
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		logWarning("read graph webhook body: %v", err)
		w.WriteHeader(http.StatusAccepted)
		return
	}
	var envelope graphNotificationEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		logWarning("decode graph webhook body: %v", err)
		w.WriteHeader(http.StatusAccepted)
		return
	}
	if !h.validGraphWebhookClientState(envelope) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	logInfo("received Outlook Graph webhook notifications=%d", len(envelope.Value))
	h.triggerRefresh()
	w.WriteHeader(http.StatusAccepted)
}

func (h *emailWebhookHandler) validGraphWebhookClientState(envelope graphNotificationEnvelope) bool {
	expected := h.config.token
	if expected == "" {
		logWarning("MAILBOX_WEBHOOK_TOKEN is required for Graph webhook ingestion")
		return false
	}
	if len(envelope.Value) == 0 {
		return false
	}
	for _, notification := range envelope.Value {
		if strings.TrimSpace(notification.ClientState) != expected {
			return false
		}
	}
	return true
}

func (h *emailWebhookHandler) validGraphWebhookRequestToken(r *http.Request) bool {
	expected := h.config.token
	if expected == "" {
		logWarning("MAILBOX_WEBHOOK_TOKEN is required for Graph webhook validation")
		return false
	}
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("webhook_token"))
	}
	if token == "" {
		token = strings.TrimSpace(r.Header.Get(defaultWebhookTokenHeader))
	}
	if token == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		token = strings.TrimPrefix(auth, "Bearer ")
	}
	return token == expected
}

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
	defer func() {
		if lock == nil {
			return
		}
		unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := lock.Unlock(unlockCtx); err != nil {
			logWarning("release Outlook webhook refresh lock failed: %v", err)
		}
	}()

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
