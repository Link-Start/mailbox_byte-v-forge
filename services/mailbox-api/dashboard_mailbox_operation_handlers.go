package main

import (
	"errors"
	"net/http"
	"strings"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"mailboxapi/internal/httpx"
)

func (s *dashboardServer) handleMailboxOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp, err := s.mailboxClient.ListMailboxOperations(r.Context(), &mailboxv1.ListMailboxOperationsRequest{
		Limit:        int32(httpx.QueryInt(r, "limit", 50)),
		Status:       publicOperationStatus(strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status")))),
		Action:       publicOperationAction(strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("action")))),
		EmailAddress: strings.TrimSpace(r.URL.Query().Get("email_address")),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if resp.GetErrorMessage() != "" {
		writeError(w, http.StatusBadGateway, errors.New(resp.GetErrorMessage()))
		return
	}
	writeProtoJSON(w, http.StatusOK, resp)
}

func (s *dashboardServer) handleMailboxOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	operationID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/operations/"), "/")
	if operationID == "" {
		writeError(w, http.StatusBadRequest, errors.New("operation_id is required"))
		return
	}
	resp, err := s.mailboxClient.GetMailboxOperation(r.Context(), &mailboxv1.GetMailboxOperationRequest{OperationId: operationID})
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	if resp.GetErrorMessage() != "" {
		writeError(w, http.StatusNotFound, errors.New(resp.GetErrorMessage()))
		return
	}
	writeProtoJSON(w, http.StatusOK, resp)
}
