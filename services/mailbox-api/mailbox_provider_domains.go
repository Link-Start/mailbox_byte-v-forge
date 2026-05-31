package main

import (
	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
)

func (c mailboxProviderRuntimeConfig) domainsForProvider(provider string) []string {
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
	return append([]string{}, s.byProvider[normalizeMailboxProviderInput(provider)]...)
}

func (s *mailboxProviderDomainStore) set(provider string, domains []string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byProvider[normalizeMailboxProviderInput(provider)] = append([]string{}, domains...)
}

func (c mailboxProviderRuntimeConfig) ListDomains(req *mailboxv1.ListMailboxDomainsRequest) *mailboxv1.ListMailboxDomainsResponse {
	providerKey := normalizeMailboxProviderInput(req.GetProviderKey())
	provider := providerByKey(providerKey)
	if providerKey != "" {
		if provider == nil || provider.domains == nil {
			return &mailboxv1.ListMailboxDomainsResponse{Domains: []*mailboxv1.MailboxDomain{}}
		}
		return &mailboxv1.ListMailboxDomainsResponse{Domains: provider.domains(c.domainsForProvider(provider.key))}
	}
	domains := []*mailboxv1.MailboxDomain{}
	for _, provider := range mailboxProviderPlugins() {
		if provider.domains != nil {
			domains = append(domains, provider.domains(c.domainsForProvider(provider.key))...)
		}
	}
	return &mailboxv1.ListMailboxDomainsResponse{Domains: domains}
}

func (c mailboxProviderRuntimeConfig) SyncDomains(req *mailboxv1.SyncMailboxDomainsRequest) *mailboxv1.SyncMailboxDomainsResponse {
	providerKey := normalizeMailboxProviderInput(req.GetProviderKey())
	provider := providerByKey(providerKey)
	if providerKey != "" && provider == nil {
		return &mailboxv1.SyncMailboxDomainsResponse{ErrorMessage: "provider cannot sync domains"}
	}
	providers := mailboxProviderPlugins()
	if provider != nil {
		providers = []*mailboxProviderPlugin{provider}
	}
	for _, candidate := range providers {
		if candidate.loadDomains == nil {
			continue
		}
		c.domainStore.set(candidate.key, candidate.loadDomains())
	}
	domains := c.ListDomains(&mailboxv1.ListMailboxDomainsRequest{ProviderKey: providerKey}).GetDomains()
	return &mailboxv1.SyncMailboxDomainsResponse{Domains: domains, SyncedCount: int32(len(domains))}
}
