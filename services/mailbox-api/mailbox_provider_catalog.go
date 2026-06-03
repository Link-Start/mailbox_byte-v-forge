package main

import (
	"sync"

	"mailboxapi/internal/mailboxprovider"
)

type mailboxProviderRuntimeConfig struct {
	registry     *mailboxprovider.Registry
	domainStore  *mailboxProviderDomainStore
	registration outlookRegistrationConfig
}

type mailboxProviderDomainStore struct {
	mu         sync.RWMutex
	byProvider map[string][]string
}

var (
	defaultMailboxProvidersOnce  sync.Once
	defaultMailboxProvidersValue *mailboxprovider.Registry
	defaultMailboxProvidersErr   error
)

func defaultMailboxProviderRegistry() *mailboxprovider.Registry {
	defaultMailboxProvidersOnce.Do(func() {
		defaultMailboxProvidersValue, defaultMailboxProvidersErr = mailboxprovider.NewRegistry(
			outlookMailboxProvider(),
			cloudflareMailboxProvider(),
		)
	})
	if defaultMailboxProvidersErr != nil {
		panic(defaultMailboxProvidersErr)
	}
	return defaultMailboxProvidersValue
}

func loadMailboxProviderRuntimeConfig() mailboxProviderRuntimeConfig {
	registry := defaultMailboxProviderRegistry()
	cfg := mailboxProviderRuntimeConfig{
		registry:     registry,
		domainStore:  &mailboxProviderDomainStore{byProvider: map[string][]string{}},
		registration: loadOutlookRegistrationConfig(),
	}
	for _, provider := range cfg.capabilityPlugins() {
		cfg.domainStore.set(provider.Key(), provider.LoadDomains())
	}
	return cfg
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

func defaultMailboxProvider() string {
	return defaultMailboxProviderRegistry().DefaultKey()
}

func normalizeMailboxProviderInput(provider string) string {
	value := mailboxprovider.NormalizeKey(provider)
	if value == "" {
		return ""
	}
	if definition := defaultMailboxProviderRegistry().ByKey(value); definition != nil {
		return definition.Key()
	}
	return value
}

func normalizeMailboxProviderKey(provider string) string {
	return mailboxprovider.NormalizeKey(provider)
}

func providerByKey(provider string) mailboxprovider.Plugin {
	return defaultMailboxProviderRegistry().ByKey(provider)
}

func capabilityProviderByKey(provider string) mailboxprovider.CapabilityPlugin {
	if plugin := providerByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func providerStorageByKey(provider string) mailboxprovider.StorageExtension {
	if plugin := providerByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func providerRetentionByKey(provider string) mailboxprovider.InboxRetentionPolicy {
	if plugin := providerByKey(provider); plugin != nil {
		return plugin
	}
	return nil
}

func mailboxProviderCapabilityPlugins() []mailboxprovider.CapabilityPlugin {
	return defaultMailboxProviderRegistry().CapabilityPlugins()
}

func mailboxProviderStorageExtensions() []mailboxprovider.StorageExtension {
	return defaultMailboxProviderRegistry().StorageExtensions()
}

func mailboxProviderVirtualSources() []mailboxprovider.VirtualMailboxSource {
	return defaultMailboxProviderRegistry().VirtualMailboxSources()
}
