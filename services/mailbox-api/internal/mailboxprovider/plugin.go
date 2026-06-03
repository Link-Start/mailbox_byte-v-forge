package mailboxprovider

import (
	"context"

	mailboxv1 "github.com/byte-v-forge/common-lib/gen/go/byte/v/forge/contracts/mailbox/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mailboxapi/pb"
)

type Identity interface {
	Key() string
	Aliases() []string
	DisplayName() string
}

type CapabilityPlugin interface {
	Identity
	StoredInboxOnly() bool
	Capabilities() *mailboxv1.MailboxProviderCapabilities
	LoadDomains() []string
	Domains([]string) []*mailboxv1.MailboxDomain
	MatchesAddress(string, RuntimeContext) bool
	PrepareProjection(*pb.EmailMailbox)
}

type StorageExtension interface {
	Identity
	SchemaStatements() []string
	SelectJoin() string
	SelectFields() SelectFields
	Upsert(context.Context, pgx.Tx, *pb.EmailMailbox, int64) error
	AuthFilter(string, *[]any) (string, bool)
	CanValidatePoll() bool
	ValidatePoll(MailboxRecord) error
	CanUpdateAuth() bool
	UpdateAuth(context.Context, pgx.Tx, string, string, string, int64) error
	CanUpdateTokens() bool
	UpdateTokens(context.Context, *pgxpool.Pool, string, string, string) error
	PrepareLegacyData() []string
}

type InboxRetentionPolicy interface {
	Identity
	PruneInbound(context.Context, pgx.Tx, InboxRetention) error
}

type VirtualMailboxSource interface {
	Identity
	HasVirtualMailboxes() bool
	VirtualMailboxes(context.Context, *pgxpool.Pool, ListQuery) ([]*pb.EmailMailbox, error)
	IncludeVirtual(string) bool
}

type Plugin interface {
	CapabilityPlugin
	StorageExtension
	InboxRetentionPolicy
	VirtualMailboxSource
}

type Definition struct {
	ProviderKey           string
	AliasKeys             []string
	DisplayNameValue      string
	StoredInboxOnlyValue  bool
	SchemaStatementsFunc  func() []string
	SelectJoinValue       string
	SelectFieldsValue     SelectFields
	CapabilitiesFunc      func() *mailboxv1.MailboxProviderCapabilities
	LoadDomainsFunc       func() []string
	DomainsFunc           func([]string) []*mailboxv1.MailboxDomain
	MatchesAddressFunc    func(string, RuntimeContext) bool
	UpsertFunc            UpsertFunc
	AuthFilterFunc        AuthFilterFunc
	ValidatePollFunc      ValidatePollFunc
	UpdateAuthFunc        UpdateAuthFunc
	UpdateTokensFunc      UpdateTokensFunc
	PruneInboundFunc      PruneInboundFunc
	VirtualMailboxesFunc  VirtualMailboxesFunc
	IncludeVirtualFunc    func(string) bool
	PrepareProjectionFunc func(*pb.EmailMailbox)
	PrepareLegacyDataFunc func() []string
}

type definitionPlugin struct {
	definition Definition
}

func NewDefinitionPlugin(definition Definition) Plugin {
	definition.ProviderKey = NormalizeKey(definition.ProviderKey)
	for index, alias := range definition.AliasKeys {
		definition.AliasKeys[index] = NormalizeKey(alias)
	}
	return definitionPlugin{definition: definition}
}

func (p definitionPlugin) Key() string { return p.definition.ProviderKey }

func (p definitionPlugin) Aliases() []string {
	return append([]string{}, p.definition.AliasKeys...)
}

func (p definitionPlugin) DisplayName() string { return p.definition.DisplayNameValue }

func (p definitionPlugin) StoredInboxOnly() bool { return p.definition.StoredInboxOnlyValue }

func (p definitionPlugin) SchemaStatements() []string {
	if p.definition.SchemaStatementsFunc == nil {
		return nil
	}
	return p.definition.SchemaStatementsFunc()
}

func (p definitionPlugin) SelectJoin() string { return p.definition.SelectJoinValue }

func (p definitionPlugin) SelectFields() SelectFields { return p.definition.SelectFieldsValue }

func (p definitionPlugin) Capabilities() *mailboxv1.MailboxProviderCapabilities {
	if p.definition.CapabilitiesFunc == nil {
		return nil
	}
	return p.definition.CapabilitiesFunc()
}

func (p definitionPlugin) LoadDomains() []string {
	if p.definition.LoadDomainsFunc == nil {
		return nil
	}
	return p.definition.LoadDomainsFunc()
}

func (p definitionPlugin) Domains(configured []string) []*mailboxv1.MailboxDomain {
	if p.definition.DomainsFunc == nil {
		return nil
	}
	return p.definition.DomainsFunc(configured)
}

func (p definitionPlugin) MatchesAddress(email string, cfg RuntimeContext) bool {
	return p.definition.MatchesAddressFunc != nil && p.definition.MatchesAddressFunc(email, cfg)
}

func (p definitionPlugin) Upsert(ctx context.Context, tx pgx.Tx, mailbox *pb.EmailMailbox, now int64) error {
	if p.definition.UpsertFunc == nil {
		return nil
	}
	return p.definition.UpsertFunc(ctx, tx, mailbox, now)
}

func (p definitionPlugin) AuthFilter(authStatus string, args *[]any) (string, bool) {
	if p.definition.AuthFilterFunc == nil {
		return "", false
	}
	return p.definition.AuthFilterFunc(authStatus, args), true
}

func (p definitionPlugin) CanValidatePoll() bool { return p.definition.ValidatePollFunc != nil }

func (p definitionPlugin) ValidatePoll(row MailboxRecord) error {
	return p.definition.ValidatePollFunc(row)
}

func (p definitionPlugin) CanUpdateAuth() bool { return p.definition.UpdateAuthFunc != nil }

func (p definitionPlugin) UpdateAuth(ctx context.Context, tx pgx.Tx, email string, authStatus string, lastError string, now int64) error {
	return p.definition.UpdateAuthFunc(ctx, tx, email, authStatus, lastError, now)
}

func (p definitionPlugin) CanUpdateTokens() bool { return p.definition.UpdateTokensFunc != nil }

func (p definitionPlugin) UpdateTokens(ctx context.Context, pool *pgxpool.Pool, email string, refreshToken string, accessToken string) error {
	return p.definition.UpdateTokensFunc(ctx, pool, email, refreshToken, accessToken)
}

func (p definitionPlugin) PruneInbound(ctx context.Context, tx pgx.Tx, retention InboxRetention) error {
	if p.definition.PruneInboundFunc == nil {
		return nil
	}
	return p.definition.PruneInboundFunc(ctx, tx, retention)
}

func (p definitionPlugin) HasVirtualMailboxes() bool { return p.definition.VirtualMailboxesFunc != nil }

func (p definitionPlugin) VirtualMailboxes(ctx context.Context, pool *pgxpool.Pool, query ListQuery) ([]*pb.EmailMailbox, error) {
	return p.definition.VirtualMailboxesFunc(ctx, pool, query)
}

func (p definitionPlugin) IncludeVirtual(authStatus string) bool {
	return p.definition.IncludeVirtualFunc == nil || p.definition.IncludeVirtualFunc(authStatus)
}

func (p definitionPlugin) PrepareProjection(mailbox *pb.EmailMailbox) {
	if p.definition.PrepareProjectionFunc != nil {
		p.definition.PrepareProjectionFunc(mailbox)
	}
}

func (p definitionPlugin) PrepareLegacyData() []string {
	if p.definition.PrepareLegacyDataFunc == nil {
		return nil
	}
	return p.definition.PrepareLegacyDataFunc()
}
