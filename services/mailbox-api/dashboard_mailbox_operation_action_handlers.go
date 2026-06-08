package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func (s *dashboardServer) handleMailboxRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req mailboxv1.StartMailboxOAuthRequest
	if err := readProtoJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.GetLimit() <= 0 {
		req.Limit = 100
	}
	req.EmailAddress = strings.TrimSpace(req.GetEmailAddress())
	if req.GetEmailAddress() == "" {
		req.OnlyMissing = true
	}

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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req mailboxv1.FetchMailboxInboxesRequest
	if err := readProtoJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.GetLimitPerMailbox() <= 0 {
		req.LimitPerMailbox = 10
	}
	if req.GetLimitPerMailbox() > 100 {
		req.LimitPerMailbox = 100
	}
	if req.GetMaxMailboxes() <= 0 {
		req.MaxMailboxes = 100
	}
	if req.GetMaxMailboxes() > 500 {
		req.MaxMailboxes = 500
	}
	req.EmailAddress = strings.TrimSpace(req.GetEmailAddress())
	req.ParserProfile = strings.TrimSpace(req.GetParserProfile())

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
