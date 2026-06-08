package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mailboxapi/internal/redisx"

	"mailboxapi/internal/inboxapp"
)

type emailWebhookHandler struct {
	inbox       *inboxapp.Service
	watcher     *MailWatcher
	refreshLock *redisx.BestEffortLocker
	config      emailWebhookConfig
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

func startWebhookServer(ctx context.Context, addr string, config emailWebhookConfig, providers mailboxProviderRuntimeConfig, inbox *inboxapp.Service, watcher *MailWatcher, refreshLock *redisx.BestEffortLocker, errCh chan<- error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return
	}

	handler := &emailWebhookHandler{
		inbox:       inbox,
		watcher:     watcher,
		refreshLock: refreshLock,
		config:      config,
	}
	mux := http.NewServeMux()
	for _, route := range newMailboxWebhookRegistryForProviders(providers, mailboxWebhookDependencies{handler: handler}).Routes() {
		mux.HandleFunc(route.path, route.handler)
	}
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
