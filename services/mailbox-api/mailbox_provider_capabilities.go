package main

import (
	"github.com/byte-v-forge/common-lib/emailx"
	browserautomationv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/browserautomation/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxmodel"
)

func (c mailboxProviderRuntimeConfig) ListCapabilities(req *mailboxv1.ListMailboxProviderCapabilitiesRequest) *mailboxv1.ListMailboxProviderCapabilitiesResponse {
	providers := []*mailboxv1.MailboxProviderCapabilities{}
	providerKey := normalizeMailboxProviderInput(req.GetProviderKey())
	for _, provider := range mailboxProviderCapabilityPlugins() {
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
	for _, provider := range mailboxProviderCapabilityPlugins() {
		if !provider.StoredInboxOnly() || !provider.MatchesAddress(email, c) {
			continue
		}
		mailbox := &mailboxmodel.Record{
			EmailAddress: email,
			ProviderKey:  provider.Key(),
			Domain:       domainForEmail(email),
		}
		prepareMailboxProjection(mailbox)
		return mailbox, true
	}
	return nil, false
}

func (c mailboxProviderRuntimeConfig) ProviderForInboxAddress(email string, messages []*mailboxv1.EmailInboxMessage) string {
	for _, message := range messages {
		if provider := normalizeEmailProvider(message.GetProviderKey()); provider != "" {
			return provider
		}
	}
	if mailbox, ok := c.StoredInboxOnlyMailbox(email); ok {
		return mailbox.GetProviderKey()
	}
	return defaultMailboxProvider()
}

func (c mailboxProviderRuntimeConfig) IsStoredInboxOnlyAddress(email string) bool {
	_, ok := c.StoredInboxOnlyMailbox(email)
	return ok
}

func newMailboxActivitiesForProviders(cfg mailboxProviderRuntimeConfig, browserClient browserautomationv1.BrowserAutomationServiceClient, emailBackend emailBackend, mailboxStore *MailboxStore, operations *operationStore, hot *mailboxHotStream) *mailboxActivities {
	return &mailboxActivities{
		outlookRegistration: newOutlookRegistrationRunner(cfg.registration, browserClient, nil),
		emailBackend:        emailBackend,
		mailboxStore:        mailboxStore,
		operations:          operations,
		hot:                 hot,
	}
}
