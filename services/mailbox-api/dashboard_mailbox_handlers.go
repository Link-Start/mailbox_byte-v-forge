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
		limit := int32(httpx.QueryInt(r, "limit", 100))
		authStatus := strings.TrimSpace(r.URL.Query().Get("auth_status"))
		if authStatus == "" {
			authStatus = strings.TrimSpace(r.URL.Query().Get("status"))
		}
		resp, err := s.mailboxClient.ListMailboxes(r.Context(), &mailboxv1.ListEmailMailboxesRequest{
			AuthStatus:   mailboxmodel.PublicAuthStatus(authStatus),
			ProviderKey:  strings.TrimSpace(r.URL.Query().Get("provider_key")),
			EmailAddress: strings.TrimSpace(r.URL.Query().Get("email_address")),
			Cursor:       strings.TrimSpace(r.URL.Query().Get("cursor")),
			Limit:        limit,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeProtoJSON(w, http.StatusOK, resp)
	case http.MethodPost:
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *dashboardServer) handleMailbox(w http.ResponseWriter, r *http.Request) {
	route, statusCode, err := parseDashboardMailboxRoute(r.URL.Path)
	if err != nil {
		writeError(w, statusCode, err)
		return
	}
	switch route.kind {
	case dashboardMailboxRouteInbox:
		s.handleMailboxStoredInbox(w, r, route.email)
		return
	case dashboardMailboxRouteInboxMessage:
		s.handleMailboxStoredInboxMessage(w, r, route.email, route.messageID)
		return
	case dashboardMailboxRouteMailbox:
	}
	switch r.Method {
	case http.MethodDelete:
		resp, err := s.mailboxClient.DeleteMailbox(r.Context(), &mailboxv1.DeleteMailboxRequest{EmailAddress: strings.TrimSpace(route.email)})
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeProtoJSON(w, http.StatusOK, resp)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *dashboardServer) handleMailboxStoredInbox(w http.ResponseWriter, r *http.Request, email string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp, err := s.mailboxClient.ListMailboxInbox(r.Context(), &mailboxv1.ListMailboxInboxRequest{
		EmailAddress:  strings.TrimSpace(email),
		Limit:         int32(httpx.QueryInt(r, "limit", 20)),
		ParserProfile: strings.TrimSpace(r.URL.Query().Get("parser_profile")),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeProtoJSONWithErrorMessage(w, http.StatusOK, http.StatusBadGateway, resp)
}

func (s *dashboardServer) handleMailboxStoredInboxMessage(w http.ResponseWriter, r *http.Request, email string, messageID string) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp, err := s.mailboxClient.GetMailboxInboxMessage(r.Context(), &mailboxv1.GetMailboxInboxMessageRequest{
		EmailAddress:  strings.TrimSpace(email),
		MessageId:     strings.TrimSpace(messageID),
		ProviderKey:   strings.TrimSpace(r.URL.Query().Get("provider_key")),
		ParserProfile: strings.TrimSpace(r.URL.Query().Get("parser_profile")),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeProtoJSONWithErrorMessage(w, http.StatusOK, http.StatusBadGateway, resp)
}
