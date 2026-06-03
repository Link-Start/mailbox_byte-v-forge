package main

import (
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxprovider"
)

func (c mailboxProviderRuntimeConfig) domainsForProvider(provider string) []string {
	return c.DomainsForProvider(provider)
}

func (c mailboxProviderRuntimeConfig) DomainsForProvider(provider string) []string {
	if c.domainStore == nil {
		return nil
	}
	return c.domainStore.get(provider)
}

func (s *mailboxProviderDomainStore) get(provider string) []string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string{}, s.byProvider[mailboxprovider.NormalizeKey(provider)]...)
}

func (s *mailboxProviderDomainStore) set(provider string, domains []string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byProvider[mailboxprovider.NormalizeKey(provider)] = append([]string{}, domains...)
}

func (c mailboxProviderRuntimeConfig) ListDomains(req *mailboxv1.ListMailboxDomainsRequest) *mailboxv1.ListMailboxDomainsResponse {
	providerKey := c.normalizeProviderInput(req.GetProviderKey())
	provider := c.capabilityProviderByKey(providerKey)
	if providerKey != "" {
		if provider == nil {
			return &mailboxv1.ListMailboxDomainsResponse{Domains: []*mailboxv1.MailboxDomain{}}
		}
		return &mailboxv1.ListMailboxDomainsResponse{Domains: provider.Domains(c.DomainsForProvider(provider.Key()))}
	}
	domains := []*mailboxv1.MailboxDomain{}
	for _, provider := range c.capabilityPlugins() {
		domains = append(domains, provider.Domains(c.DomainsForProvider(provider.Key()))...)
	}
	return &mailboxv1.ListMailboxDomainsResponse{Domains: domains}
}

func (c mailboxProviderRuntimeConfig) SyncDomains(req *mailboxv1.SyncMailboxDomainsRequest) *mailboxv1.SyncMailboxDomainsResponse {
	providerKey := c.normalizeProviderInput(req.GetProviderKey())
	provider := c.capabilityProviderByKey(providerKey)
	if providerKey != "" && provider == nil {
		return &mailboxv1.SyncMailboxDomainsResponse{ErrorMessage: "provider cannot sync domains"}
	}
	providers := c.capabilityPlugins()
	if provider != nil {
		providers = []mailboxprovider.CapabilityPlugin{provider}
	}
	for _, candidate := range providers {
		c.domainStore.set(candidate.Key(), candidate.LoadDomains())
	}
	domains := c.ListDomains(&mailboxv1.ListMailboxDomainsRequest{ProviderKey: providerKey}).GetDomains()
	return &mailboxv1.SyncMailboxDomainsResponse{Domains: domains, SyncedCount: int32(len(domains))}
}
