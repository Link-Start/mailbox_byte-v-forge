package main

import mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

func outlookProviderCapabilities(maxMessages int) *mailboxv1.MailboxProviderCapabilities {
	return &mailboxv1.MailboxProviderCapabilities{
		Key:         emailProviderOutlook,
		DisplayName: "Outlook",
		Actions: []*mailboxv1.MailboxProviderActionCapability{
			{
				Action:        mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_IMPORT_MAILBOX,
				BulkSupported: true,
				RequiredCredentials: []mailboxv1.MailboxCredentialKind{
					mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD,
					mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN,
					mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_ACCESS_TOKEN,
				},
			},
			{
				Action: mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_RUN_OAUTH,
				RequiredCredentials: []mailboxv1.MailboxCredentialKind{
					mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_PASSWORD,
				},
				RequiredAuthStatuses: []mailboxv1.MailboxAuthStatus{
					mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_OAUTH_PENDING,
					mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_AUTH_FAILED,
					mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_NEEDS_MANUAL_VERIFICATION,
				},
				BulkSupported: true,
			},
			{
				Action: mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_FETCH_INBOX,
				RequiredCredentials: []mailboxv1.MailboxCredentialKind{
					mailboxv1.MailboxCredentialKind_MAILBOX_CREDENTIAL_KIND_OAUTH_REFRESH_TOKEN,
				},
				RequiredAuthStatuses: []mailboxv1.MailboxAuthStatus{
					mailboxv1.MailboxAuthStatus_MAILBOX_AUTH_STATUS_AUTHORIZED,
				},
				BulkSupported: true,
			},
		},
		RetentionPolicy: &mailboxv1.MailboxMessageRetentionPolicy{
			Scope:       mailboxv1.MailboxMessageRetentionScope_MAILBOX_MESSAGE_RETENTION_SCOPE_MAILBOX,
			MaxMessages: int32(maxMessages),
		},
	}
}
