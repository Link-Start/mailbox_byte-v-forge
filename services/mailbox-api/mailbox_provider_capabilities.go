package main

import (
	"github.com/byte-v-forge/common-lib/emailx"
	browserautomationv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/browserautomation/v1"
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/pb"
)

func (c mailboxProviderRuntimeConfig) ListCapabilities(req *mailboxv1.ListMailboxProviderCapabilitiesRequest) *mailboxv1.ListMailboxProviderCapabilitiesResponse {
	providers := []*mailboxv1.MailboxProviderCapabilities{}
	providerKey := normalizeMailboxProviderInput(req.GetProviderKey())
	for _, provider := range mailboxProviderPlugins() {
		if providerKey != "" && providerKey != provider.key {
			continue
		}
		if provider.capabilities != nil {
			providers = append(providers, provider.capabilities())
		}
	}
	return &mailboxv1.ListMailboxProviderCapabilitiesResponse{Providers: providers}
}

func (c mailboxProviderRuntimeConfig) StoredInboxOnlyMailbox(email string) (*pb.EmailMailbox, bool) {
	email = emailx.Normalize(email)
	if email == "" {
		return nil, false
	}
	for _, provider := range mailboxProviderPlugins() {
		if !provider.storedInboxOnly || provider.matchesAddress == nil || !provider.matchesAddress(email, c) {
			continue
		}
		mailbox := &pb.EmailMailbox{
			EmailAddress: email,
			ProviderKey:  provider.key,
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

func newMailboxActivitiesForProviders(cfg mailboxProviderRuntimeConfig, browserClient browserautomationv1.BrowserAutomationServiceClient, emailBackend emailBackend, operations *operationStore, hot *mailboxHotStream) *mailboxActivities {
	return &mailboxActivities{
		outlookRegistration: newOutlookRegistrationRunner(cfg.registration, browserClient, nil),
		emailBackend:        emailBackend,
		operations:          operations,
		hot:                 hot,
	}
}
