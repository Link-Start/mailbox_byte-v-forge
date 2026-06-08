package main

import (
	"context"

	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"
)

func (s *server) ListMailboxDomains(_ context.Context, req *mailboxv1.ListMailboxDomainsRequest) (*mailboxv1.ListMailboxDomainsResponse, error) {
	return s.providers.ListDomains(req), nil
}

func (s *server) SyncMailboxDomains(_ context.Context, req *mailboxv1.SyncMailboxDomainsRequest) (*mailboxv1.SyncMailboxDomainsResponse, error) {
	return s.providers.SyncDomains(req), nil
}

func (s *server) ListMailboxProviderCapabilities(_ context.Context, req *mailboxv1.ListMailboxProviderCapabilitiesRequest) (*mailboxv1.ListMailboxProviderCapabilitiesResponse, error) {
	return s.providers.ListCapabilities(req), nil
}
