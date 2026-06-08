package main

import (
	"context"

	browserautomationv1 "mailboxapi/internal/contracts/browserautomationv1"
	"mailboxapi/internal/mailboxprovider"

	"mailboxapi/pb"
)

type mailboxRegistrationRunner interface {
	RunMailboxRegistration(context.Context, *pb.RunMailboxRegistrationRequest) (*pb.RunMailboxRegistrationResponse, error)
}

type mailboxOAuthRunner interface {
	RunMailboxOAuth(context.Context, *pb.RunMailboxOAuthRequest) (*pb.RunMailboxOAuthResponse, error)
}

type mailboxProviderActionRegistry struct {
	defaultProvider string
	registration    map[string]mailboxRegistrationRunner
	oauth           map[string]mailboxOAuthRunner
}

type mailboxProviderActionDependencies struct {
	browserClient browserautomationv1.BrowserAutomationServiceClient
}

type mailboxProviderActionPlugin interface {
	RegisterMailboxProviderActions(*mailboxProviderActionRegistry, mailboxProviderActionDependencies)
}

func newMailboxProviderActionRegistry(defaultProvider string) *mailboxProviderActionRegistry {
	return &mailboxProviderActionRegistry{
		defaultProvider: mailboxprovider.NormalizeKey(defaultProvider),
		registration:    map[string]mailboxRegistrationRunner{},
		oauth:           map[string]mailboxOAuthRunner{},
	}
}

func newMailboxProviderActionRegistryForProviders(providers mailboxProviderRuntimeConfig, deps mailboxProviderActionDependencies) *mailboxProviderActionRegistry {
	registry := newMailboxProviderActionRegistry(providers.defaultProvider())
	if providers.registry == nil {
		return registry
	}
	for _, provider := range providers.registry.All() {
		actionPlugin, ok := provider.(mailboxProviderActionPlugin)
		if !ok {
			continue
		}
		actionPlugin.RegisterMailboxProviderActions(registry, deps)
	}
	return registry
}

func (r *mailboxProviderActionRegistry) RegisterRegistration(provider string, runner mailboxRegistrationRunner) {
	if r == nil || runner == nil {
		return
	}
	if key := mailboxprovider.NormalizeKey(provider); key != "" {
		r.registration[key] = runner
	}
}

func (r *mailboxProviderActionRegistry) RegisterOAuth(provider string, runner mailboxOAuthRunner) {
	if r == nil || runner == nil {
		return
	}
	if key := mailboxprovider.NormalizeKey(provider); key != "" {
		r.oauth[key] = runner
	}
}
