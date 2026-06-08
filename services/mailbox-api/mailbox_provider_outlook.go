package main

import (
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxprovider"
)

type outlookProviderConfig struct {
	maxMessages  int
	registration outlookRegistrationConfig
	watcher      outlookWatcherConfig
}

type outlookMailboxProviderPlugin struct {
	mailboxprovider.Plugin
	registration outlookRegistrationConfig
	watcher      outlookWatcherConfig
}

func outlookMailboxProvider(config outlookProviderConfig) mailboxprovider.Plugin {
	return outlookMailboxProviderPlugin{
		Plugin: mailboxprovider.NewDefinitionPlugin(mailboxprovider.Definition{
			ProviderKey:          emailProviderOutlook,
			AliasKeys:            []string{"microsoft", "graph"},
			DisplayNameValue:     "Outlook",
			SchemaStatementsFunc: outlookSchemaStatements,
			CapabilitiesFunc: func() *mailboxv1.MailboxProviderCapabilities {
				return outlookProviderCapabilities(config.maxMessages)
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
		}),
		registration: config.registration,
		watcher:      config.watcher,
	}
}

func (p outlookMailboxProviderPlugin) RegisterMailboxProviderActions(registry *mailboxProviderActionRegistry, deps mailboxProviderActionDependencies) {
	runner := newOutlookRegistrationRunner(p.registration, deps.browserClient, nil)
	registry.RegisterRegistration(p.Key(), runner)
	registry.RegisterOAuth(p.Key(), runner)
}

func (p outlookMailboxProviderPlugin) RegisterMailboxInboxSources(registry *mailboxInboxSourceRegistry, deps mailboxInboxSourceDependencies) {
	registry.Register(newOutlookInboxSource(p.Key(), p.watcher, deps.mailboxes))
}

func (p outlookMailboxProviderPlugin) RegisterMailboxWebhookRoutes(registry *mailboxWebhookRegistry, deps mailboxWebhookDependencies) {
	registry.Handle("/webhooks/email/microsoft-graph", deps.handler.handleGraphNotification)
}
