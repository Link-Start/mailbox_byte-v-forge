package mailboxprovider

import (
	mailboxv1 "mailboxapi/internal/contracts/mailboxv1"

	"mailboxapi/internal/mailboxmodel"
)

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
