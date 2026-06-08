package main

import (
	"strings"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxprovider"
)

type cloudflareMailboxProviderPlugin struct {
	mailboxprovider.Plugin
}

func cloudflareMailboxProvider(config cloudflareProviderConfig) mailboxprovider.Plugin {
	return cloudflareMailboxProviderPlugin{Plugin: mailboxprovider.NewDefinitionPlugin(mailboxprovider.Definition{
		ProviderKey:          emailProviderCloudflare,
		AliasKeys:            []string{"cf", "cloudflare-email-relay"},
		DisplayNameValue:     "Cloudflare",
		StoredInboxOnlyValue: true,
		CapabilitiesFunc: func() *mailboxv1.MailboxProviderCapabilities {
			return &mailboxv1.MailboxProviderCapabilities{
				Key:         emailProviderCloudflare,
				DisplayName: "Cloudflare",
				Actions: []*mailboxv1.MailboxProviderActionCapability{
					{Action: mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_RECEIVE_WEBHOOK},
					{Action: mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_AUTO_CREATE_MAILBOX},
					{Action: mailboxv1.MailboxProviderAction_MAILBOX_PROVIDER_ACTION_SYNC_DOMAINS},
				},
				RetentionPolicy: &mailboxv1.MailboxMessageRetentionPolicy{
					Scope:       mailboxv1.MailboxMessageRetentionScope_MAILBOX_MESSAGE_RETENTION_SCOPE_DOMAIN,
					MaxMessages: int32(config.maxMessagesPerDomain),
				},
			}
		},
		LoadDomainsFunc: config.loadEmailDomains,
		DomainsFunc: func(configured []string) []*mailboxv1.MailboxDomain {
			domains := make([]*mailboxv1.MailboxDomain, 0, len(configured))
			for _, domain := range configured {
				domains = append(domains, &mailboxv1.MailboxDomain{
					ProviderKey: emailProviderCloudflare,
					Domain:      domain,
					Enabled:     true,
				})
			}
			return domains
		},
		MatchesAddressFunc: func(email string, cfg mailboxprovider.RuntimeContext) bool {
			domain := domainForEmail(email)
			if domain == "" {
				return false
			}
			for _, candidate := range cfg.DomainsForProvider(emailProviderCloudflare) {
				if domain == strings.Trim(strings.ToLower(strings.TrimSpace(candidate)), ".") {
					return true
				}
			}
			return false
		},
		RetentionPolicyValue: mailboxprovider.MessageRetention{
			Scope:       mailboxprovider.RetentionScopeDomain,
			MaxMessages: config.maxMessagesPerDomain,
		},
		IncludeVirtualFunc: func(authStatus string) bool {
			return authStatus == ""
		},
		PrepareProjectionFunc: func(mailbox *mailboxmodel.Record) {
			mailbox.AuthStatus = ""
			mailbox.Password = ""
			mailbox.RefreshToken = ""
			mailbox.AccessToken = ""
			mailbox.LastError = ""
		},
	})}
}

func (p cloudflareMailboxProviderPlugin) RegisterMailboxWebhookRoutes(registry *mailboxWebhookRegistry, deps mailboxWebhookDependencies) {
	registry.Handle("/webhooks/email/cloudflare", deps.handler.handleInboundEmailWebhook(p.Key()))
}
