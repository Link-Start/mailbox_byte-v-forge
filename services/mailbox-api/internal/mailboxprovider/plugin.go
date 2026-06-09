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
