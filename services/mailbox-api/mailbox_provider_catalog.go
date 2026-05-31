package main

import (
	"context"
	"strings"
	"sync"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/pb"
)

type mailboxProviderRuntimeConfig struct {
	domainStore  *mailboxProviderDomainStore
	registration outlookRegistrationConfig
}

type mailboxProviderDomainStore struct {
	mu         sync.RWMutex
	byProvider map[string][]string
}

type mailboxProviderPlugin struct {
	key               string
	aliases           []string
	displayName       string
	storedInboxOnly   bool
	schemaStatements  func() []string
	selectJoin        string
	selectFields      mailboxProviderSelectFields
	capabilities      func() *mailboxv1.MailboxProviderCapabilities
	loadDomains       func() []string
	domains           func([]string) []*mailboxv1.MailboxDomain
	matchesAddress    func(string, mailboxProviderRuntimeConfig) bool
	upsert            func(context.Context, pgx.Tx, *pb.EmailMailbox, int64) error
	authFilter        func(string, *[]any) string
	validatePoll      func(*mailboxRow) error
	updateAuth        func(context.Context, pgx.Tx, string, string, string, int64) error
	updateTokens      func(context.Context, *pgxpool.Pool, string, string, string) error
	pruneInbound      func(context.Context, pgx.Tx, mailboxInboxRetention) error
	virtualMailboxes  func(context.Context, *pgxpool.Pool, mailboxListQuery) ([]*pb.EmailMailbox, error)
	includeVirtual    func(string) bool
	prepareProjection func(*pb.EmailMailbox)
	prepareLegacyData func() []string
}

type mailboxProviderSelectFields struct {
	password     string
	refreshToken string
	accessToken  string
	authStatus   string
	lastError    string
}

type mailboxInboxRetention struct {
	touchedMailboxes map[string]struct{}
	touchedDomains   map[string]struct{}
}

func loadMailboxProviderRuntimeConfig() mailboxProviderRuntimeConfig {
	cfg := mailboxProviderRuntimeConfig{
		domainStore:  &mailboxProviderDomainStore{byProvider: map[string][]string{}},
		registration: loadOutlookRegistrationConfig(),
	}
	for _, provider := range mailboxProviderPlugins() {
		if provider.loadDomains != nil {
			cfg.domainStore.set(provider.key, provider.loadDomains())
		}
	}
	return cfg
}

func mailboxProviderPlugins() []*mailboxProviderPlugin {
	return []*mailboxProviderPlugin{
		outlookMailboxProvider(),
		cloudflareMailboxProvider(),
	}
}

func defaultMailboxProvider() string {
	providers := mailboxProviderPlugins()
	if len(providers) == 0 {
		return ""
	}
	return providers[0].key
}

func normalizeMailboxProviderInput(provider string) string {
	value := strings.ToLower(strings.TrimSpace(provider))
	if value == "" {
		return ""
	}
	if definition := providerByKey(value); definition != nil {
		return definition.key
	}
	return value
}

func providerByKey(provider string) *mailboxProviderPlugin {
	value := strings.ToLower(strings.TrimSpace(provider))
	for _, definition := range mailboxProviderPlugins() {
		if definition.key == value {
			return definition
		}
		for _, alias := range definition.aliases {
			if alias == value {
				return definition
			}
		}
	}
	return nil
}
