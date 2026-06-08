package main

import (
	"context"
	"net/http"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func (s *dashboardServer) handleMailboxRegister(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	resp, err := s.mailboxClient.RegisterMailbox(ctx, &mailboxv1.RegisterMailboxRequest{})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}

	writeMailboxOperationStart(w, resp)
}

func (s *dashboardServer) handleMailboxOAuth(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req mailboxv1.StartMailboxOAuthRequest
	if err := readProtoJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	normalizeDashboardOAuthRequest(&req)

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	resp, err := s.mailboxClient.RunMailboxOAuth(ctx, &req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeMailboxOperationStart(w, resp)
}

func (s *dashboardServer) handleMailboxInbox(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	var req mailboxv1.FetchMailboxInboxesRequest
	if err := readProtoJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	normalizeDashboardInboxFetchRequest(&req)

	ctx, cancel := context.WithTimeout(r.Context(), s.config.inboxTimeout)
	defer cancel()

	resp, err := s.mailboxClient.FetchMailboxInboxes(ctx, &req)
	if err != nil {
		if status.Code(err) == codes.DeadlineExceeded {
			writeError(w, http.StatusGatewayTimeout, err)
			return
		}
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeProtoJSON(w, http.StatusOK, resp)
}

type mailboxOperationStartResponse interface {
	proto.Message
	GetStarted() bool
	GetOperationId() string
	GetErrorMessage() string
}

func writeMailboxOperationStart(w http.ResponseWriter, resp mailboxOperationStartResponse) {
	statusCode := http.StatusAccepted
	if !resp.GetStarted() || resp.GetErrorMessage() != "" {
		statusCode = http.StatusBadGateway
	}
	writeProtoJSON(w, statusCode, resp)
}
