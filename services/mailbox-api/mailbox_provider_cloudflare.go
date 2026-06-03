package main

import (
	"context"
	"strings"

	"github.com/byte-v-forge/common-lib/envx"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/jackc/pgx/v5"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxpg"
	"mailboxapi/internal/mailboxprovider"
)

func cloudflareMailboxProvider() mailboxprovider.Plugin {
	return mailboxprovider.NewDefinitionPlugin(mailboxprovider.Definition{
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
					MaxMessages: int32(envx.Int("MAILBOX_CLOUDFLARE_MAX_MESSAGES_PER_DOMAIN", defaultCloudflareMaxDomain)),
				},
			}
		},
		LoadDomainsFunc: loadCloudflareEmailDomains,
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
		PruneInboundFunc: func(ctx context.Context, tx pgx.Tx, retention mailboxprovider.InboxRetention) error {
			for domain := range retention.TouchedDomains {
				if err := mailboxpg.PruneDomainMessages(ctx, tx, emailProviderCloudflare, domain, envx.Int("MAILBOX_CLOUDFLARE_MAX_MESSAGES_PER_DOMAIN", defaultCloudflareMaxDomain)); err != nil {
					return err
				}
			}
			return nil
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
	})
}
