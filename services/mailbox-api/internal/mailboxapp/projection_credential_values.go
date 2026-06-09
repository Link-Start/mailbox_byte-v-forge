package mailboxapp

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
	"mailboxapi/internal/mailboxmodel"
)

func credentials(mailbox *mailboxmodel.Record) []*mailboxv1.MailboxCredentialValue {
	if mailbox == nil {
		return nil
	}
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

func applyCredentialInput(record *mailboxmodel.Record, input *mailboxv1.MailboxCredentialInput) *mailboxmodel.Record {
	if record == nil {
		return nil
	}
	for _, credential := range input.GetCredentials() {
		value := strings.TrimSpace(credential.GetValue())
		switch credential.GetKind() {
		case mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD:
			record.Password = value
		case mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN:
			record.RefreshToken = value
		case mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN:
			record.AccessToken = value
		default:
		}
	}
	return record
}
