package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
)

type dashboardMailboxRoute struct {
	email     string
	messageID string
	kind      dashboardMailboxRouteKind
}

type dashboardMailboxRouteKind int

const (
	dashboardMailboxRouteMailbox dashboardMailboxRouteKind = iota
	dashboardMailboxRouteInbox
	dashboardMailboxRouteInboxMessage
)

func parseDashboardMailboxRoute(path string) (dashboardMailboxRoute, int, error) {
	routePath := strings.Trim(strings.TrimPrefix(path, "/mailboxes/"), "/")
	parts := strings.Split(routePath, "/")
	email, err := decodeRequiredPathPart(parts[0], "email_address is required")
	if err != nil {
		return dashboardMailboxRoute{}, http.StatusBadRequest, err
	}
	if len(parts) == 1 {
		return dashboardMailboxRoute{email: email, kind: dashboardMailboxRouteMailbox}, http.StatusOK, nil
	}
	if len(parts) == 2 && parts[1] == "inbox" {
		return dashboardMailboxRoute{email: email, kind: dashboardMailboxRouteInbox}, http.StatusOK, nil
	}
	if len(parts) == 3 && parts[1] == "inbox" {
		messageID, err := decodeRequiredPathPart(parts[2], "message_id is required")
		if err != nil {
			return dashboardMailboxRoute{}, http.StatusBadRequest, err
		}
		return dashboardMailboxRoute{email: email, messageID: messageID, kind: dashboardMailboxRouteInboxMessage}, http.StatusOK, nil
	}
	return dashboardMailboxRoute{}, http.StatusNotFound, errors.New("mailbox endpoint not found")
}

func decodeRequiredPathPart(value string, message string) (string, error) {
	decoded, err := url.PathUnescape(value)
	if err != nil || strings.TrimSpace(decoded) == "" {
		return "", errors.New(message)
	}
	return decoded, nil
}
