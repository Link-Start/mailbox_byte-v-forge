package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/byte-v-forge/common-lib/protojsonx"

	"mailboxapi/pb"
)

func (h *emailWebhookHandler) handleInboundEmailWebhook(providerKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !h.validWebhookToken(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2<<20))
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		var event pb.InboundEmailWebhook
		if err := protojsonx.Unmarshal(raw, &event); err != nil {
			http.Error(w, "invalid email event", http.StatusBadRequest)
			return
		}
		event.ProviderKey = providerKey
		messages, err := h.inbox.RecordInboundEmail(r.Context(), &event)
		if err != nil {
			logWarning("record %s email webhook: %v", providerKey, err)
			http.Error(w, "record email event failed", http.StatusInternalServerError)
			return
		}
		h.watcher.DispatchMailboxEvents(r.Context(), messages)
		logInfo("recorded %s email event recipients=%d message_id=%s", providerKey, len(event.GetRecipients()), event.GetMessageId())
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *emailWebhookHandler) validWebhookToken(r *http.Request) bool {
	expected := h.config.token
	if expected == "" {
		logWarning("MAILBOX_WEBHOOK_TOKEN is required for email webhook ingestion")
		return false
	}
	token := strings.TrimSpace(r.Header.Get(defaultWebhookTokenHeader))
	return token == expected
}
