package main

import (
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxprovider"
)

func outlookMailboxProvider(maxMessages int) mailboxprovider.Plugin {
	return mailboxprovider.NewDefinitionPlugin(mailboxprovider.Definition{
		ProviderKey:          emailProviderOutlook,
		AliasKeys:            []string{"microsoft", "graph"},
		DisplayNameValue:     "Outlook",
		SchemaStatementsFunc: outlookSchemaStatements,
		CapabilitiesFunc: func() *mailboxv1.MailboxProviderCapabilities {
			return outlookProviderCapabilities(maxMessages)
		},
		ValidatePollFunc: validateOutlookPollableMailbox,
		TokenFieldsValue: mailboxprovider.TokenFields{
			Table:              "mailbox_outlook_accounts",
			EmailColumn:        "mailbox_email",
			PasswordColumn:     "password",
			RefreshTokenColumn: "refresh_token",
			AccessTokenColumn:  "access_token",
			AuthStatusColumn:   "auth_status",
			LastErrorColumn:    "last_error",
			CreatedAtColumn:    "created_at",
			UpdatedAtColumn:    "updated_at",
		},
		PrepareLegacyDataFunc: outlookLegacyStatements,
	})
}
