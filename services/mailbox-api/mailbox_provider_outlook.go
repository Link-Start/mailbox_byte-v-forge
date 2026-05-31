package main

import (
	"fmt"
	"strings"
)

func outlookMailboxProvider() *mailboxProviderPlugin {
	return &mailboxProviderPlugin{
		key:              emailProviderOutlook,
		aliases:          []string{"microsoft", "graph"},
		displayName:      "Outlook",
		schemaStatements: outlookSchemaStatements,
		selectJoin:       "LEFT JOIN mailbox_outlook_accounts outlook ON outlook.mailbox_email = m.email",
		selectFields:     outlookSelectFields(),
		capabilities:     outlookProviderCapabilities,
		upsert:           upsertOutlookMailboxData,
		authFilter: func(authStatus string, args *[]any) string {
			*args = append(*args, strings.TrimSpace(authStatus))
			return fmt.Sprintf("outlook.auth_status = $%d", len(*args))
		},
		validatePoll:      validateOutlookPollableMailbox,
		updateAuth:        updateOutlookAuthStatus,
		updateTokens:      updateOutlookTokens,
		prepareLegacyData: outlookLegacyStatements,
	}
}

func outlookSelectFields() mailboxProviderSelectFields {
	return mailboxProviderSelectFields{
		password:     "CASE WHEN m.provider = 'outlook' THEN outlook.password ELSE '' END",
		refreshToken: "CASE WHEN m.provider = 'outlook' THEN outlook.refresh_token ELSE '' END",
		accessToken:  "CASE WHEN m.provider = 'outlook' THEN outlook.access_token ELSE '' END",
		authStatus:   "CASE WHEN m.provider = 'outlook' THEN outlook.auth_status ELSE '' END",
		lastError:    "CASE WHEN m.provider = 'outlook' THEN outlook.last_error ELSE '' END",
	}
}
