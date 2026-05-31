package main

import (
	"github.com/byte-v-forge/common-lib/envx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
)

func outlookProviderCapabilities() *mailboxv1.MailboxProviderCapabilities {
	return &mailboxv1.MailboxProviderCapabilities{
		Key:         emailProviderOutlook,
		DisplayName: "Outlook",
		Actions: []*mailboxv1.MailboxProviderActionCapability{
			{
				Action:        mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_IMPORT_MAILBOX,
				BulkSupported: true,
			},
			{
				Action: mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_RUN_OAUTH,
				RequiredCredentials: []mailboxv1.MailboxCredentialKind{
					mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD,
				},
				RequiredAuthStatuses: []string{authStatusOAuthPending, authStatusAuthFailed, authStatusNeedsManualVerify},
				BulkSupported:        true,
			},
			{
				Action: mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_FETCH_INBOX,
				RequiredCredentials: []mailboxv1.MailboxCredentialKind{
					mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN,
				},
				RequiredAuthStatuses: []string{authStatusAuthorized},
				BulkSupported:        true,
			},
		},
		RetentionPolicy: &mailboxv1.MailboxMessageRetentionPolicy{
			Scope:       mailboxv1.MailboxMessageRetentionScope_MAILBOX_MESSAGE_RETENTION_SCOPE_MAILBOX,
			MaxMessages: int32(envx.Int("MAILBOX_OUTLOOK_MAX_MESSAGES_PER_MAILBOX", defaultOutlookMaxMessages)),
		},
	}
}
