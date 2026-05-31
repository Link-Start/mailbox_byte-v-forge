package main

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/byte-v-forge/common-lib/protojsonx"

	"mailboxapi/pb"
)

func (h *graphWebhookHandler) handleCloudflareEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !validWebhookToken(r) {
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
	event.ProviderKey = emailProviderCloudflare
	messages, err := h.store.RecordInboundEmail(r.Context(), &event)
	if err != nil {
		logWarning("record Cloudflare email webhook: %v", err)
		http.Error(w, "record email event failed", http.StatusInternalServerError)
		return
	}
	h.watcher.DispatchMailboxEvents(r.Context(), messages)
	logInfo("recorded Cloudflare email event recipients=%d message_id=%s", len(event.GetRecipients()), event.GetMessageId())
	w.WriteHeader(http.StatusAccepted)
}

func validWebhookToken(r *http.Request) bool {
	expected := strings.TrimSpace(os.Getenv("MAILBOX_WEBHOOK_TOKEN"))
	if expected == "" {
		logWarning("MAILBOX_WEBHOOK_TOKEN is required for email webhook ingestion")
		return false
	}
	token := strings.TrimSpace(r.Header.Get(defaultWebhookTokenHeader))
	return token == expected
}
