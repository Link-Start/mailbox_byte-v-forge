package main

import (
	"errors"
	"net/http"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/httpx"
	"mailboxapi/internal/mailboxmodel"
)

func (s *dashboardServer) handleMailboxes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resp, err := s.mailboxClient.ListMailboxes(r.Context(), listMailboxesRequest(r))
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeProtoJSON(w, http.StatusOK, resp)
	case http.MethodPost:
		s.handleMailboxUpsert(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *dashboardServer) handleMailboxUpsert(w http.ResponseWriter, r *http.Request) {
	var req mailboxv1.UpsertEmailMailboxRequest
	if err := readProtoJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.GetMailbox() == nil || strings.TrimSpace(req.GetMailbox().GetEmailAddress()) == "" {
		writeError(w, http.StatusBadRequest, errors.New("mailbox.email_address is required"))
		return
	}
	req.Mailbox.EmailAddress = strings.TrimSpace(req.GetMailbox().GetEmailAddress())
	req.Mailbox.ProviderKey = strings.TrimSpace(req.GetMailbox().GetProviderKey())
	resp, err := s.mailboxClient.UpsertMailbox(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	email := strings.ToLower(strings.TrimSpace(resp.GetMailbox().GetEmailAddress()))
	if email == "" {
		writeError(w, http.StatusBadGateway, errors.New("mailbox returned empty mailbox"))
		return
	}
	writeProtoJSON(w, http.StatusCreated, resp)
}

func listMailboxesRequest(r *http.Request) *mailboxv1.ListEmailMailboxesRequest {
	authStatus := strings.TrimSpace(r.URL.Query().Get("auth_status"))
	if authStatus == "" {
		authStatus = strings.TrimSpace(r.URL.Query().Get("status"))
	}
	return &mailboxv1.ListEmailMailboxesRequest{
		AuthStatus:   mailboxmodel.PublicAuthStatus(authStatus),
		ProviderKey:  strings.TrimSpace(r.URL.Query().Get("provider_key")),
		EmailAddress: strings.TrimSpace(r.URL.Query().Get("email_address")),
		Cursor:       strings.TrimSpace(r.URL.Query().Get("cursor")),
		Limit:        int32(httpx.QueryInt(r, "limit", 100)),
	}
}
