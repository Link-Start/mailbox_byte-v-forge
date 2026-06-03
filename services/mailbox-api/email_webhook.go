package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/byte-v-forge/common-lib/redisx"

	"mailboxapi/internal/inboxapp"
)

type graphWebhookHandler struct {
	inbox       *inboxapp.Service
	watcher     *MailWatcher
	refreshLock *redisx.BestEffortLocker
}

type graphNotificationEnvelope struct {
	Value []graphNotification `json:"value"`
}

type graphNotification struct {
	SubscriptionID string `json:"subscriptionId"`
	ClientState    string `json:"clientState"`
	Resource       string `json:"resource"`
	ChangeType     string `json:"changeType"`
}

func startWebhookServer(ctx context.Context, addr string, inbox *inboxapp.Service, watcher *MailWatcher, refreshLock *redisx.BestEffortLocker, errCh chan<- error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return
	}

	handler := &graphWebhookHandler{
		inbox:       inbox,
		watcher:     watcher,
		refreshLock: refreshLock,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/email/cloudflare", handler.handleCloudflareEmail)
	mux.HandleFunc("/webhooks/email/microsoft-graph", handler.handleGraphNotification)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logWarning("shutdown webhook server: %v", err)
		}
	}()
	go func() {
		logInfo("Starting mailbox webhook server on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("serve webhook: %w", err)
		}
	}()
}
