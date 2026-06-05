package main

import (
	"context"
	"fmt"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"

	"mailboxapi/internal/mailboxmodel"
	"mailboxapi/internal/mailboxpg"
	"mailboxapi/internal/mailboxprovider"
)

type mailboxInboxSource interface {
	ProviderKey() string
	DefaultMessageLimit() int
	DefaultPollInterval() int
	InboxOverlap() int
	FetchInboxMessages(context.Context, *mailboxmodel.Record, int, int64) ([]*mailboxv1.EmailInboxMessage, error)
}

type mailboxInboxSourceDependencies struct {
	mailboxes *mailboxpg.Repository
}

type mailboxInboxSourcePlugin interface {
	RegisterMailboxInboxSources(*mailboxInboxSourceRegistry, mailboxInboxSourceDependencies)
}

type mailboxInboxSourceRegistry struct {
	defaultProvider string
	byProvider      map[string]mailboxInboxSource
}

func newMailboxInboxSourceRegistry(defaultProvider string) *mailboxInboxSourceRegistry {
	return &mailboxInboxSourceRegistry{
		defaultProvider: mailboxprovider.NormalizeKey(defaultProvider),
		byProvider:      map[string]mailboxInboxSource{},
	}
}

func newMailboxInboxSourceRegistryForProviders(providers mailboxProviderRuntimeConfig, deps mailboxInboxSourceDependencies) *mailboxInboxSourceRegistry {
	registry := newMailboxInboxSourceRegistry(providers.defaultProvider())
	if providers.registry == nil {
		return registry
	}
	for _, provider := range providers.registry.All() {
		sourcePlugin, ok := provider.(mailboxInboxSourcePlugin)
		if !ok {
			continue
		}
		sourcePlugin.RegisterMailboxInboxSources(registry, deps)
	}
	return registry
}

func (r *mailboxInboxSourceRegistry) Register(source mailboxInboxSource) {
	if r == nil || source == nil {
		return
	}
	if key := mailboxprovider.NormalizeKey(source.ProviderKey()); key != "" {
		r.byProvider[key] = source
	}
}

func (r *mailboxInboxSourceRegistry) SourceForMailbox(mailbox *mailboxmodel.Record) (mailboxInboxSource, error) {
	if r == nil {
		return nil, fmt.Errorf("mailbox inbox source registry is required")
	}
	provider := ""
	if mailbox != nil {
		provider = mailbox.GetProviderKey()
	}
	key := mailboxprovider.NormalizeKey(provider)
	if key == "" {
		key = r.defaultProvider
	}
	source := r.byProvider[key]
	if source == nil {
		return nil, fmt.Errorf("mailbox provider cannot fetch inbox: %s", key)
	}
	return source, nil
}

func (r *mailboxInboxSourceRegistry) DefaultMessageLimit() int {
	if r == nil {
		return defaultMessageLimit
	}
	if source := r.byProvider[r.defaultProvider]; source != nil {
		return source.DefaultMessageLimit()
	}
	return defaultMessageLimit
}

func (r *mailboxInboxSourceRegistry) DefaultPollInterval() int {
	if r == nil {
		return defaultPollIntervalSeconds
	}
	if source := r.byProvider[r.defaultProvider]; source != nil {
		return source.DefaultPollInterval()
	}
	return defaultPollIntervalSeconds
}
