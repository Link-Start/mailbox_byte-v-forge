package mailboxprovider

import (
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxmodel"
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
	PrepareProjection(*mailboxmodel.Record)
}

type StorageExtension interface {
	Identity
	SchemaStatements() []string
	CanValidatePoll() bool
	ValidatePoll(MailboxRecord) error
	TokenFields() (TokenFields, bool)
	PrepareLegacyData() []string
}

type InboxRetentionPolicy interface {
	Identity
	RetentionPolicy() (MessageRetention, bool)
}

type VirtualMailboxSource interface {
	Identity
	StoredInboxOnly() bool
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
	CapabilitiesFunc      func() *mailboxv1.MailboxProviderCapabilities
	LoadDomainsFunc       func() []string
	DomainsFunc           func([]string) []*mailboxv1.MailboxDomain
	MatchesAddressFunc    func(string, RuntimeContext) bool
	ValidatePollFunc      ValidatePollFunc
	TokenFieldsValue      TokenFields
	RetentionPolicyValue  MessageRetention
	IncludeVirtualFunc    func(string) bool
	PrepareProjectionFunc func(*mailboxmodel.Record)
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

func (p definitionPlugin) CanValidatePoll() bool { return p.definition.ValidatePollFunc != nil }

func (p definitionPlugin) ValidatePoll(row MailboxRecord) error {
	return p.definition.ValidatePollFunc(row)
}

func (p definitionPlugin) TokenFields() (TokenFields, bool) {
	fields := p.definition.TokenFieldsValue
	return fields, fields.HasTokenStorage()
}

func (p definitionPlugin) RetentionPolicy() (MessageRetention, bool) {
	policy := p.definition.RetentionPolicyValue
	return policy, policy.HasRetention()
}

func (p definitionPlugin) IncludeVirtual(authStatus string) bool {
	return p.definition.IncludeVirtualFunc == nil || p.definition.IncludeVirtualFunc(authStatus)
}

func (p definitionPlugin) PrepareProjection(mailbox *mailboxmodel.Record) {
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
