package mailboxapp

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/mailboxmodel"
)

func credentialState(mailbox *mailboxmodel.Record) *mailboxv1.MailboxCredentialState {
	if mailbox == nil {
		return nil
	}
	passwordPresent := strings.TrimSpace(mailbox.GetPassword()) != ""
	refreshTokenPresent := strings.TrimSpace(mailbox.GetRefreshToken()) != ""
	accessTokenPresent := strings.TrimSpace(mailbox.GetAccessToken()) != ""
	return &mailboxv1.MailboxCredentialState{
		PasswordPresent:          passwordPresent,
		OauthRefreshTokenPresent: refreshTokenPresent,
		OauthAccessTokenPresent:  accessTokenPresent,
		PresentCredentials: presentCredentialKinds(
			passwordPresent,
			refreshTokenPresent,
			accessTokenPresent,
		),
	}
}

func presentCredentialKinds(passwordPresent bool, refreshTokenPresent bool, accessTokenPresent bool) []mailboxv1.MailboxCredentialKind {
	present := []mailboxv1.MailboxCredentialKind{}
	appendPresent := func(kind mailboxv1.MailboxCredentialKind, exists bool) {
		if exists {
			present = append(present, kind)
		}
	}
	appendPresent(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD, passwordPresent)
	appendPresent(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN, refreshTokenPresent)
	appendPresent(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN, accessTokenPresent)
	return present
}
