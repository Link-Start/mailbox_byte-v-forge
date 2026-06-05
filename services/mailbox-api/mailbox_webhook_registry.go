package main

import (
	"net/http"
	"strings"
)

type mailboxWebhookRoute struct {
	path    string
	handler http.HandlerFunc
}

type mailboxWebhookDependencies struct {
	handler *emailWebhookHandler
}

type mailboxWebhookPlugin interface {
	RegisterMailboxWebhookRoutes(*mailboxWebhookRegistry, mailboxWebhookDependencies)
}

type mailboxWebhookRegistry struct {
	routes []mailboxWebhookRoute
}

func newMailboxWebhookRegistryForProviders(providers mailboxProviderRuntimeConfig, deps mailboxWebhookDependencies) *mailboxWebhookRegistry {
	registry := &mailboxWebhookRegistry{}
	if providers.registry == nil {
		return registry
	}
	for _, provider := range providers.registry.All() {
		webhookPlugin, ok := provider.(mailboxWebhookPlugin)
		if !ok {
			continue
		}
		webhookPlugin.RegisterMailboxWebhookRoutes(registry, deps)
	}
	return registry
}

func (r *mailboxWebhookRegistry) Handle(path string, handler http.HandlerFunc) {
	if r == nil || handler == nil {
		return
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	r.routes = append(r.routes, mailboxWebhookRoute{path: path, handler: handler})
}

func (r *mailboxWebhookRegistry) Routes() []mailboxWebhookRoute {
	if r == nil {
		return nil
	}
	return append([]mailboxWebhookRoute{}, r.routes...)
}
