package main

import (
	"strings"

	"mailboxapi/internal/emailx"

	"mailboxapi/pb"
)

func selectOAuthTargets(req *pb.RunMailboxOAuthRequest) []*pb.MailboxRegistrationAccount {
	requested := emailx.Normalize(req.GetEmailAddress())
	limit := req.GetLimit()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	targets := make([]*pb.MailboxRegistrationAccount, 0, len(req.GetAccounts()))
	for _, account := range req.GetAccounts() {
		email := emailx.Normalize(account.GetEmailAddress())
		if email == "" {
			continue
		}
		if requested != "" && email != requested {
			continue
		}
		if req.GetOnlyMissing() && strings.TrimSpace(account.GetRefreshToken()) != "" {
			continue
		}
		targets = append(targets, &pb.MailboxRegistrationAccount{
			EmailAddress: email,
			Password:     strings.TrimSpace(account.GetPassword()),
			RefreshToken: strings.TrimSpace(account.GetRefreshToken()),
			AccessToken:  strings.TrimSpace(account.GetAccessToken()),
			Source:       strings.TrimSpace(account.GetSource()),
		})
		if requested == "" && int32(len(targets)) >= limit {
			break
		}
	}
	return targets
}
