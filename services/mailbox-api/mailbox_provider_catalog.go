package main

import (
	"sync"

	"mailboxapi/internal/mailboxprovider"
)

type mailboxProviderRuntimeConfig struct {
	registry    *mailboxprovider.Registry
	domainStore *mailboxProviderDomainStore
}

type mailboxProviderDomainStore struct {
	mu         sync.RWMutex
	byProvider map[string][]string
}

func loadMailboxProviderRuntimeConfig(config mailboxProviderConfig, outlook outlookRuntimeConfig) (mailboxProviderRuntimeConfig, error) {
	registry, err := mailboxprovider.NewRegistry(
		outlookMailboxProvider(outlookProviderConfig{
			maxMessages:  config.outlookMaxMessages,
			registration: outlook.registration,
			watcher:      outlook.watcher,
		}),
		cloudflareMailboxProvider(config.cloudflare),
	)
	if err != nil {
		return mailboxProviderRuntimeConfig{}, err
	}
	cfg := mailboxProviderRuntimeConfig{
		registry:    registry,
		domainStore: &mailboxProviderDomainStore{byProvider: map[string][]string{}},
	}
	for _, provider := range cfg.capabilityPlugins() {
		cfg.domainStore.set(provider.Key(), provider.LoadDomains())
	}
	return cfg, nil
}

func (c mailboxProviderRuntimeConfig) defaultProvider() string {
	if c.registry == nil {
		return ""
	}
	return c.registry.DefaultKey()
}

func (c mailboxProviderRuntimeConfig) normalizeProviderInput(provider string) string {
	value := mailboxprovider.NormalizeKey(provider)
	if value == "" {
		return ""
	}
	if c.registry != nil {
		if definition := c.registry.ByKey(value); definition != nil {
			return definition.Key()
		}
	}
	return value
}

func (c mailboxProviderRuntimeConfig) capabilityProviderByKey(provider string) mailboxprovider.CapabilityPlugin {
	if c.registry == nil {
		return nil
	}
	if plugin := c.registry.ByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func (c mailboxProviderRuntimeConfig) capabilityPlugins() []mailboxprovider.CapabilityPlugin {
	if c.registry == nil {
		return nil
	}
	return c.registry.CapabilityPlugins()
}
