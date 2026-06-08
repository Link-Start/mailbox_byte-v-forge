package main

import (
	"net/http"
	"sync"
)

type oauthEntry struct {
	refreshToken string
	manager      *OAuthManager
}

type outlookInboxSource struct {
	providerKey   string
	messageLimit  int
	pollInterval  int
	inboxOverlap  int
	httpClient    *http.Client
	oauthConfig   outlookOAuthConfig
	mailboxes     mailboxRepository
	mu            sync.Mutex
	oauthManagers map[string]oauthEntry
}

func newOutlookInboxSource(providerKey string, cfg outlookWatcherConfig, mailboxes mailboxRepository) *outlookInboxSource {
	return &outlookInboxSource{
		providerKey:   providerKey,
		messageLimit:  cfg.messageLimit,
		pollInterval:  cfg.pollInterval,
		inboxOverlap:  cfg.inboxOverlap,
		httpClient:    &http.Client{Timeout: cfg.httpTimeout},
		oauthConfig:   cfg.oauth,
		mailboxes:     mailboxes,
		oauthManagers: map[string]oauthEntry{},
	}
}

func (s *outlookInboxSource) ProviderKey() string {
	return s.providerKey
}

func (s *outlookInboxSource) DefaultMessageLimit() int {
	return s.messageLimit
}

func (s *outlookInboxSource) DefaultPollInterval() int {
	return s.pollInterval
}

func (s *outlookInboxSource) InboxOverlap() int {
	return s.inboxOverlap
}
