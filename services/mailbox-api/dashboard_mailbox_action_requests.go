package main

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func normalizeDashboardOAuthRequest(req *mailboxv1.StartMailboxOAuthRequest) {
	if req.GetLimit() <= 0 {
		req.Limit = 100
	}
	req.EmailAddress = strings.TrimSpace(req.GetEmailAddress())
	if req.GetEmailAddress() == "" {
		req.OnlyMissing = true
	}
}

func normalizeDashboardInboxFetchRequest(req *mailboxv1.FetchMailboxInboxesRequest) {
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
}
