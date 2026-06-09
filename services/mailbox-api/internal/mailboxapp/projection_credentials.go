package mailboxapp

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/emailx"
	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/redactx"
)

const mailboxErrorSnippetLimit = 600

func RecordFromCredentialInput(input *mailboxv1.MailboxCredentialInput) *mailboxmodel.Record {
	if input == nil {
		return nil
	}
	return applyCredentialInput(&mailboxmodel.Record{
		EmailAddress: emailx.Normalize(input.GetEmailAddress()),
		ProviderKey:  strings.TrimSpace(input.GetProviderKey()),
		AuthStatus:   mailboxmodel.AuthStatusValue(input.GetAuthStatus()),
		LastError:    safeText(input.GetLastError()),
	}, input)
}

func CredentialInput(mailbox *mailboxmodel.Record) *mailboxv1.MailboxCredentialInput {
	if mailbox == nil {
		return nil
	}
	return &mailboxv1.MailboxCredentialInput{
		EmailAddress: mailbox.GetEmailAddress(),
		ProviderKey:  mailbox.GetProviderKey(),
		Credentials:  credentials(mailbox),
		AuthStatus:   mailboxmodel.PublicAuthStatus(mailbox.GetAuthStatus()),
		LastError:    mailbox.GetLastError(),
	}
}

func safeText(value string) string {
	return redactx.TextSnippet(value, mailboxErrorSnippetLimit)
}
