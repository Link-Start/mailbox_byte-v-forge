package main

import (
	"fmt"
	"strings"

	"mailboxapi/internal/mailboxprovider"
)

func outlookMailboxProvider() mailboxprovider.Plugin {
	return mailboxprovider.NewDefinitionPlugin(mailboxprovider.Definition{
		ProviderKey:          emailProviderOutlook,
		AliasKeys:            []string{"microsoft", "graph"},
		DisplayNameValue:     "Outlook",
		SchemaStatementsFunc: outlookSchemaStatements,
		SelectJoinValue:      "LEFT JOIN mailbox_outlook_accounts outlook ON outlook.mailbox_email = m.email",
		SelectFieldsValue:    outlookSelectFields(),
		CapabilitiesFunc:     outlookProviderCapabilities,
		UpsertFunc:           upsertOutlookMailboxData,
		AuthFilterFunc: func(authStatus string, args *[]any) string {
			*args = append(*args, strings.TrimSpace(authStatus))
			return fmt.Sprintf("outlook.auth_status = $%d", len(*args))
		},
		ValidatePollFunc: validateOutlookPollableMailbox,
		UpdateAuthFunc:   updateOutlookAuthStatus,
		TokenFieldsValue: mailboxprovider.TokenFields{
			Table:              "mailbox_outlook_accounts",
			EmailColumn:        "mailbox_email",
			RefreshTokenColumn: "refresh_token",
			AccessTokenColumn:  "access_token",
			AuthStatusColumn:   "auth_status",
			LastErrorColumn:    "last_error",
			UpdatedAtColumn:    "updated_at",
		},
		PrepareLegacyDataFunc: outlookLegacyStatements,
	})
}

func outlookSelectFields() mailboxprovider.SelectFields {
	return mailboxprovider.SelectFields{
		Password:     "CASE WHEN m.provider = 'outlook' THEN outlook.password ELSE '' END",
		RefreshToken: "CASE WHEN m.provider = 'outlook' THEN outlook.refresh_token ELSE '' END",
		AccessToken:  "CASE WHEN m.provider = 'outlook' THEN outlook.access_token ELSE '' END",
		AuthStatus:   "CASE WHEN m.provider = 'outlook' THEN outlook.auth_status ELSE '' END",
		LastError:    "CASE WHEN m.provider = 'outlook' THEN outlook.last_error ELSE '' END",
	}
}
