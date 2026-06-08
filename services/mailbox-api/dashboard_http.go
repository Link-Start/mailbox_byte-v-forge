package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"mailboxapi/internal/hotstream"
	"mailboxapi/internal/httpsse"

	"mailboxapi/pb"
)

type dashboardServer struct {
	mailboxClient pb.MailboxServiceClient
	hotstream     hotstream.Subscriber
	staticDir     string
	config        dashboardConfig
}

func startDashboardHTTP(ctx context.Context, listenAddr, staticDir string, config dashboardConfig, mailboxClient pb.MailboxServiceClient, stream hotstream.Subscriber, errCh chan<- error) {
	if strings.TrimSpace(listenAddr) == "" {
		return
	}
	if strings.TrimSpace(staticDir) == "" {
		staticDir = "/app/dashboard/mailbox"
	}
	dashboard := &dashboardServer{mailboxClient: mailboxClient, hotstream: stream, staticDir: staticDir, config: config}
	mux := http.NewServeMux()
	mux.Handle("/api/mailbox/", http.StripPrefix("/api/mailbox", dashboard.routes()))
	mux.Handle("/dashboard/mailbox/", http.StripPrefix("/dashboard/mailbox/", spaFileServer(staticDir)))
	mux.HandleFunc("/healthz", dashboard.handleHealth)
	mux.Handle("/", spaFileServer(staticDir))
	server := &http.Server{Addr: listenAddr, Handler: withCORS(mux), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	go func() {
		log.Printf("mailbox dashboard BFF listening on %s", listenAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("mailbox dashboard BFF failed: %w", err)
		}
	}()
}

func (s *dashboardServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/mailboxes/register", s.handleMailboxRegister)
	mux.HandleFunc("/mailboxes/oauth", s.handleMailboxOAuth)
	mux.HandleFunc("/mailboxes/inbox", s.handleMailboxInbox)
	mux.HandleFunc("/streams/state", s.streamState)
	mux.HandleFunc("/domains", s.handleMailboxDomains)
	mux.HandleFunc("/provider-capabilities", s.handleMailboxProviderCapabilities)
	mux.HandleFunc("/operations/", s.handleMailboxOperation)
	mux.HandleFunc("/operations", s.handleMailboxOperations)
	mux.HandleFunc("/mailboxes/", s.handleMailbox)
	mux.HandleFunc("/mailboxes", s.handleMailboxes)
	return mux
}

func (s *dashboardServer) streamState(w http.ResponseWriter, r *http.Request) {
	httpsse.ServeHotStream(w, r, s.hotstream, httpsse.FilterFromRequest(r, hotstream.Filter{
		SourceServices: []string{mailboxHotStreamSource},
	}), httpsse.ServeOptions{})
}

func (s *dashboardServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
