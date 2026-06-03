package main

import (
	"github.com/byte-v-forge/common-lib/emailx"
	browserautomationv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/browserautomation/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxpg"
)

func (c mailboxProviderRuntimeConfig) ListCapabilities(req *mailboxv1.ListMailboxProviderCapabilitiesRequest) *mailboxv1.ListMailboxProviderCapabilitiesResponse {
	providers := []*mailboxv1.MailboxProviderCapabilities{}
	providerKey := c.normalizeProviderInput(req.GetProviderKey())
	for _, provider := range c.capabilityPlugins() {
		if providerKey != "" && providerKey != provider.Key() {
			continue
		}
		if capabilities := provider.Capabilities(); capabilities != nil {
			providers = append(providers, capabilities)
		}
	}
	return &mailboxv1.ListMailboxProviderCapabilitiesResponse{Providers: providers}
}

func (c mailboxProviderRuntimeConfig) StoredInboxOnlyMailbox(email string) (*mailboxmodel.Record, bool) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, false
	}
	for _, provider := range c.capabilityPlugins() {
		if !provider.StoredInboxOnly() || !provider.MatchesAddress(email, c) {
			continue
		}
		mailbox := &mailboxmodel.Record{
			EmailAddress: email,
			ProviderKey:  provider.Key(),
			Domain:       domainForEmail(email),
		}
		c.prepareProjection(mailbox)
		return mailbox, true
	}
	return nil, false
}

func (c mailboxProviderRuntimeConfig) ProviderForInboxAddress(email string, messages []*mailboxv1.EmailInboxMessage) string {
	for _, message := range messages {
		if provider := c.normalizeProviderInput(message.GetProviderKey()); provider != "" {
			return provider
		}
	}
	if mailbox, ok := c.StoredInboxOnlyMailbox(email); ok {
		return mailbox.GetProviderKey()
	}
	return c.defaultProvider()
}

func (c mailboxProviderRuntimeConfig) IsStoredInboxOnlyAddress(email string) bool {
	_, ok := c.StoredInboxOnlyMailbox(email)
	return ok
}

func newMailboxActivitiesForProviders(cfg mailboxProviderRuntimeConfig, registrationCfg outlookRegistrationConfig, browserClient browserautomationv1.BrowserAutomationServiceClient, emailBackend emailBackend, mailboxRepo *mailboxpg.Repository, operations *operationStore, hot *mailboxHotStream) *mailboxActivities {
	providerActions := newMailboxProviderActionRegistry(cfg.defaultProvider())
	outlookRegistration := newOutlookRegistrationRunner(registrationCfg, browserClient, nil)
	providerActions.RegisterRegistration(emailProviderOutlook, outlookRegistration)
	providerActions.RegisterOAuth(emailProviderOutlook, outlookRegistration)
	return &mailboxActivities{
		providerActions: providerActions,
		emailBackend:    emailBackend,
		mailboxRepo:     mailboxRepo,
		operations:      operations,
		hot:             hot,
	}
}
