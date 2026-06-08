package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
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
