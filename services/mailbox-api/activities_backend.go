package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/byte-v-forge/common-lib/emailx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxmodel"
)

func (a *mailboxActivities) upsertMailbox(ctx context.Context, mailbox *mailboxmodel.Record) error {
	resp, err := a.emailBackend.UpsertMailbox(ctx, &mailboxv1.UpsertEmailMailboxRequest{Mailbox: mailboxCredentialInput(mailbox)})
	if err != nil {
		return fmt.Errorf("upsert mailbox %s: %s", emailx.Redact(mailbox.GetEmailAddress()), safeMailboxError(err))
	}
	if resp == nil || resp.GetMailbox() == nil {
		return fmt.Errorf("email service returned empty mailbox for %s", emailx.Redact(mailbox.GetEmailAddress()))
	}
	return nil
}

func (a *mailboxActivities) markEmailAuthStatus(ctx context.Context, email string, authStatus string, lastError string) error {
	resp, err := a.emailBackend.MarkEmailAuthStatus(ctx, &mailboxv1.MarkEmailAuthStatusRequest{
		EmailAddress: emailx.Normalize(email),
		AuthStatus:   publicMailboxAuthStatus(authStatus),
		LastError:    strings.TrimSpace(lastError),
	})
	if err != nil {
		return fmt.Errorf("mark mailbox auth status %s: %s", emailx.Redact(email), safeMailboxError(err))
	}
	if resp == nil || resp.GetMailbox() == nil {
		return fmt.Errorf("email service returned empty auth status response for %s", emailx.Redact(email))
	}
	return nil
}

func mailboxCredentialInput(mailbox *mailboxmodel.Record) *mailboxv1.MailboxCredentialInput {
	if mailbox == nil {
		return nil
	}
	return &mailboxv1.MailboxCredentialInput{
		EmailAddress: mailbox.GetEmailAddress(),
		ProviderKey:  mailbox.GetProviderKey(),
		Credentials:  mailboxCredentialValues(mailbox),
		AuthStatus:   publicMailboxAuthStatus(mailbox.GetAuthStatus()),
		LastError:    mailbox.GetLastError(),
	}
}

func mailboxCredentialValues(mailbox *mailboxmodel.Record) []*mailboxv1.MailboxCredentialValue {
	values := []*mailboxv1.MailboxCredentialValue{}
	appendValue := func(kind mailboxv1.MailboxCredentialKind, value string) {
		value = strings.TrimSpace(value)
		if value != "" {
			values = append(values, &mailboxv1.MailboxCredentialValue{Kind: kind, Value: value})
		}
	}
	appendValue(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD, mailbox.GetPassword())
	appendValue(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN, mailbox.GetRefreshToken())
	appendValue(mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN, mailbox.GetAccessToken())
	return values
}
