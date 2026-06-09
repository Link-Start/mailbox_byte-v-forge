package main

import (
	"net/http"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/httpx"
)

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
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	resp, err := s.mailboxClient.ListMailboxInbox(r.Context(), &mailboxv1.ListMailboxInboxRequest{
		EmailAddress:  strings.TrimSpace(email),
		Limit:         int32(httpx.QueryInt(r, "limit", 20)),
		ParserProfile: strings.TrimSpace(r.URL.Query().Get("parser_profile")),
		Cursor:        strings.TrimSpace(r.URL.Query().Get("cursor")),
		Query:         strings.TrimSpace(r.URL.Query().Get("q")),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeProtoJSONWithErrorMessage(w, http.StatusOK, http.StatusBadGateway, resp)
}

func (s *dashboardServer) handleMailboxStoredInboxMessage(w http.ResponseWriter, r *http.Request, email string, messageID string) {
	if !requireMethod(w, r, http.MethodGet) {
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
