package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"google.golang.org/protobuf/proto"
	"mailboxapi/internal/protojsonhttp"
)

func readProtoJSON(r *http.Request, dst proto.Message) error {
	return protojsonhttp.ReadRequest(r, dst)
}

func writeProtoJSON(w http.ResponseWriter, status int, value proto.Message) {
	_ = protojsonhttp.WriteResponse(w, status, value)
}

type dashboardErrorMessageResponse interface {
	proto.Message
	GetErrorMessage() string
}

func writeProtoJSONWithErrorMessage(w http.ResponseWriter, status int, errorStatus int, value dashboardErrorMessageResponse) {
	if value.GetErrorMessage() != "" {
		writeError(w, errorStatus, errors.New(value.GetErrorMessage()))
		return
	}
	writeProtoJSON(w, status, value)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": safeMailboxError(err)})
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	w.Header().Set("Allow", method)
	w.WriteHeader(http.StatusMethodNotAllowed)
	return false
}
