package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/emailx"
	"github.com/byte-v-forge/common-lib/envx"
	"github.com/byte-v-forge/common-lib/redisx"
)

func (h *graphWebhookHandler) handleGraphNotification(w http.ResponseWriter, r *http.Request) {
	if token := r.URL.Query().Get("validationToken"); token != "" {
		if !validGraphWebhookRequestToken(r) {
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
	if !validGraphWebhookClientState(envelope) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	logInfo("received Outlook Graph webhook notifications=%d", len(envelope.Value))
	h.triggerRefresh()
	w.WriteHeader(http.StatusAccepted)
}

func validGraphWebhookClientState(envelope graphNotificationEnvelope) bool {
	expected := webhookSecret()
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

func validGraphWebhookRequestToken(r *http.Request) bool {
	expected := webhookSecret()
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

func webhookSecret() string {
	return strings.TrimSpace(os.Getenv("MAILBOX_WEBHOOK_TOKEN"))
}

func (h *graphWebhookHandler) triggerRefresh() {
	if h.refreshLock == nil {
		logWarning("Outlook webhook refresh lock is not configured")
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

func (h *graphWebhookHandler) refreshMailboxes(lock *redisx.Lock) {
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := lock.Unlock(unlockCtx); err != nil {
			logWarning("release Outlook webhook refresh lock failed: %v", err)
		}
	}()

	timeout := envx.Int("OUTLOOK_WEBHOOK_FETCH_TIMEOUT_SECONDS", defaultWebhookTimeout)
	if timeout <= 0 {
		timeout = defaultWebhookTimeout
	}
	limit := envx.Int("OUTLOOK_WEBHOOK_MAX_MAILBOXES", defaultWebhookMaxMailboxes)
	if limit <= 0 {
		limit = defaultWebhookMaxMailboxes
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	mailboxes, err := h.watcher.store.ListOAuthMailboxes(ctx, int32(limit))
	if err != nil {
		logWarning("list OAuth mailboxes for webhook refresh: %v", err)
		return
	}
	fetched := 0
	failed := 0
	for _, mailbox := range mailboxes {
		if _, err := h.watcher.FetchMailboxInbox(ctx, mailbox, int32(h.watcher.messageLimit), 0); err != nil {
			failed++
			logWarning("webhook mailbox refresh failed for %s: %v", emailx.Redact(mailbox.GetEmailAddress()), err)
			continue
		}
		fetched++
	}
	logInfo("completed Outlook webhook refresh mailboxes=%d fetched=%d failed=%d", len(mailboxes), fetched, failed)
}
