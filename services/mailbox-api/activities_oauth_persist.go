package main

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/pb"
)

func (a *mailboxActivities) persistOAuthResults(ctx context.Context, results []*pb.MailboxOAuthResult, accounts []*pb.MailboxRegistrationAccount) error {
	accountByEmail := make(map[string]*pb.MailboxRegistrationAccount, len(accounts))
	for _, account := range accounts {
		if email := emailx.Normalize(account.GetEmailAddress()); email != "" {
			accountByEmail[email] = account
		}
	}
	for _, result := range results {
		email := emailx.Normalize(result.GetEmailAddress())
		if email == "" {
			continue
		}
		refreshToken := strings.TrimSpace(result.GetRefreshToken())
		account := accountByEmail[email]
		password := ""
		existingRefreshToken := ""
		if account != nil {
			password = strings.TrimSpace(account.GetPassword())
			existingRefreshToken = strings.TrimSpace(account.GetRefreshToken())
		}
		if result.GetSuccess() && refreshToken != "" {
			if err := a.upsertMailbox(ctx, &mailboxmodel.Record{
				EmailAddress: email,
				Password:     password,
				RefreshToken: refreshToken,
				AccessToken:  strings.TrimSpace(result.GetAccessToken()),
				AuthStatus:   emailAuthAuthorized,
				LastError:    "",
			}); err != nil {
				return err
			}
			continue
		}
		if !result.GetSuccess() {
			errorMessage := result.GetErrorMessage()
			authStatus := mailboxOAuthFailureStatus(errorMessage)
			if mailboxOAuthFailureIsRuntime(errorMessage) && existingRefreshToken != "" {
				authStatus = emailAuthAuthorized
			}
			if err := a.markEmailAuthStatus(ctx, email, authStatus, safeMailboxText(errorMessage)); err != nil {
				return err
			}
		}
	}
	return nil
}
