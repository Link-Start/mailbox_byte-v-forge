package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/accountmodel"
	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/pb"
)

const oauthAccountScanMaxPages = 20

func (a *mailboxActivities) oauthAccounts(ctx context.Context, emailAddress string, onlyMissing bool, limit int32) ([]*pb.MailboxRegistrationAccount, error) {
	requestedEmail := emailx.Normalize(emailAddress)
	if requestedEmail != "" {
		return a.oauthAccountByEmail(ctx, requestedEmail, onlyMissing)
	}
	selectedLimit := accountmodel.NormalizePageLimit(int(limit))
	accounts := make([]*pb.MailboxRegistrationAccount, 0, selectedLimit)
	cursor := ""
	for page := 0; len(accounts) < selectedLimit && page < oauthAccountScanMaxPages; page++ {
		resp, err := a.emailBackend.ListMailboxes(ctx, &pb.ListEmailMailboxesRequest{Limit: int32(selectedLimit), Cursor: cursor})
		if err != nil {
			return nil, fmt.Errorf("list mailboxes: %s", safeMailboxError(err))
		}
		if resp == nil {
			return nil, fmt.Errorf("email service returned empty mailbox list")
		}
		accounts = appendOAuthAccounts(accounts, resp.GetMailboxes(), "", onlyMissing, selectedLimit)
		cursor = strings.TrimSpace(resp.GetNextCursor())
		if cursor == "" {
			break
		}
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no mailbox accounts eligible for OAuth")
	}
	return accounts, nil
}

func (a *mailboxActivities) oauthAccountByEmail(ctx context.Context, requestedEmail string, onlyMissing bool) ([]*pb.MailboxRegistrationAccount, error) {
	resp, err := a.emailBackend.ListMailboxes(ctx, &pb.ListEmailMailboxesRequest{Limit: 1, EmailAddress: requestedEmail})
	if err != nil {
		return nil, fmt.Errorf("list mailbox: %s", safeMailboxError(err))
	}
	if resp == nil {
		return nil, fmt.Errorf("email service returned empty mailbox list")
	}
	accounts := appendOAuthAccounts(nil, resp.GetMailboxes(), requestedEmail, onlyMissing, 1)
	if len(accounts) == 0 {
		return nil, fmt.Errorf("mailbox not found or not eligible for OAuth: %s", emailx.Redact(requestedEmail))
	}
	return accounts, nil
}

func appendOAuthAccounts(accounts []*pb.MailboxRegistrationAccount, mailboxes []*pb.EmailMailbox, requestedEmail string, onlyMissing bool, limit int) []*pb.MailboxRegistrationAccount {
	for _, mailbox := range mailboxes {
		email := emailx.Normalize(mailbox.GetEmailAddress())
		if email == "" {
			continue
		}
		if requestedEmail != "" && email != requestedEmail {
			continue
		}
		if strings.TrimSpace(mailbox.GetPassword()) == "" {
			continue
		}
		authStatus := mailboxAuthStatus(mailbox.GetRefreshToken(), mailbox.GetAuthStatus())
		if onlyMissing && (authStatus == emailAuthAuthorized || authStatus == emailAuthNeedsManualVerification) {
			continue
		}
		accounts = append(accounts, &pb.MailboxRegistrationAccount{
			EmailAddress: email,
			Password:     strings.TrimSpace(mailbox.GetPassword()),
			RefreshToken: strings.TrimSpace(mailbox.GetRefreshToken()),
			AccessToken:  strings.TrimSpace(mailbox.GetAccessToken()),
			Source:       "mailboxes",
		})
		if requestedEmail == "" && len(accounts) >= limit {
			break
		}
	}
	return accounts
}
