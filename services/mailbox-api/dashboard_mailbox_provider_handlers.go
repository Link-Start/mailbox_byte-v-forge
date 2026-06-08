package main

import (
	"net/http"
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func (s *dashboardServer) handleMailboxDomains(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resp, err := s.mailboxClient.ListMailboxDomains(r.Context(), &mailboxv1.ListMailboxDomainsRequest{
			ProviderKey: requestProviderKey(r),
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeProtoJSONWithErrorMessage(w, http.StatusOK, http.StatusBadGateway, resp)
	case http.MethodPost:
		var req mailboxv1.SyncMailboxDomainsRequest
		if err := readProtoJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		resp, err := s.mailboxClient.SyncMailboxDomains(r.Context(), &req)
		if err != nil {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeProtoJSONWithErrorMessage(w, http.StatusOK, http.StatusBadGateway, resp)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *dashboardServer) handleMailboxProviderCapabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp, err := s.mailboxClient.ListMailboxProviderCapabilities(r.Context(), &mailboxv1.ListMailboxProviderCapabilitiesRequest{
		ProviderKey: requestProviderKey(r),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeProtoJSONWithErrorMessage(w, http.StatusOK, http.StatusBadGateway, resp)
}

func requestProviderKey(r *http.Request) string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.URL.Query().Get("provider_key"))
}
